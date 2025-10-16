package relay

import (
	"github.com/rustyeddy/devices"
	"github.com/rustyeddy/devices/drivers"
	"github.com/warthog618/go-gpiocdev"
)

type Relay struct {
	*devices.Device
	*drivers.DigitalPin
}

func New(name string, offset int) *Relay {
	relay := &Relay{
		Device: devices.NewDevice(name, name),
	}
	g := drivers.GetGPIO()
	relay.DigitalPin = g.Pin(name, offset, gpiocdev.AsOutput(0))
	return relay
}

func (r *Relay) Callback(val bool) {
	switch val {
	case false:
		r.Off()

	case true:
		r.On()
	}
	return
}
