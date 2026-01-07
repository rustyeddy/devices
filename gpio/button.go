package gpio

import (
	"context"
	"time"

	"github.com/rustyeddy/devices"
)

type Button struct {
	name   string
	out    chan bool
	events chan devices.Event
}

func (b *Button) Name() string                 { return b.name }
func (b *Button) Out() <-chan bool             { return b.out }
func (b *Button) Events() <-chan devices.Event { return b.events }

func (b *Button) Run(ctx context.Context) error {
	defer close(b.out)
	defer close(b.events)

	b.events <- devices.Event{Device: b.name, Kind: devices.EventOpen, Time: time.Now()}

	for {
		select {
		case <-ctx.Done():
			b.events <- devices.Event{Device: b.name, Kind: devices.EventClose, Time: time.Now()}
			return nil

			// hardware polling / interrupt handler here
		}
	}
}

func NewButton(name string) *Button {
	return &Button{
		name:   name,
		out:    make(chan bool),
		events: make(chan devices.Event),
	}
}

// func New(name string, index int, opts ...drivers.PinOptions) (*Button, error) {
// 	gpio := drivers.GetGPIO[bool]()
// 	p, err := gpio.SetPin(name, index, drivers.PinInput)
// 	if err != nil {
// 		return nil, err
// 	}
// 	b := &Button{
// 		DeviceBase: devices.NewDeviceBase[bool](name),
// 		Pin:        p,
// 	}
// 	return b, nil
// }
