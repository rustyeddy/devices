package gpio

import (
	"context"

	"github.com/rustyeddy/devices"
	"github.com/rustyeddy/devices/drivers"
)

// LEDConfig configures a GPIO LED (output line).
type LEDConfig struct {
	Name    string
	Factory drivers.Factory
	Chip    string
	Offset  int
	Initial bool
}

// LED controls a GPIO output line intended to drive an LED.
type LED struct {
	devices.Base
	in  chan bool
	out chan bool

	cfg   LEDConfig
	line  drivers.OutputLine
	state bool
}

// NewLED constructs an LED with the given configuration.
func NewLED(cfg LEDConfig) *LED {
	if cfg.Chip == "" {
		cfg.Chip = "gpiochip0"
	}

	return &LED{
		Base:  devices.NewBase(cfg.Name, 16),
		in:    make(chan bool, 16),
		out:   make(chan bool, 16),
		cfg:   cfg,
		state: cfg.Initial,
	}
}

// In returns the command channel for the LED.
func (l *LED) In() chan<- bool { return l.in }

// Out returns the LED state stream.
func (l *LED) Out() <-chan bool { return l.out }

// Descriptor returns the LED metadata.
func (l *LED) Descriptor() devices.Descriptor {
	return devices.Descriptor{
		Name:      l.Name(),
		Kind:      "led",
		ValueType: "bool",
		Access:    devices.ReadWrite,
		Tags:      []string{"gpio", "output", "led"},
		Attributes: map[string]string{
			"chip":   l.cfg.Chip,
			"offset": itoa(l.cfg.Offset),
		},
	}
}

// Run opens the GPIO line and applies LED commands.
func (l *LED) Run(ctx context.Context) error {
	l.Emit(devices.EventOpen, "run", nil, nil)

	if l.cfg.Factory == nil {
		err := devicesErr("led factory is nil")
		l.Emit(devices.EventError, "factory missing", err, nil)
		return err
	}

	line, err := l.cfg.Factory.OpenOutput(l.cfg.Chip, l.cfg.Offset, l.state)
	if err != nil {
		l.Emit(devices.EventError, "open output failed", err, nil)
		return err
	}
	l.line = line

	// publish initial state immediately
	select {
	case l.out <- l.state:
	default:
	}

	defer func() {
		_ = l.line.Close()
		close(l.out)
		l.Emit(devices.EventClose, "stop", nil, nil)
		l.Close()
	}()

	for {
		select {
		case v := <-l.in:
			l.state = v

			if err := l.line.Write(v); err != nil {
				l.Emit(devices.EventError, "write failed", err, nil)
				continue
			}

			select {
			case l.out <- l.state:
			default:
			}
			l.Emit(devices.EventInfo, "set", nil, map[string]string{"value": boolToStr(v)})

		case <-ctx.Done():
			return nil
		}
	}
}
