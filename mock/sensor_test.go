package mock

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSensor_EmitsOnIntervalAndCloses(t *testing.T) {
	t.Parallel()

	s := NewSensor[int](SensorConfig[int]{
		Name:        "counter",
		Interval:    5 * time.Millisecond,
		Initial:     0,
		EmitInitial: true,
		Next: func(curr int) int {
			return curr + 1
		},
	})

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- s.Run(ctx) }()

	// initial
	require.Equal(t, 0, <-s.Out())

	// next couple ticks
	require.Eventually(t, func() bool {
		v := <-s.Out()
		return v >= 1
	}, 250*time.Millisecond, 5*time.Millisecond)

	cancel()
	require.NoError(t, <-errCh)

	// out should be closed
	_, ok := <-s.Out()
	require.False(t, ok)
}
