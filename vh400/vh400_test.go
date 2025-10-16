package vh400

import (
	"testing"

	"github.com/rustyeddy/devices"
)

func TestVH400(t *testing.T) {
	devices.Mock(true)

	v := New("vh400", 1)
	if v.Name() != "vh400" {
		t.Errorf("vh400 name expected (%s) got (%s)", "vh400", v.Name())
	}

	val, err := v.Read()
	if err != nil {
		t.Errorf("VH400 Read failed %s", err)
	}

	if val == 0.0 {
		t.Errorf("Expected value but got 0")
	}
}

func TestVH400_Name_ReturnsCorrectName(t *testing.T) {
	devices.Mock(true)
	v := New("vh400test", 2)
	if v == nil {
		t.Fatalf("New returned nil")
	}
	if v.Name() != "vh400test" {
		t.Errorf("Name() expected (%s) got (%s)", "vh400test", v.Name())
	}
}

func TestVH400_Read_ReturnsVWC(t *testing.T) {
	devices.Mock(true)
	v := New("vh400", 3)
	if v == nil {
		t.Fatalf("New returned nil")
	}
	val, err := v.Read()
	if err != nil {
		t.Errorf("Read() returned error: %v", err)
	}
	if val == 0.0 {
		t.Errorf("Read() returned zero VWC, expected non-zero value")
	}
}

func TestVH400_New_NonMockReturnsNilOnError(t *testing.T) {
	devices.Mock(false)
	v := New("vh400fail", 99)
	if v != nil {
		t.Errorf("Expected nil when ADS1115.Pin fails, got non-nil")
	}
	devices.Mock(true)
}

func Test_vwc_CurveSegments(t *testing.T) {
	tests := []struct {
		volts   float64
		wantMin float64
		wantMax float64
	}{
		{0.5, 4.0, 6.0},     // 0.0 <= volts <= 1.1
		{1.2, 12.0, 13.0},   // 1.1 < volts <= 1.3
		{1.5, 24.0, 25.0},   // 1.3 < volts <= 1.82
		{2.0, 44.0, 45.0},   // 1.82 < volts <= 2.2
		{2.5, 147.0, 149.0}, // 2.2 < volts <= 3.1
	}
	for _, tt := range tests {
		got := vwc(tt.volts)
		if got < tt.wantMin || got > tt.wantMax {
			t.Errorf("vwc(%v) = %v, want between %v and %v", tt.volts, got, tt.wantMin, tt.wantMax)
		}
	}
}

func TestVWCInvalidVoltage(t *testing.T) {
	got := vwc(3.5)
	if got != 0.0 {
		t.Errorf("vwc(3.5) expected 0.0, got %v", got)
	}
	got = vwc(-0.5)
	if got != 0.0 {
		t.Errorf("vwc(-0.5) expected 0.0, got %v", got)
	}
}
