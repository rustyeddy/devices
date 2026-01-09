package gpio

import (
	"context"
	"errors"

	"github.com/rustyeddy/devices"
	"github.com/rustyeddy/devices/drivers"
)

type RelayConfig struct {
	Name    string
	Factory drivers.Factory
	Chip    string
	Offset  int
	Initial bool
}

type Relay struct {
	devices.Base
	in  chan bool
	out chan bool

	cfg   RelayConfig
	line  drivers.OutputLine
	state bool
}

func NewRelay(cfg RelayConfig) *Relay {
	return &Relay{
		Base:  devices.NewBase(cfg.Name, 16),
		in:    make(chan bool, 16),
		out:   make(chan bool, 16),
		cfg:   cfg,
		state: cfg.Initial,
	}
}

func (r *Relay) In() chan<- bool  { return r.in }
func (r *Relay) Out() <-chan bool { return r.out }

func (r *Relay) Descriptor() devices.Descriptor {
	return devices.Descriptor{
		Name:      r.Name(),
		Kind:      "relay",
		ValueType: "bool",
		Access:    devices.ReadWrite,
		Tags:      []string{"gpio", "output"},
		Attributes: map[string]string{
			"chip":   r.cfg.Chip,
			"offset": itoa(r.cfg.Offset),
		},
	}
}

func (r *Relay) Run(ctx context.Context) error {
	r.Emit(devices.EventOpen, "run", nil, nil)

	if r.cfg.Factory == nil {
		err := errors.New("relay factory is nil")
		r.Emit(devices.EventError, "factory missing", err, nil)
		return err
	}

	line, err := r.cfg.Factory.OpenOutput(r.cfg.Chip, r.cfg.Offset, r.state)
	if err != nil {
		r.Emit(devices.EventError, "open output failed", err, nil)
		return err
	}
	r.line = line

	// publish initial state immediately
	select {
	case r.out <- r.state:
	default:
	}

	defer func() {
		_ = r.line.Close()
		close(r.out)
		r.Emit(devices.EventClose, "stop", nil, nil)
		r.CloseEvents()
	}()

	for {
		select {
		case v := <-r.in:
			r.state = v

			if err := r.line.Write(v); err != nil {
				r.Emit(devices.EventError, "write failed", err, nil)
				continue
			}

			select {
			case r.out <- r.state:
			default:
			}
			r.Emit(devices.EventInfo, "set", nil, map[string]string{"value": boolToStr(v)})

		case <-ctx.Done():
			return nil
		}
	}
}

func boolToStr(v bool) string {
	if v {
		return "true"
	}
	return "false"
}
