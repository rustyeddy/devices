package sensors

import (
	"context"
	"testing"
	"time"

	"github.com/rustyeddy/devices"
	"github.com/stretchr/testify/require"
)

type fakeTicker struct {
	ch chan time.Time
}

func (f *fakeTicker) C() <-chan time.Time { return f.ch }
func (f *fakeTicker) Stop()               {}

func TestRunPoller_EmitInitialAndTick_DropOnFull(t *testing.T) {
	t.Parallel()

	base := devices.NewBase("sensor", 16)

	out := make(chan int, 1) // intentionally small to test drop-on-full

	ft := &fakeTicker{ch: make(chan time.Time, 10)}
	cfg := PollConfig[int]{
		Interval:       1 * time.Second,
		EmitInitial:    true,
		DropOnFull:     true,
		SampleEventMsg: "sample",
		NewTicker:      func(time.Duration) Ticker { return ft },
		Read: func(ctx context.Context) (int, error) {
			// simple counter-like behavior
			return 42, nil
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- RunPoller[int](ctx, &base, out, cfg)
	}()

	// initial emits once
	require.Equal(t, 42, <-out)

	// fill channel, then tick -> should drop due to DropOnFull
	out <- 99
	ft.ch <- time.Now()

	// channel still has the filled value, no new value should replace it
	require.Equal(t, 99, <-out)

	cancel()
	require.NoError(t, <-errCh)

	_, ok := <-out
	require.False(t, ok, "out should be closed on stop")
}

func TestRunPoller_BlockWhenNotDropOnFull(t *testing.T) {
	t.Parallel()

	base := devices.NewBase("sensor", 16)
	out := make(chan int, 1)

	ft := &fakeTicker{ch: make(chan time.Time, 10)}
	readCalls := 0
	cfg := PollConfig[int]{
		Interval:    1 * time.Second,
		EmitInitial: false,
		DropOnFull:  false, // block
		NewTicker:   func(time.Duration) Ticker { return ft },
		Read: func(ctx context.Context) (int, error) {
			readCalls++
			return readCalls, nil
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- RunPoller[int](ctx, &base, out, cfg)
	}()

	// first tick publishes 1
	ft.ch <- time.Now()
	require.Equal(t, 1, <-out)

	// fill channel
	out <- 777

	// next tick: publish would block, so we must drain to allow it to proceed
	ft.ch <- time.Now()

	// drain fill
	require.Equal(t, 777, <-out)

	// now the blocked publish should land
	require.Equal(t, 2, <-out)

	cancel()
	require.NoError(t, <-errCh)
}
