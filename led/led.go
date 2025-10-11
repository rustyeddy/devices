package led

import (
	"github.com/rustyeddy/devices"
	"github.com/rustyeddy/devices/drivers"
	"github.com/warthog618/go-gpiocdev"
)

type LED struct {
	*devices.Device
	*drivers.DigitalPin
}

func New(name string, offset int) *LED {
	led := &LED{
		Device: devices.NewDevice(name, "mqtt"),
	}
	g := drivers.GetGPIO()
	led.DigitalPin = g.Pin(name, offset, gpiocdev.AsOutput(0))
	return led
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
