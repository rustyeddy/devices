package mock

import (
	"context"
	"errors"
	"time"

	"github.com/rustyeddy/devices"
)

// SensorConfig configures a mock sensor that emits values on an interval.
type SensorConfig[T any] struct {
	Name string

	// Interval must be > 0.
	Interval time.Duration

	// Buf controls channel buffer sizes; default 16.
	Buf int

	// Initial is the starting value.
	Initial T

	// Next produces the next value given the current value.
	// If nil, the sensor will repeatedly emit Initial.
	Next func(curr T) T

	// EmitInitial controls whether Initial is emitted immediately on start.
	EmitInitial bool
}

// Sensor is a mock source device that emits values on a timer.
// Useful for testing orchestration without hardware.
type Sensor[T any] struct {
	devices.Base
	out chan T

	cfg   SensorConfig[T]
	state T
}

// NewSensor constructs a mock sensor.
func NewSensor[T any](cfg SensorConfig[T]) *Sensor[T] {
	if cfg.Buf <= 0 {
		cfg.Buf = 16
	}
	return &Sensor[T]{
		Base:  devices.NewBase(cfg.Name, cfg.Buf),
		out:   make(chan T, cfg.Buf),
		cfg:   cfg,
		state: cfg.Initial,
	}
}

// Out returns the sensor output stream.
func (s *Sensor[T]) Out() <-chan T { return s.out }

// Run starts emitting sensor values until ctx is canceled.
func (s *Sensor[T]) Run(ctx context.Context) error {
	s.Emit(devices.EventOpen, "run", nil, nil)

	if s.cfg.Interval <= 0 {
		err := errors.New("mock sensor interval must be > 0")
		s.Emit(devices.EventError, "invalid interval", err, nil)
		return err
	}

	t := time.NewTicker(s.cfg.Interval)
	defer t.Stop()

	emit := func(v T) {
		// best-effort publish
		select {
		case s.out <- v:
		default:
		}
	}

	if s.cfg.EmitInitial {
		emit(s.state)
	}

	defer func() {
		close(s.out)
		s.Emit(devices.EventClose, "stop", nil, nil)
		s.Close()
	}()

	for {
		select {
		case <-t.C:
			if s.cfg.Next != nil {
				s.state = s.cfg.Next(s.state)
			}
			emit(s.state)
			s.Emit(devices.EventInfo, "tick", nil, nil)

		case <-ctx.Done():
			return nil
		}
	}
}
