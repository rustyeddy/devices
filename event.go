package devices

import "time"

// EventKind describes the type of device event.
type EventKind string

const (
	// EventOpen signals the device started running.
	EventOpen EventKind = "open"
	// EventClose signals the device stopped running.
	EventClose EventKind = "close"
	// EventError reports a device error.
	EventError EventKind = "error"
	// EventEdge reports a change in device state.
	EventEdge EventKind = "edge"
	// EventInfo reports an informational message.
	EventInfo EventKind = "info"
)

// Event captures a device notification.
type Event struct {
	Device string
	Kind   EventKind
	Time   time.Time

	// Optional fields
	Msg  string
	Err  error
	Meta map[string]string
}
