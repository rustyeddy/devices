//go:build linux

package drivers

import (
	"context"
	"fmt"
	"time"

	"github.com/warthog618/go-gpiocdev"
)

// GPIOCDevFactory opens GPIO lines using go-gpiocdev.
type GPIOCDevFactory struct{}

// NewGPIOCDevFactory returns a GPIOCDev-backed Factory.
func NewGPIOCDevFactory() Factory { return &GPIOCDevFactory{} }

// OpenInput requests a GPIO line for edge events.
func (f *GPIOCDevFactory) OpenInput(chip string, offset int, edge Edge, bias Bias, debounce time.Duration) (InputLine, error) {
	if chip == "" {
		chip = "gpiochip0"
	}

	opts := []gpiocdev.LineReqOption{
		gpiocdev.AsInput,
	}

	// bias
	switch bias {
	case BiasPullUp:
		opts = append(opts, gpiocdev.WithPullUp)
	case BiasPullDown:
		opts = append(opts, gpiocdev.WithPullDown)
	}

	// edges + handler
	evtQ := make(chan gpiocdev.LineEvent, 32)

	// configure edges
	switch edge {
	case EdgeBoth:
		opts = append(opts, gpiocdev.WithBothEdges)
	case EdgeRising:
		// go-gpiocdev does rising via event handler + rising config (no separate option),
		// but WithBothEdges is simplest. For “rising only”, we filter later.
		opts = append(opts, gpiocdev.WithBothEdges)
	case EdgeFalling:
		opts = append(opts, gpiocdev.WithBothEdges)
	case EdgeNone:
		// no edge options
	default:
		opts = append(opts, gpiocdev.WithBothEdges)
	}

	// debounce
	if debounce > 0 && edge != EdgeNone {
		opts = append(opts, gpiocdev.WithDebounce(debounce))
	}

	// per-line handler pushes into evtQ
	opts = append(opts, gpiocdev.WithEventHandler(func(evt gpiocdev.LineEvent) {
		// non-blocking to avoid wedging kernel event delivery
		select {
		case evtQ <- evt:
		default:
			// drop if overwhelmed
		}
	}))

	line, err := gpiocdev.RequestLine(chip, offset, opts...)
	if err != nil {
		return nil, fmt.Errorf("request input line %s:%d: %w", chip, offset, err)
	}

	return &gpiocdevInputLine{
		line: line,
		edge: edge,
		evtQ: evtQ,
	}, nil
}

// OpenOutput requests a GPIO line for output.
func (f *GPIOCDevFactory) OpenOutput(chip string, offset int, initial bool) (OutputLine, error) {
	if chip == "" {
		chip = "gpiochip0"
	}

	initVal := 0
	if initial {
		initVal = 1
	}

	line, err := gpiocdev.RequestLine(chip, offset, gpiocdev.AsOutput(initVal))
	if err != nil {
		return nil, fmt.Errorf("request output line %s:%d: %w", chip, offset, err)
	}

	return &gpiocdevOutputLine{line: line}, nil
}

// gpiocdevInputLine wraps a gpiocdev Line for input events.
type gpiocdevInputLine struct {
	line *gpiocdev.Line
	edge Edge
	evtQ chan gpiocdev.LineEvent
}

func (l *gpiocdevInputLine) Read() (bool, error) {
	v, err := l.line.Value()
	if err != nil {
		return false, err
	}
	return v != 0, nil
}

func (l *gpiocdevInputLine) Events(ctx context.Context) (<-chan LineEvent, error) {
	out := make(chan LineEvent, 32)

	go func() {
		defer close(out)
		for {
			select {
			case evt, ok := <-l.evtQ:
				if !ok {
					return
				}

				edge := EdgeNone
				switch evt.Type {
				case gpiocdev.LineEventRisingEdge:
					edge = EdgeRising
				case gpiocdev.LineEventFallingEdge:
					edge = EdgeFalling
				}

				// filter if caller asked for only rising/falling
				if l.edge == EdgeRising && edge != EdgeRising {
					continue
				}
				if l.edge == EdgeFalling && edge != EdgeFalling {
					continue
				}
				if l.edge == EdgeNone {
					continue
				}

				val, _ := l.Read() // best effort; some kernels include state in event, but simplest is re-read
				select {
				//				case out <- LineEvent{Time: evt.Timestamp, Edge: edge, Value: val}:
				case out <- LineEvent{Time: time.Now(), Edge: edge, Value: val}:
				default:
					// drop if consumer slow
				}

			case <-ctx.Done():
				return
			}
		}
	}()

	return out, nil
}

func (l *gpiocdevInputLine) Close() error {
	if l.line != nil {
		_ = l.line.Reconfigure(gpiocdev.AsInput)
		_ = l.line.Close()
		l.line = nil
	}
	if l.evtQ != nil {
		close(l.evtQ)
		l.evtQ = nil
	}
	return nil
}

// gpiocdevOutputLine wraps a gpiocdev Line for output writes.
type gpiocdevOutputLine struct {
	line *gpiocdev.Line
}

func (l *gpiocdevOutputLine) Write(v bool) error {
	val := 0
	if v {
		val = 1
	}
	return l.line.SetValue(val)
}

func (l *gpiocdevOutputLine) Close() error {
	if l.line != nil {
		_ = l.line.Close()
		l.line = nil
	}
	return nil
}
