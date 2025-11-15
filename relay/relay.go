package relay

import (
	"github.com/rustyeddy/devices"
	"github.com/rustyeddy/devices/drivers"
	"github.com/rustyeddy/otto/messanger"
)

type Relay struct {
	drivers.Pin[bool]
	*devices.DeviceBase[bool]
}

func New(name string, index int) (*Relay, error) {
	gpio := drivers.GetGPIO[bool]()
	p, err := gpio.SetPin(name, index, drivers.PinInput)
	if err != nil {
		return nil, err
	}
	relay := &Relay{
		DeviceBase: devices.NewDeviceBase[bool](name),
		Pin:        p,
	}
	return relay, nil
}

func (r *Relay) Get() (bool, error) {
	return r.Pin.Get()
}

func (r *Relay) Set(v bool) error {
	return r.Pin.Set(v)
}

func (r *Relay) HandleMsg(msg *messanger.Msg) error {
	dataStr := string(msg.Data)
	switch dataStr {
	case "on":
		return r.Set(true)
	case "off":
		return r.Set(false)
	default:
		return nil
	}
}
