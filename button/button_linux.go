//go:build linux

package button

import (
	"log/slog"
	"time"

	"github.com/rustyeddy/devices"
	"github.com/rustyeddy/devices/drivers"
	"github.com/warthog618/go-gpiocdev"
)

// Button is a simple momentary GPIO Device that can be detected with
// the Button is pushed (rising edge) or when it is released (falling
// edge). Low is open, high is closed.
type Button struct {
	*drivers.DigitalPin
	devices.Device[int]
	EvtQ chan gpiocdev.LineEvent
}

// New creates a new button with the given name, offset represents the
// pin number and a series of line options. Todo reference the gpiodev
// manual for LineReq options
func New(name string, offset int, opts ...gpiocdev.LineReqOption) *Button {
	var evtQ chan gpiocdev.LineEvent
	evtQ = make(chan gpiocdev.LineEvent)
	bopts := []gpiocdev.LineReqOption{
		gpiocdev.WithPullUp,
		gpiocdev.WithDebounce(10 * time.Millisecond),
		gpiocdev.WithEventHandler(func(evt gpiocdev.LineEvent) {
			evtQ <- evt
		}),
	}

	for _, o := range opts {
		bopts = append(bopts, o)
	}

	b := &Button{
		DigitalPin: drivers.NewDigitalPin(name, offset, bopts...),
	}
	b.Device = b

	b.EvtQ = evtQ
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
