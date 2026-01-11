package gpio

import (
	"context"
	"testing"
	"time"

	"github.com/rustyeddy/devices/drivers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewButtonDefaults(t *testing.T) {
	t.Parallel()

	btn := NewButton(ButtonConfig{Name: "btn"})
	assert.Equal(t, "btn", btn.Name())
	assert.Equal(t, drivers.EdgeBoth, btn.cfg.Edge)
	assert.Equal(t, drivers.BiasPullUp, btn.cfg.Bias)
	assert.Equal(t, 30*time.Millisecond, btn.cfg.Debounce)
}

func TestButtonRunEmitsState(t *testing.T) {
	t.Parallel()

	f := drivers.NewVPIOFactory()
	btn := NewButton(ButtonConfig{
		Name:     "btn",
		Factory:  f,
		Chip:     "chip0",
		Offset:   1,
		Edge:     drivers.EdgeBoth,
		Bias:     drivers.BiasPullUp,
		Debounce: time.Nanosecond,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- btn.Run(ctx)
	}()

	initial := <-btn.Out()
	assert.False(t, initial)

	f.InjectEdge("chip0", 1, drivers.EdgeRising, true)
	after := <-btn.Out()
	assert.True(t, after)

	cancel()
	require.NoError(t, <-errCh)
}
