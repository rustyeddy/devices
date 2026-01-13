package mock

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSwitch_Run_EmitsInitialAndUpdates(t *testing.T) {
	t.Parallel()

	sw := NewSwitch(SwitchConfig{Name: "sw", Initial: false})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() { errCh <- sw.Run(ctx) }()

	// initial
	assert.False(t, <-sw.Out())

	// update
	sw.In() <- true
	assert.True(t, <-sw.Out())

	cancel()
	require.NoError(t, <-errCh)
}

func TestSwitch_ClosesOutOnStop(t *testing.T) {
	t.Parallel()

	sw := NewSwitch(SwitchConfig{Name: "sw", Initial: false})
	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() { errCh <- sw.Run(ctx) }()

	<-sw.Out() // drain initial
	cancel()

	require.NoError(t, <-errCh)

	_, ok := <-sw.Out()
	assert.False(t, ok, "out channel should be closed")
}
