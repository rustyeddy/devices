package gpio

import (
	"context"

	"github.com/rustyeddy/devices"
)

type Relay struct {
	name   string
	in     chan bool
	out    chan bool
	events chan devices.Event
}

func (r *Relay) Name() string                 { return r.name }
func (r *Relay) In() chan<- bool              { return r.in }
func (r *Relay) Out() <-chan bool             { return r.out }
func (r *Relay) Events() <-chan devices.Event { return r.events }

func (r *Relay) Run(ctx context.Context) error {
	defer close(r.out)
	defer close(r.events)

	state := false
	for {
		select {
		case v := <-r.in:
			state = v
			// apply GPIO write
			select {
			case r.out <- state:
			default:
			}
			r.events <- devices.Event{Device: r.name, Kind: devices.EventInfo, Msg: "set"}

		case <-ctx.Done():
			return nil
		}
	}
}

func NewRelay(name string) *Relay {
	return &Relay{
		name:   name,
		in:     make(chan bool),
		out:    make(chan bool),
		events: make(chan devices.Event),
	}
}

// func New(name string, index int) (*Relay, error) {
// 	gpio := drivers.GetGPIO[bool]()
// 	p, err := gpio.SetPin(name, index, drivers.PinInput)
// 	if err != nil {
// 		return nil, err
// 	}
// 	relay := &Relay{
// 		DeviceBase: devices.NewDeviceBase[bool](name),
// 		Pin:        p,
// 	}
// 	return relay, nil
// }

// func (r *Relay) Get() (bool, error) {
// 	return r.Pin.Get()
// }

// func (r *Relay) Set(v bool) error {
// 	return r.Pin.Set(v)
// }
