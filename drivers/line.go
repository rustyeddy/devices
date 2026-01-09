package drivers

import (
	"context"
	"time"
)

type Bias string

const (
	BiasDefault  Bias = "default"
	BiasPullUp   Bias = "pullup"
	BiasPullDown Bias = "pulldown"
)

type Edge string

const (
	EdgeNone    Edge = "none"
	EdgeRising  Edge = "rising"
	EdgeFalling Edge = "falling"
	EdgeBoth    Edge = "both"
)

type LineEvent struct {
	Time  time.Time
	Edge  Edge
	Value bool
}

type InputLine interface {
	Read() (bool, error)
	Events(ctx context.Context) (<-chan LineEvent, error)
	Close() error
}

type OutputLine interface {
	Write(v bool) error
	Close() error
}

type Factory interface {
	OpenInput(chip string, offset int, edge Edge, bias Bias, debounce time.Duration) (InputLine, error)
	OpenOutput(chip string, offset int, initial bool) (OutputLine, error)
}
