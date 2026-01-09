package devices

import (
	"sync"
	"time"
)

// Base struct for the devices package with internal structures.
type Base struct {
	name   string
	events chan Event
	once   sync.Once
}

func NewBase(name string, eventBuf int) Base {
	if eventBuf <= 0 {
		eventBuf = 16
	}
	return Base{
		name:   name,
		events: make(chan Event, eventBuf),
	}
}

func (b *Base) Name() string {
	return b.name
}

func (b *Base) Events() <-chan Event {
	return b.events
}

// Emit tries to publish an event without deadlocking.  For noisy
// devices, non-blocking is a good default.  For critical events, you
// can call EmitBlocking (below).
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

// EmitBlocking also will emit an event however it will block for
// critical events
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

// CloseEvents safely closes the events channel
func (b *Base) CloseEvents() {
	b.once.Do(func() { close(b.events) })
}
