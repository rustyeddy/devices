package drivers

import (
	"context"
	"time"
)

type Bias string

const (
	// BiasDefault leaves bias configuration to the driver default.
	BiasDefault Bias = "default"
	// BiasPullUp enables an internal pull-up if supported.
	BiasPullUp Bias = "pullup"
	// BiasPullDown enables an internal pull-down if supported.
	BiasPullDown Bias = "pulldown"
)

// Edge selects which transitions are reported.
type Edge string

const (
	// EdgeNone disables edge events.
	EdgeNone Edge = "none"
	// EdgeRising reports low-to-high transitions.
	EdgeRising Edge = "rising"
	// EdgeFalling reports high-to-low transitions.
	EdgeFalling Edge = "falling"
	// EdgeBoth reports rising and falling transitions.
	EdgeBoth Edge = "both"
)

// LineEvent describes a GPIO edge notification.
type LineEvent struct {
	Time  time.Time
	Edge  Edge
	Value bool
}

// InputLine reads values and emits edge events.
type InputLine interface {
	Read() (bool, error)
	Events(ctx context.Context) (<-chan LineEvent, error)
	Close() error
}

// OutputLine writes values to a GPIO line.
type OutputLine interface {
	Write(v bool) error
	Close() error
}

// Factory opens GPIO input and output lines.
type Factory interface {
	OpenInput(chip string, offset int, edge Edge, bias Bias, debounce time.Duration) (InputLine, error)
	OpenOutput(chip string, offset int, initial bool) (OutputLine, error)
}
