package devices

import (
	"sync"
	"time"
)

// Base provides device name and event publishing helpers.
type Base struct {
	name   string
	events chan Event
	once   sync.Once
}

// NewBase constructs a Base with a buffered event channel.
func NewBase(name string, eventBuf int) Base {
	if eventBuf <= 0 {
		eventBuf = 16
	}
	return Base{
		name:   name,
		events: make(chan Event, eventBuf),
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
func (b *Base) Emit(kind EventKind, msg string, err error, meta map[string]string) {
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
		// drop if slow consumer
	}
}

// EmitBlocking publishes an event and blocks until delivered.
func (b *Base) EmitBlocking(kind EventKind, msg string, err error, meta map[string]string) {
	b.events <- Event{
		Device: b.name,
		Kind:   kind,
		Time:   time.Now(),
		Msg:    msg,
		Err:    err,
		Meta:   meta,
	}
}

// CloseEvents closes the events channel once.
func (b *Base) CloseEvents() {
	b.once.Do(func() { close(b.events) })
}
