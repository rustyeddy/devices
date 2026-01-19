package devices

import (
	"context"
	"sync"
	"time"
)

// Base provides device name and event publishing helpers.
// Base owns its event channel and manages its lifecycle.
type Base struct {
	name   string
	events chan Event
	ctx    context.Context
	cancel context.CancelFunc
	once   sync.Once
}

// NewBase constructs a Base with a buffered event channel.
// The returned Base must have Close() called when done to properly clean up resources.
func NewBase(name string, eventBuf int) Base {
	if eventBuf <= 0 {
		eventBuf = 16
	}
	ctx, cancel := context.WithCancel(context.Background())
	return Base{
		name:   name,
		events: make(chan Event, eventBuf),
		ctx:    ctx,
		cancel: cancel,
	}
}

// Name returns the device name.
func (b *Base) Name() string {
	return b.name
}

// Events returns the event stream channel.
func (b *Base) Events() <-chan Event {
	return b.events
}

// Emit publishes an event without blocking.
// Events are silently dropped if the channel buffer is full or if Base is closed.
func (b *Base) Emit(kind EventKind, msg string, err error, meta map[string]string) {
	select {
	case <-b.ctx.Done():
		return
	default:
	}

	e := Event{
		Device: b.name,
		Kind:   kind,
		Time:   time.Now(),
		Msg:    msg,
		Err:    err,
		Meta:   meta,
	}

	select {
	case b.events <- e:
	default:
		// drop if slow consumer. TODO log dropped events, but must be rate limited.
	}
}

// EmitBlocking publishes an event and blocks until delivered.
// Returns immediately without sending if Base is closed.
// This method can block indefinitely if there is no consumer.
func (b *Base) EmitBlocking(kind EventKind, msg string, err error, meta map[string]string) {
	select {
	case <-b.ctx.Done():
		return
	default:
	}

	e := Event{
		Device: b.name,
		Kind:   kind,
		Time:   time.Now(),
		Msg:    msg,
		Err:    err,
		Meta:   meta,
	}

	select {
	case <-b.ctx.Done():
		return
	case b.events <- e:
	}
}

// EmitWithContext publishes an event with cancellation support.
// Returns context.Canceled if ctx is canceled, context.DeadlineExceeded if ctx times out,
// or nil if the event was successfully delivered.
// Returns immediately without error if Base is closed.
func (b *Base) EmitWithContext(ctx context.Context, kind EventKind, msg string, err error, meta map[string]string) error {
	select {
	case <-b.ctx.Done():
		return nil
	default:
	}

	e := Event{
		Device: b.name,
		Kind:   kind,
		Time:   time.Now(),
		Msg:    msg,
		Err:    err,
		Meta:   meta,
	}

	select {
	case <-b.ctx.Done():
		return nil
	case b.events <- e:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Close shuts down the Base and closes its event channel.
// As the channel owner, Base is responsible for closing the channel exactly once.
// After Close is called, all Emit methods will return immediately without sending.
// Close is safe to call multiple times.
func (b *Base) Close() error {
	b.once.Do(func() {
		b.cancel() // Signal shutdown
		close(b.events)
	})
	return nil
}
