package led

import (
	"context"
	"testing"

	"github.com/rustyeddy/devices/drivers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLEDRunAppliesCommands(t *testing.T) {
	t.Parallel()

	f := drivers.NewVPIOFactory()
	led := New(LEDConfig{
		Name:    "led",
		Factory: f,
		Chip:    "chip0",
		Offset:  12,
		Initial: false,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- led.Run(ctx)
	}()

	initial := <-led.Out()
	assert.False(t, initial)

	led.In() <- true
	updated := <-led.Out()
	assert.True(t, updated)

	cancel()
	require.NoError(t, <-errCh)
}
