package relay

import (
	"context"
	"errors"

	"github.com/rustyeddy/devices"
	"github.com/rustyeddy/devices/drivers"
	"github.com/rustyeddy/devices/sensors"
)

// RelayConfig configures a GPIO relay.
type RelayConfig struct {
	Name    string
	Factory drivers.Factory
	Chip    string
	Offset  int
	Initial bool
}

// Relay controls a GPIO output line.
type Relay struct {
	devices.Base
	in  chan bool
	out chan bool

	cfg   RelayConfig
	line  drivers.OutputLine
	state bool
}

// NewRelay constructs a Relay with the given configuration.
func NewRelay(cfg RelayConfig) *Relay {
	return &Relay{
		Base:  devices.NewBase(cfg.Name, 16),
		in:    make(chan bool, 16),
		out:   make(chan bool, 16),
		cfg:   cfg,
		state: cfg.Initial,
	}
}

// In returns the command channel for the relay.
func (r *Relay) In() chan<- bool { return r.in }

// Out returns the state stream for the relay.
func (r *Relay) Out() <-chan bool { return r.out }

// Descriptor returns the relay metadata.
func (r *Relay) Descriptor() devices.Descriptor {
	return devices.Descriptor{
		Name:      r.Name(),
		Kind:      "relay",
		ValueType: "bool",
		Access:    devices.ReadWrite,
		Tags:      []string{"gpio", "output"},
		Attributes: map[string]string{
			"chip":   r.cfg.Chip,
			"offset": sensors.Itoa(r.cfg.Offset),
		},
	}
}

// Run opens the GPIO line and applies relay commands.
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
		r.Close()
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
