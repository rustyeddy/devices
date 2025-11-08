//go:build !linux

package led

import (
	"errors"
	"log/slog"

	"github.com/rustyeddy/devices"
	"github.com/rustyeddy/devices/drivers"
)

type LED struct {
	id  string
	pin int

	On  func()
	Off func()

	*drivers.DigitalPin
	devices.Device[int]
}

func New(id string, pin int) *LED {
	slog.Info("Creating stub LED (non-Linux platform)", "id", id, "pin", pin)
	led := &LED{
		id:  id,
		pin: pin,
	}
	led.Device = led
	return led
}

func (l *LED) Open() error {
	l.DigitalPin = drivers.NewDigitalPin(l.id, l.pin, drivers.AsOutput(0))
	return nil
}

func (l *LED) Get() (int, error) {
	v, err := l.DigitalPin.Get()
	return v, err
}

func (l *LED) Set(v int) error {
	err := l.DigitalPin.Set(v)
	return err
}

func (l *LED) Close() error {
	return errors.New("TODO Need to implement LED close")
}

func (l *LED) ID() string {
	return l.id
}

func (l *LED) Type() devices.Type {
	return devices.TypeInt
}

func (l *LED) Callback(val bool) {
	switch val {
	case false:
		l.Off()

	case true:
		l.On()
	}
	return
}
