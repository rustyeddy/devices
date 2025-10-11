package relay

import (
	"testing"

	"github.com/rustyeddy/devices"
)

func TestRelay(t *testing.T) {
	devices.Mock(true)

	relay := New("relay", 5)
	if relay.Name != "relay" {
		t.Errorf("relay expected Name (%s) got (%s)", "relay", relay.Name)
	}

	relay.Callback(true)
	v, err := relay.Value()
	if err != nil {
		t.Fatalf("relay.Value() got error %v", err)
	}
	if v != 1 {
		t.Errorf("relay expected (1) got (%d)", v)
	}

	relay.Callback(false)
	v, err = relay.Value()
	if err != nil {
		t.Fatalf("relay.Value() got error %v", err)
	}
	if v != 0 {
		t.Errorf("relay expected (0) got (%d)", v)
	}
}
