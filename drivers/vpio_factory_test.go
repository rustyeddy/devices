package drivers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVPIOFactoryReadWrite(t *testing.T) {
	t.Parallel()

	f := NewVPIOFactory()
	out, err := f.OpenOutput("chip0", 7, false)
	require.NoError(t, err)

	require.NoError(t, out.Write(true))

	in, err := f.OpenInput("chip0", 7, EdgeNone, BiasDefault, 0)
	require.NoError(t, err)

	val, err := in.Read()
	require.NoError(t, err)
	assert.True(t, val)
}

func TestVPIOFactoryEvents(t *testing.T) {
	t.Parallel()

	f := NewVPIOFactory()
	in, err := f.OpenInput("chip0", 3, EdgeRising, BiasDefault, 0)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	events, err := in.Events(ctx)
	require.NoError(t, err)

	f.InjectEdge("chip0", 3, EdgeFalling, false)
	f.InjectEdge("chip0", 3, EdgeRising, true)

	ev := <-events
	assert.Equal(t, EdgeRising, ev.Edge)
	assert.True(t, ev.Value)
	assert.False(t, ev.Time.IsZero())

	cancel()
	for range events {
	}
}

func TestVPIOFactoryEdgeNone(t *testing.T) {
	t.Parallel()

	f := NewVPIOFactory()
	in, err := f.OpenInput("chip0", 11, EdgeNone, BiasDefault, 0)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	events, err := in.Events(ctx)
	require.NoError(t, err)

	f.InjectEdge("chip0", 11, EdgeRising, true)

	select {
	case <-events:
		t.Fatal("expected no events for EdgeNone")
	default:
	}

	cancel()
	for range events {
	}
}
