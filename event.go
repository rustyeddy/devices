package devices

import "time"

// EventKind describes the type of device event.
// EventKind is a string to allow custom event types and easy debugging.
type EventKind string

const (
	// EventOpen signals the device started running.
	// Typically emitted when Run() begins.
	EventOpen EventKind = "open"

	// EventClose signals the device stopped running.
	// Typically emitted before Run() returns.
	EventClose EventKind = "close"

	// EventError reports a device error.
	// The Event.Err field should contain the actual error.
	EventError EventKind = "error"

	// EventEdge reports a change in device state.
	// Common for buttons, switches, and GPIO pins.
	// Metadata often includes state details (e.g., "state": "high").
	EventEdge EventKind = "edge"

	// EventInfo reports an informational message.
	// Used for logging and diagnostics.
	EventInfo EventKind = "info"

	// EventRead signals a successful read operation.
	// Common for sensors and input devices.
	EventRead EventKind = "read"

	// EventWrite signals a successful write operation.
	// Common for actuators and output devices.
	EventWrite EventKind = "write"
)

// Event captures a device notification.
//
// Events are emitted through a device's Events() channel to notify consumers
// of lifecycle changes, errors, state transitions, and informational messages.
//
// Required fields:
//   - Device: name of the device emitting the event
//   - Kind: type of event (use predefined EventKind constants)
//   - Time: when the event occurred
//
// Optional fields:
//   - Msg: human-readable description
//   - Err: the actual error (when Kind is EventError)
//   - Meta: key-value pairs for additional context
//
// Example usage:
//
//	ev := Event{
//	    Device: "temp-sensor",
//	    Kind:   EventRead,
//	    Time:   time.Now(),
//	    Meta: map[string]string{
//	        "value": "23.5",
//	        "unit":  "C",
//	    },
//	}
//
// Zero value: All fields are zero/nil. The zero value is valid but not useful.
type Event struct {
	Device string
	Kind   EventKind
	Time   time.Time

	// Optional fields
	Msg  string            // Human-readable message
	Err  error             // Actual error (use with EventError)
	Meta map[string]string // Additional context (nil is valid)
}
