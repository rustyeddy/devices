package mock

import (
	"context"

	"github.com/rustyeddy/devices"
)

// SwitchConfig configures a mock boolean device.
type SwitchConfig struct {
	Name    string
	Initial bool
	Buf     int // channel buffer size; default 16
}

// Switch is an in-memory duplex bool device.
// Useful for testing higher-level orchestration without hardware.
type Switch struct {
	devices.Base
	in  chan bool
	out chan bool

	state bool
}

// NewSwitch constructs a new mock Switch.
func NewSwitch(cfg SwitchConfig) *Switch {
	if cfg.Buf <= 0 {
		cfg.Buf = 16
	}
	return &Switch{
		Base:  devices.NewBase(cfg.Name, cfg.Buf),
		in:    make(chan bool, cfg.Buf),
		out:   make(chan bool, cfg.Buf),
		state: cfg.Initial,
	}
}

func (s *Switch) In() chan<- bool  { return s.in }
func (s *Switch) Out() <-chan bool { return s.out }

func (s *Switch) Run(ctx context.Context) error {
	s.Emit(devices.EventOpen, "run", nil, nil)

	// publish initial state immediately
	select {
	case s.out <- s.state:
	default:
	}

	defer func() {
		close(s.out)
		s.Emit(devices.EventClose, "stop", nil, nil)
		s.CloseEvents()
	}()

	for {
		select {
		case v := <-s.in:
			s.state = v

			// publish latest state (best-effort)
			select {
			case s.out <- s.state:
			default:
			}

			s.Emit(devices.EventInfo, "set", nil, map[string]string{"value": boolToStr(v)})

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
