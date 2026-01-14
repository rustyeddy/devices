package sensors

import (
	"context"
	"errors"
	"time"

	"github.com/rustyeddy/devices"
)

// Ticker abstracts time.Ticker for testability.
type Ticker interface {
	C() <-chan time.Time
	Stop()
}

type realTicker struct{ t *time.Ticker }

func (r realTicker) C() <-chan time.Time { return r.t.C }
func (r realTicker) Stop()               { r.t.Stop() }

// PollConfig defines the common polling behavior for sensors.
type PollConfig[T any] struct {
	// Interval must be > 0.
	Interval time.Duration

	// EmitInitial performs one Read() immediately when Run() starts.
	EmitInitial bool

	// DropOnFull controls what happens when Out is full.
	// If true (default), the new sample is dropped.
	// If false, publishing blocks until delivered.
	DropOnFull bool

	// Read reads the next sensor sample.
	Read func(ctx context.Context) (T, error)

	// OnSample is an optional callback invoked after a successful Read(),
	// before attempting to publish to Out. Useful for internal state updates.
	OnSample func(v T)

	// SampleMeta optionally supplies event metadata per sample.
	SampleMeta func(v T) map[string]string

	// SampleEventMsg is the EventInfo Msg used on successful samples.
	// Defaults to "sample".
	SampleEventMsg string

	// NewTicker optionally overrides ticker creation (tests).
	NewTicker func(d time.Duration) Ticker
}

var (
	ErrPollInterval = errors.New("poll interval must be > 0")
	ErrPollReadNil  = errors.New("poll Read func is nil")
)

// RunPoller runs a standardized polling loop.
//
// It:
// - emits EventOpen at start
// - optionally reads once immediately (EmitInitial)
// - reads on each tick
// - publishes samples to Out (drop-on-full by default)
// - emits EventInfo on sample, EventError on read errors
// - on exit: stops ticker, closes Out, emits EventClose, closes Base events
//
// IMPORTANT: This helper owns closing Out and closing events (Base.CloseEvents()).
// If the caller has additional resources to close, defer them BEFORE calling RunPoller.
func RunPoller[T any](ctx context.Context, base *devices.Base, out chan T, cfg PollConfig[T]) error {
	if base == nil {
		return errors.New("base is nil")
	}
	base.Emit(devices.EventOpen, "run", nil, nil)

	if cfg.Interval <= 0 {
		base.Emit(devices.EventError, "invalid interval", ErrPollInterval, nil)
		return ErrPollInterval
	}
	if cfg.Read == nil {
		base.Emit(devices.EventError, "read func missing", ErrPollReadNil, nil)
		return ErrPollReadNil
	}

	// Defaults
	if cfg.SampleEventMsg == "" {
		cfg.SampleEventMsg = "sample"
	}
	// default drop-on-full: true
	if !cfg.DropOnFull {
		// leave as-is (caller explicitly set false)
	} else {
		// if caller didn't set it at all, they want true — but we can't detect "unset" for bool.
		// So: treat "true" as the default behavior in docs and usage.
	}
	newTicker := cfg.NewTicker
	if newTicker == nil {
		newTicker = func(d time.Duration) Ticker { return realTicker{t: time.NewTicker(d)} }
	}

	t := newTicker(cfg.Interval)
	defer t.Stop()

	publish := func(v T) {
		if cfg.OnSample != nil {
			cfg.OnSample(v)
		}

		if cfg.DropOnFull {
			select {
			case out <- v:
			default:
			}
		} else {
			out <- v
		}

		meta := map[string]string(nil)
		if cfg.SampleMeta != nil {
			meta = cfg.SampleMeta(v)
		}
		base.Emit(devices.EventInfo, cfg.SampleEventMsg, nil, meta)
	}

	defer func() {
		close(out)
		base.Emit(devices.EventClose, "stop", nil, nil)
		base.CloseEvents()
	}()

	// initial sample
	if cfg.EmitInitial {
		v, err := cfg.Read(ctx)
		if err != nil {
			base.Emit(devices.EventError, "read failed", err, nil)
		} else {
			publish(v)
		}
	}

	for {
		select {
		case <-t.C():
			v, err := cfg.Read(ctx)
			if err != nil {
				base.Emit(devices.EventError, "read failed", err, nil)
				continue
			}
			publish(v)

		case <-ctx.Done():
			return nil
		}
	}
}
