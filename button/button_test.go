package button

import (
	"testing"

	"github.com/rustyeddy/devices"
	"github.com/rustyeddy/devices/drivers"
	"github.com/stretchr/testify/assert"
)

func init() {
	devices.SetMock(true)
}

func TestNewButtonDefaultOptions(t *testing.T) {
	btn, err := New("testbtn", 5, drivers.PinPullDown)
	assert.NoError(t, err)
	assert.NotNil(t, btn)
	assert.NotNil(t, btn.Pin)
	assert.Equal(t, "testbtn", btn.DeviceBase.Name())
}

func TestNewButtonCustomOptions(t *testing.T) {
	// opt := gpiocdev.WithPullDown
	btn, err := New("custombtn", 7, drivers.PinPullDown)
	assert.NoError(t, err)
	assert.NotNil(t, btn)
	assert.NotNil(t, btn.Pin)
}
