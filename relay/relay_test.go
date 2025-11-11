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
	relay := New("testrelay", 6)
	require.NotNil(t, relay, "New() returned nil")
	require.NotNil(t, relay.Device, "Device field not initialized")
	require.NotNil(t, relay.DigitalPin, "DigitalPin field not initialized")
	require.Equal(t, "testrelay", relay.ID())
}

func TestRelayCallbackOn(t *testing.T) {
	relay := New("relay_on", 7)
	relay.Callback(true)
	val, err := relay.Get()
	if err != nil {
		t.Fatalf("relay.Get() got error %v", err)
	}
	if val != 1 {
		t.Error("On should set value to 1 when Callback(true)")
	}
}

func TestRelay_Callback_Off(t *testing.T) {
	relay := New("relay_off", 8)
	relay.Callback(false)
	val, err := relay.Get()
	if err != nil {
		t.Fatalf("relay.Get() got error %v", err)
	}
	if val != 0 {
		t.Error("Off should set value to 0 when Callback(false)")
	}
}

func TestRelay_Callback_NoPanic(t *testing.T) {
	relay := New("relay_nopanic", 9)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Callback panicked: %v", r)
		}
	}()
	relay.Callback(true)
	relay.Callback(false)
}

func TestRelayCallbackDefaultBehavior(t *testing.T) {
	relay := New("relay_default", 10)
	// Use default On/Off methods from DigitalPin
	relay.Callback(true)
	v, err := relay.Get()
	if err != nil {
		t.Fatalf("relay.Get() got error %v", err)
	}
	if v != 1 {
		t.Errorf("relay expected (1) got (%d)", v)
	}
	relay.Callback(false)
	v, err = relay.Get()
	if err != nil {
		t.Fatalf("relay.Get() got error %v", err)
	}
	if v != 0 {
		t.Errorf("relay expected (0) got (%d)", v)
	}
}
