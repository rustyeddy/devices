package devices

import "context"

// Device is the minimal lifecycle contract.
// A device owns its goroutines and channels.
type Device interface {
	Name() string

	// Run blocks until ctx is canceled or a fatal error occurs.
	// The device must close all its output channels before returning.
	Run(ctx context.Context) error

	// Events emits lifecycle, error, and edge notifications.
	Events() <-chan Event
}

// Source emits typed values (sensors, buttons, GPS fixes, etc.)
type Source[T any] interface {
	Device
	Out() <-chan T
}

// Sink consumes typed commands (relay, motor, LED, PWM, etc.)
type Sink[T any] interface {
	Device
	In() chan<- T
}

// Duplex does both (e.g. dimmer, motor controller).
type Duplex[T any] interface {
	Source[T]
	Sink[T]
}
