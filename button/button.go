package button

import (
	"github.com/rustyeddy/devices"
	"github.com/rustyeddy/devices/drivers"

)

type Button struct {
	drivers.Pin[bool]
	*devices.DeviceBase[bool]
}

func New(name string, index int, opts ...drivers.PinOptions) (*Button, error) {
	gpio := drivers.GetGPIO[bool]()
	p, err := gpio.SetPin(name, index, drivers.PinInput)
	if err != nil {
		return nil, err
	}
	b := &Button{
		DeviceBase: devices.NewDeviceBase[bool](name),
		Pin: p,
	}
	return b, nil
}

// func (b *Button) Name() string {
// 	return b.name
// }

// func (b *Button) Index() int {
// 	return b.index
// }

// func (b *Button) Open() error {

// 	v := GetVPIO[bool]()
// 	_, err := v.Pin(b.name, uint(b.index), DirectionInput)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (b *Button) Close() error {
// 	v := GetVPIO[bool]()
// 	return v.Close(b.index)
// }
