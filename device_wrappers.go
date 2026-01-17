package devices

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

//
// RateLimit
//

// RateLimit emits at most one value per period.
// Implementation: keeps the most recent value seen and emits it on each tick.
func RateLimit[T any](src Sensor[T], period time.Duration) Sensor[T] {
	return &rateLimit[T]{
		src:    src,
		period: period,
		out:    make(chan T, 1),
	}
}

type rateLimit[T any] struct {
	src     Sensor[T]
	period  time.Duration
	out     chan T
	closeMu sync.Once
	cancel  context.CancelFunc
}

func (r *rateLimit[T]) Read() <-chan T { return r.out }

func (r *rateLimit[T]) Run(ctx context.Context) error {
	ctx, r.cancel = context.WithCancel(ctx)

	// Start source first.
	if err := r.src.Run(ctx); err != nil {
		return err
	}

	in := r.src.Read()

	go func() {
		defer close(r.out)

		t := time.NewTicker(r.period)
		defer t.Stop()

		var (
			last T
			have bool
		)

		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-in:
				if !ok {
					return
				}
				last = v
				have = true
			case <-t.C:
				if have {
					// Non-blocking-ish: if downstream is slow, keep latest in buffer.
					select {
					case r.out <- last:
					default:
						// Drop tick if consumer is behind; next tick will send latest.
					}
				}
			}
		}
	}()

	return nil
}

func (r *rateLimit[T]) Close() error {
	r.closeMu.Do(func() {
		if r.cancel != nil {
			r.cancel()
		}
	})
	return r.src.Close()
}

//
// Debounce
//

// DebounceComparable emits a value only after it has remained unchanged for d.
// Requires T comparable.
func DebounceComparable[T comparable](src Sensor[T], d time.Duration) Sensor[T] {
	return DebounceFunc(src, d, func(a, b T) bool { return a == b })
}

// DebounceFunc is Debounce with a custom equality comparator.
func DebounceFunc[T any](src Sensor[T], d time.Duration, equal func(a, b T) bool) Sensor[T] {
	return &debounce[T]{
		src:   src,
		d:     d,
		equal: equal,
		out:   make(chan T, 1),
	}
}

type debounce[T any] struct {
	src   Sensor[T]
	d     time.Duration
	equal func(a, b T) bool
	out   chan T

	closeMu sync.Once
	cancel  context.CancelFunc
}

func (d *debounce[T]) Read() <-chan T { return d.out }

func (d *debounce[T]) Run(ctx context.Context) error {
	ctx, d.cancel = context.WithCancel(ctx)

	if err := d.src.Run(ctx); err != nil {
		return err
	}

	in := d.src.Read()

	go func() {
		defer close(d.out)

		var (
			pending     T
			havePending bool
			timer       *time.Timer
			timerC      <-chan time.Time
		)

		stopTimer := func() {
			if timer != nil {
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				timer = nil
				timerC = nil
			}
		}

		resetTimer := func() {
			stopTimer()
			timer = time.NewTimer(d.d)
			timerC = timer.C
		}

		for {
			select {
			case <-ctx.Done():
				stopTimer()
				return

			case v, ok := <-in:
				if !ok {
					stopTimer()
					return
				}

				if !havePending || !d.equal(v, pending) {
					pending = v
					havePending = true
					resetTimer()
				} else {
					// Same value; push debounce window out.
					resetTimer()
				}

			case <-timerC:
				// Window elapsed; emit pending
				if havePending {
					select {
					case d.out <- pending:
					case <-ctx.Done():
						return
					}
				}
				// After emitting, wait for next change.
				havePending = false
				stopTimer()
			}
		}
	}()

	return nil
}

func (d *debounce[T]) Close() error {
	d.closeMu.Do(func() {
		if d.cancel != nil {
			d.cancel()
		}
	})
	return d.src.Close()
}

//
// LastValue
//

// LastValued is a Sensor with an additional Last() getter.
type LastValued[T any] interface {
	Sensor[T]
	Last() (T, bool)
}

// LastValue wraps a sensor and remembers the last value observed.
func LastValue[T any](src Sensor[T]) LastValued[T] {
	return &lastValue[T]{
		src: src,
		out: make(chan T, 1),
	}
}

type lastValue[T any] struct {
	src Sensor[T]
	out chan T

	// store last value + have flag
	mu   sync.RWMutex
	last T
	have bool

	closeMu sync.Once
	cancel  context.CancelFunc
}

func (l *lastValue[T]) Read() <-chan T { return l.out }

func (l *lastValue[T]) Last() (T, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.last, l.have
}

func (l *lastValue[T]) Run(ctx context.Context) error {
	ctx, l.cancel = context.WithCancel(ctx)

	if err := l.src.Run(ctx); err != nil {
		return err
	}

	in := l.src.Read()

	go func() {
		defer close(l.out)
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-in:
				if !ok {
					return
				}
				l.mu.Lock()
				l.last = v
				l.have = true
				l.mu.Unlock()

				select {
				case l.out <- v:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return nil
}

func (l *lastValue[T]) Close() error {
	l.closeMu.Do(func() {
		if l.cancel != nil {
			l.cancel()
		}
	})
	return l.src.Close()
}

//
// FanOut
//

// FanOutHub lets multiple consumers subscribe to the same sensor stream.
type FanOutHub[T any] interface {
	Run(ctx context.Context) error
	Close() error
	Subscribe() <-chan T
	Subscribers() int
}

// FanOut creates a hub that broadcasts each reading to all subscribers.
// n = initial subscribers created; buf = per-subscriber buffer.
func FanOut[T any](src Sensor[T], n int, buf int) FanOutHub[T] {
	h := &fanOut[T]{
		src: src,
		buf: buf,
	}
	for i := 0; i < n; i++ {
		h.Subscribe()
	}
	return h
}

type fanOut[T any] struct {
	src Sensor[T]
	buf int

	mu     sync.RWMutex
	subs   []chan T
	closed atomic.Bool

	closeMu sync.Once
	cancel  context.CancelFunc
}

func (f *fanOut[T]) Subscribers() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return len(f.subs)
}

func (f *fanOut[T]) Subscribe() <-chan T {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.closed.Load() {
		ch := make(chan T)
		close(ch)
		return ch
	}

	ch := make(chan T, f.buf)
	f.subs = append(f.subs, ch)
	return ch
}

func (f *fanOut[T]) Run(ctx context.Context) error {
	ctx, f.cancel = context.WithCancel(ctx)

	if err := f.src.Run(ctx); err != nil {
		return err
	}

	in := f.src.Read()

	go func() {
		defer f.closeAll()

		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-in:
				if !ok {
					return
				}

				// Snapshot subscribers under RLock; avoid holding lock during sends.
				f.mu.RLock()
				snapshot := append([]chan T(nil), f.subs...)
				f.mu.RUnlock()

				for _, ch := range snapshot {
					// Don’t let one slow subscriber block everyone.
					select {
					case ch <- v:
					default:
						// drop for this subscriber
					}
				}
			}
		}
	}()

	return nil
}

func (f *fanOut[T]) Close() error {
	f.closeMu.Do(func() {
		if f.cancel != nil {
			f.cancel()
		}
		f.closed.Store(true)
	})
	return f.src.Close()
}

func (f *fanOut[T]) closeAll() {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, ch := range f.subs {
		close(ch)
	}
	f.subs = nil
}
