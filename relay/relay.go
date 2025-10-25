package relay

import (
	"github.com/rustyeddy/devices"
	"github.com/rustyeddy/devices/drivers"
	"github.com/warthog618/go-gpiocdev"
)

type Relay struct {
	*drivers.DigitalPin
	devices.Device[int]
}

func New(name string, offset int) *Relay {
	relay := &Relay{}
	g := drivers.GetGPIO()
	relay.DigitalPin = g.Pin(name, offset, gpiocdev.AsOutput(0))
	relay.Device = relay
	return relay
}

func (r *Relay) ID() string {
	return r.DigitalPin.ID()
}

func (r *Relay) Open() error {
	return nil
}

func (r *Relay) Close() error {
	return nil
}

func (r *Relay) Get() (int, error) {
	v, err := r.DigitalPin.Get()
	return v, err
}

func (r *Relay) Set(v int) error {
	return r.DigitalPin.Set(v)
}

func (r *Relay) Type() devices.Type {
	return devices.TypeInt
}

func (r *Relay) On() error {
	return r.DigitalPin.ON()
}

func (r *Relay) Off() error {
	return r.DigitalPin.OFF()
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
