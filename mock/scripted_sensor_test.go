package mock

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestScriptedSensor_Step_StopWhenDone(t *testing.T) {
	t.Parallel()

	s := NewScriptedSensor[int](ScriptedSensorConfig[int]{
		Name:         "script",
		Step:         true,
		Values:       []int{10, 11, 12},
		StopWhenDone: true,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() { errCh <- s.Run(ctx) }()

	// drive emissions deterministically
	s.In() <- struct{}{}
	require.Equal(t, 10, <-s.Out())

	s.In() <- struct{}{}
	require.Equal(t, 11, <-s.Out())

	s.In() <- struct{}{}
	require.Equal(t, 12, <-s.Out())

	// give Run() a moment to exit after final emission
	require.Eventually(t, func() bool {
		select {
		case err := <-errCh:
			return err == nil
		default:
			return false
		}
	}, 250*time.Millisecond, 5*time.Millisecond)

	_, ok := <-s.Out()
	require.False(t, ok)
}

func TestScriptedSensor_Step_RepeatLast(t *testing.T) {
	t.Parallel()

	s := NewScriptedSensor[int](ScriptedSensorConfig[int]{
		Name:       "script",
		Step:       true,
		Values:     []int{1, 2},
		RepeatLast: true,
	})

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- s.Run(ctx) }()

	s.In() <- struct{}{}
	require.Equal(t, 1, <-s.Out())

	s.In() <- struct{}{}
	require.Equal(t, 2, <-s.Out())

	// now repeats last forever (per step)
	s.In() <- struct{}{}
	require.Equal(t, 2, <-s.Out())

	cancel()
	require.NoError(t, <-errCh)
}
