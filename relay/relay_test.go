package relay

import (
	"testing"

	"github.com/rustyeddy/devices"
	"github.com/stretchr/testify/require"
)

func init() {
	devices.SetMock(true)
}

func TestNewRelayInitializesFields(t *testing.T) {
	relay, err := New("testrelay", 6)
	require.NoError(t, err)
	require.NotNil(t, relay, "New() returned nil")
	require.NotNil(t, relay.DeviceBase, "Device field not initialized")
	require.NotNil(t, relay.Pin, "DigitalPin field not initialized")
	require.Equal(t, "testrelay", relay.ID())
}
