//go:build !linux

package button

import (
	"log/slog"
	"time"

	"github.com/rustyeddy/devices"
	"github.com/rustyeddy/devices/drivers"
)

// Button is a simple momentary GPIO Device that can be detected with
// the Button is pushed (rising edge) or when it is released (falling
// edge). Low is open, high is closed.
type Button struct {
	*drivers.DigitalPin
	devices.Device[int]
}

// New creates a new button with the given name, offset represents the
// pin number and a series of line options. This is a stub implementation
// for non-Linux platforms.
func New(name string, offset int, opts ...drivers.LineReqOption) *Button {
	// On non-Linux platforms, we create a mock button
	slog.Info("Creating stub button (non-Linux platform)", "name", name, "offset", offset)
	
	b := &Button{
		DigitalPin: drivers.NewDigitalPin(name, offset, opts...),
	}
	b.Device = b

	return b
}

func (b *Button) ID() string {
	return b.PinName()
}

func (b *Button) Open() error {
	return nil
}

func (b *Button) Close() error {
	return nil
}

// ReadPub will read the value of the button and publish the results.
func (b *Button) Get() (int, error) {
	val, err := b.DigitalPin.Get()
	if err != nil {
		slog.Error("Failed to read buttons value: ", "error", err.Error())
		return val, err
	}
	return val, err
}

func (b *Button) Set(but int) error {
	return devices.ErrNotImplemented
}

func (b *Button) Type() devices.Type {
	return devices.TypeInt
}

// EventLoop is a stub for non-Linux platforms
func (b *Button) EventLoop(done chan any, readpub func()) {
	slog.Info("Button EventLoop (stub implementation)")
	running := true
	for running {
		select {
		case <-done:
			running = false
		case <-time.After(time.Second):
			// Periodic check for simulated events
		}
	}
}
