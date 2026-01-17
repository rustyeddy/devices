package mock

import (
	"context"
	"errors"

	"github.com/rustyeddy/devices"
)

// ScriptedSensorConfig emits a fixed sequence of values.
//
// Modes:
//   - If Step is false: emits the sequence immediately (best-effort) then either stops or waits for ctx.
//   - If Step is true: waits for a "tick" on In() before emitting the next value.
//
// Completion:
//   - If RepeatLast is true: after the sequence is exhausted, keeps re-emitting the last value (on each step, if Step=true).
//   - If StopWhenDone is true: returns nil after emitting the last value.
//     (If both RepeatLast and StopWhenDone are false, it will just idle until ctx is canceled after finishing.)
type ScriptedSensorConfig[T any] struct {
	Name string
	Buf  int // default 16

	// Values is the sequence emitted, in order. Must be non-empty unless EmitInitial is true.
	Values []T

	// Initial is emitted first if EmitInitial is true.
	EmitInitial bool
	Initial     T

	// Step controls whether emission is driven by In() "ticks".
	Step bool

	// Completion behavior
	RepeatLast   bool
	StopWhenDone bool
}

// ScriptedSensor is a deterministic mock sensor for tests.
type ScriptedSensor[T any] struct {
	devices.Base
	in  chan struct{}
	out chan T

	cfg   ScriptedSensorConfig[T]
	index int
	done  bool

	last    T
	hasLast bool
}

func NewScriptedSensor[T any](cfg ScriptedSensorConfig[T]) *ScriptedSensor[T] {
	if cfg.Buf <= 0 {
		cfg.Buf = 16
	}
	return &ScriptedSensor[T]{
		Base: devices.NewBase(cfg.Name, cfg.Buf),
		in:   make(chan struct{}, cfg.Buf),
		out:  make(chan T, cfg.Buf),
		cfg:  cfg,
	}
}

// In returns the step channel (only used when cfg.Step = true).
// Sending any value (struct{}{}) advances the script by one emission.
func (s *ScriptedSensor[T]) In() chan<- struct{} { return s.in }

// Out returns the output stream.
func (s *ScriptedSensor[T]) Out() <-chan T { return s.out }

func (s *ScriptedSensor[T]) Run(ctx context.Context) error {
	s.Emit(devices.EventOpen, "run", nil, nil)

	if !s.cfg.EmitInitial && len(s.cfg.Values) == 0 {
		err := errors.New("scripted sensor requires Values or EmitInitial")
		s.Emit(devices.EventError, "invalid config", err, nil)
		return err
	}

	emit := func(v T, msg string) {
		s.last = v
		s.hasLast = true

		// best-effort publish
		select {
		case s.out <- v:
		default:
		}
		s.Emit(devices.EventInfo, msg, nil, nil)
	}

	// Emit initial if requested
	if s.cfg.EmitInitial {
		emit(s.cfg.Initial, "initial")
	}

	defer func() {
		close(s.out)
		s.Emit(devices.EventClose, "stop", nil, nil)
		s.Close()
	}()

	// Helper to emit next scripted value if any remain
	emitNext := func() (finished bool) {
		if s.index < len(s.cfg.Values) {
			emit(s.cfg.Values[s.index], "emit")
			s.index++
			if s.index >= len(s.cfg.Values) {
				s.done = true
			}
			return false
		}
		// already exhausted
		s.done = true
		return true
	}

	// Non-step mode: emit everything immediately (best-effort)
	if !s.cfg.Step {
		for s.index < len(s.cfg.Values) {
			emitNext()
		}

		if s.cfg.StopWhenDone {
			return nil
		}

		// If repeating last, just wait and re-emit on ctx? (no clock) — we’ll simply idle.
		// RepeatLast is mainly useful in Step mode.
		<-ctx.Done()
		return nil
	}

	// Step mode: each tick emits one value (or repeats last)
	for {
		select {
		case <-s.in:
			if s.done {
				if s.cfg.RepeatLast && s.hasLast {
					emit(s.last, "repeat")
					continue
				}
				if s.cfg.StopWhenDone {
					return nil
				}
				// otherwise idle after done
				continue
			}
			emitNext()

			// If we just finished and StopWhenDone is set, stop immediately.
			if s.done && s.cfg.StopWhenDone {
				return nil
			}

		case <-ctx.Done():
			return nil
		}
	}
}
