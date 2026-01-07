package devices

import "time"

type EventKind string

const (
	EventOpen  EventKind = "open"
	EventClose EventKind = "close"
	EventError EventKind = "error"
	EventEdge  EventKind = "edge"
	EventInfo  EventKind = "info"
)

type Event struct {
	Device string
	Kind   EventKind
	Time   time.Time

	// Optional fields
	Msg  string
	Err  error
	Meta map[string]string
}
