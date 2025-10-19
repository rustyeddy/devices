package led

import (
	"errors"

	"github.com/rustyeddy/devices"
	"github.com/rustyeddy/devices/drivers"
	"github.com/warthog618/go-gpiocdev"
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
	led := &LED{
		id:  id,
		pin: pin,
	}
	return led
}

func (l *LED) Open() error {
	g := drivers.GetGPIO()
	l.DigitalPin = g.Pin(l.id, l.pin, gpiocdev.AsOutput(0))
	return nil
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
