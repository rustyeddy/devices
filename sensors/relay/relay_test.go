package relay

import (
	"context"
	"testing"

	"github.com/rustyeddy/devices/drivers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRelayRunAppliesCommands(t *testing.T) {
	t.Parallel()

	f := drivers.NewVPIOFactory()
	relay := NewRelay(RelayConfig{
		Name:    "relay",
		Factory: f,
		Chip:    "chip0",
		Offset:  5,
		Initial: false,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- relay.Run(ctx)
	}()

	initial := <-relay.Out()
	assert.False(t, initial)

	relay.In() <- true
	updated := <-relay.Out()
	assert.True(t, updated)

	cancel()
	require.NoError(t, <-errCh)
}
