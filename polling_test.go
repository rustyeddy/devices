package devices

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRunPoller_EmitInitialAndTick_DropOnFull(t *testing.T) {
	t.Parallel()

	base := NewBase("sensor", 16)
	out := make(chan int, 1) // intentionally small to test drop-on-full

	ft := &FakeTicker{Q: make(chan time.Time, 10)}

	read2 := make(chan struct{}) // closed when the *second* Read() happens
	readCalls := 0

	cfg := PollConfig[int]{
		Interval:       1 * time.Second,
		EmitInitial:    true,
		DropOnFull:     true,
		SampleEventMsg: "sample",
		NewTicker:      func(time.Duration) Ticker { return ft },
		Read: func(ctx context.Context) (int, error) {
			readCalls++
			if readCalls == 2 {
				close(read2)
			}
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

	// fill channel so next publish would drop
	out <- 99

	// trigger a tick and WAIT until the poller processed it (i.e., called Read a second time)
	ft.Q <- time.Now()
	<-read2

	// channel should still have the filled value, no new value should replace it
	require.Equal(t, 99, <-out)

	// stop poller
	cancel()
	require.NoError(t, <-errCh)

	// drain any buffered values, then verify closure
	for range out {
		// drain until closed
	}
}

func TestRunPoller_BlockWhenNotDropOnFull(t *testing.T) {
	t.Parallel()

	base := NewBase("sensor", 16)
	out := make(chan int, 1)

	ft := &FakeTicker{Q: make(chan time.Time, 10)}
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
	ft.Q <- time.Now()
	require.Equal(t, 1, <-out)

	// fill channel
	out <- 777

	// next tick: publish would block, so we must drain to allow it to proceed
	ft.Q <- time.Now()

	// drain fill
	require.Equal(t, 777, <-out)

	// now the blocked publish should land
	require.Equal(t, 2, <-out)

	cancel()
	require.NoError(t, <-errCh)
}

func TestRunPoller_ContextCancelDuringBlockingSend(t *testing.T) {
	t.Parallel()

	base := NewBase("sensor", 16)
	out := make(chan int, 1)

	ft := &FakeTicker{Q: make(chan time.Time, 10)}
	readComplete := make(chan struct{})
	var readOnce sync.Once
	cfg := PollConfig[int]{
		Interval:    1 * time.Second,
		EmitInitial: false,
		DropOnFull:  false, // block when channel is full
		NewTicker:   func(time.Duration) Ticker { return ft },
		Read: func(ctx context.Context) (int, error) {
			readOnce.Do(func() { close(readComplete) })
			return 42, nil
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- RunPoller[int](ctx, &base, out, cfg)
	}()

	// fill the output channel so the next publish will block
	out <- 999

	// trigger a tick to cause a read and subsequent blocked send
	ft.Q <- time.Now()
	<-readComplete // wait for Read to complete

	// give a moment for the publish goroutine to reach the blocked send state
	// this is necessary because we can't directly observe when the goroutine blocks
	time.Sleep(10 * time.Millisecond)

	// cancel context while the send is blocked
	cancel()

	// verify the poller exits without hanging (within reasonable timeout)
	select {
	case err := <-errCh:
		require.NoError(t, err)
	case <-time.After(1 * time.Second):
		t.Fatal("poller did not exit after context cancellation - likely hanging")
	}

	// verify channel still has the original fill value (blocked send was abandoned)
	select {
	case v := <-out:
		require.Equal(t, 999, v)
	default:
		t.Fatal("expected channel to have the fill value")
	}
}
