package drivers

import (
	"context"
	"sync"
	"time"
)

// VPIOFactory provides in-memory GPIO lines for tests.
type VPIOFactory struct {
	mu    sync.Mutex
	lines map[string]*vpioLine
}

// NewVPIOFactory constructs an in-memory GPIO factory for tests.
func NewVPIOFactory() *VPIOFactory {
	return &VPIOFactory{lines: map[string]*vpioLine{}}
}

func (f *VPIOFactory) key(chip string, offset int) string { return chip + ":" + itoa(offset) }

// OpenInput returns a virtual input line backed by memory.
func (f *VPIOFactory) OpenInput(chip string, offset int, edge Edge, bias Bias, debounce time.Duration) (InputLine, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	k := f.key(chip, offset)
	l := f.lines[k]
	if l == nil {
		l = &vpioLine{
			edge:  edge,
			value: false,
			evtQ:  make(chan LineEvent, 64),
		}
		f.lines[k] = l
	}
	l.edge = edge
	return l, nil
}

// OpenOutput returns a virtual output line backed by memory.
func (f *VPIOFactory) OpenOutput(chip string, offset int, initial bool) (OutputLine, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	k := f.key(chip, offset)
	l := f.lines[k]
	if l == nil {
		l = &vpioLine{
			edge:  EdgeNone,
			value: initial,
			evtQ:  make(chan LineEvent, 64),
		}
		f.lines[k] = l
	}
	l.value = initial
	return l, nil
}

// InjectEdge lets tests simulate button edges.
func (f *VPIOFactory) InjectEdge(chip string, offset int, edge Edge, value bool) {
	f.mu.Lock()
	l := f.lines[f.key(chip, offset)]
	f.mu.Unlock()
	if l == nil {
		return
	}
	l.mu.Lock()
	l.value = value
	l.mu.Unlock()

	select {
	case l.evtQ <- LineEvent{Time: time.Now(), Edge: edge, Value: value}:
	default:
	}
}

type vpioLine struct {
	mu    sync.Mutex
	edge  Edge
	value bool
	evtQ  chan LineEvent
}

func (l *vpioLine) Read() (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.value, nil
}

func (l *vpioLine) Write(v bool) error {
	l.mu.Lock()
	l.value = v
	l.mu.Unlock()
	return nil
}

func (l *vpioLine) Events(ctx context.Context) (<-chan LineEvent, error) {
	out := make(chan LineEvent, 64)
	go func() {
		defer close(out)
		for {
			select {
			case ev := <-l.evtQ:
				// filter by configured edge
				if l.edge == EdgeRising && ev.Edge != EdgeRising {
					continue
				}
				if l.edge == EdgeFalling && ev.Edge != EdgeFalling {
					continue
				}
				if l.edge == EdgeNone {
					continue
				}
				out <- ev
			case <-ctx.Done():
				return
			}
		}
	}()
	return out, nil
}

func (l *vpioLine) Close() error { return nil }

// tiny helper
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	sign := ""
	if n < 0 {
		sign = "-"
		n = -n
	}
	var buf [12]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + (n % 10))
		n /= 10
	}
	return sign + string(buf[i:])
}
