package vh400

import (
	"testing"

	"github.com/rustyeddy/devices"
	"github.com/stretchr/testify/assert"
)

func init() {
	devices.SetMock(true)
}

func TestVH400(t *testing.T) {
	v := New("vh400", 1)
	assert.Equal(t, "vh400", v.ID())

	err := v.Open()
	assert.NoError(t, err, "VH400 failed to open")

	val, err := v.Get()
	assert.NoError(t, err)
	assert.NotEqual(t, val, 0.0)
}

func TestVH400_Name_ReturnsCorrectName(t *testing.T) {
	v := New("vh400test", 2)
	if v == nil {
		t.Fatalf("New returned nil")
	}
	if v.ID() != "vh400test" {
		t.Errorf("Name() expected (%s) got (%s)", "vh400test", v.ID())
	}
}

func TestVH400_Read_ReturnsVWC(t *testing.T) {
	v := New("vh400", 3)
	assert.NotNil(t, v)

	err := v.Open()
	assert.NoError(t, err)

	val, err := v.Get()
	assert.NoError(t, err)
	assert.NotEqual(t, val, 0.0)
}

func TestVH400_New_NonMockReturnsNilOnError(t *testing.T) {
	devices.SetMock(false)
	v := New("vh400fail", 99)
	err := v.Open()
	assert.Error(t, err)
	devices.SetMock(true)
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
