package button

import (
	"context"
	"time"

	"github.com/rustyeddy/devices"
	"github.com/rustyeddy/devices/drivers"
)

// ButtonConfig configures a GPIO button.
type ButtonConfig struct {
	Name     string
	Factory  drivers.Factory
	Chip     string // default "gpiochip0" on linux
	Offset   int
	Bias     drivers.Bias
	Edge     drivers.Edge
	Debounce time.Duration
}

// Button reports a GPIO input line as a boolean stream.
type Button struct {
	devices.Base
	out chan bool

	cfg  ButtonConfig
	line drivers.InputLine
}

// NewButton constructs a Button with defaults applied.
func NewButton(cfg ButtonConfig) *Button {
	if cfg.Debounce == 0 {
		cfg.Debounce = 30 * time.Millisecond
	}
	if cfg.Edge == "" {
		cfg.Edge = drivers.EdgeBoth
	}
	if cfg.Bias == "" {
		cfg.Bias = drivers.BiasPullUp
	}
	return &Button{
		Base: devices.NewBase(cfg.Name, 16),
		out:  make(chan bool, 16),
		cfg:  cfg,
	}
}

// Out returns the button state stream.
func (b *Button) Out() <-chan bool { return b.out }

// Descriptor returns the button metadata.
func (b *Button) Descriptor() devices.Descriptor {
	return devices.Descriptor{
		Name:      b.Name(),
		Kind:      "button",
		ValueType: "bool",
		Access:    devices.ReadOnly,
		Tags:      []string{"gpio", "input"},
		Attributes: map[string]string{
			"chip":     b.cfg.Chip,
			"offset":   itoa(b.cfg.Offset),
			"bias":     string(b.cfg.Bias),
			"edge":     string(b.cfg.Edge),
			"debounce": b.cfg.Debounce.String(),
		},
	}
}

// Run opens the GPIO line and emits button state changes.
func (b *Button) Run(ctx context.Context) error {
	b.Emit(devices.EventOpen, "run", nil, nil)

	if b.cfg.Factory == nil {
		err := devicesErr("button factory is nil")
		b.Emit(devices.EventError, "factory missing", err, nil)
		return err
	}

	line, err := b.cfg.Factory.OpenInput(b.cfg.Chip, b.cfg.Offset, b.cfg.Edge, b.cfg.Bias, b.cfg.Debounce)
	if err != nil {
		b.Emit(devices.EventError, "open input failed", err, nil)
		return err
	}
	b.line = line

	defer func() {
		_ = b.line.Close()
		close(b.out)
		b.Emit(devices.EventClose, "stop", nil, nil)
		b.Close()
	}()

	// Emit initial state so MQTT state is meaningful immediately.  Todo, what if
	initial, err := b.line.Read()
	if err == nil {
		select {
		case b.out <- initial:
		default:
		}
	}

	evCh, err := b.line.Events(ctx)
	if err != nil {
		b.Emit(devices.EventError, "events failed", err, nil)
		return err
	}

	var state = initial
	var last time.Time

	for {
		select {
		case ev, ok := <-evCh:
			if !ok {
				return nil
			}

			// Extra debounce guard (kernel debounce may already handle it, but harmless)
			if b.cfg.Debounce > 0 && !last.IsZero() && ev.Time.Sub(last) < b.cfg.Debounce {
				continue
			}
			last = ev.Time

			state = ev.Value
			select {
			case b.out <- state:
			default:
			}

			b.Emit(devices.EventEdge, "edge", nil, map[string]string{"edge": string(ev.Edge)})

		case <-ctx.Done():
			return nil
		}
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	sign := ""
	if n < 0 {
		sign = "-"
		n = -n
	}
	var buf [12]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + (n % 10))
		n /= 10
	}
	return sign + string(buf[i:])
}

type devicesErr string

func (e devicesErr) Error() string { return string(e) }
