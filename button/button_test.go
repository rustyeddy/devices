package button

import (
	"testing"

	"github.com/rustyeddy/devices"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/warthog618/go-gpiocdev"
)

func init() {
	devices.SetMock(true)
}
func TestNewButton_DefaultOptions(t *testing.T) {
	dev := New("testbtn", 5)
	btn, ok := dev.(*Button)
	require.True(t, ok)

	assert.NotNil(t, btn)
	assert.NotNil(t, btn.DigitalPin)
	assert.Equal(t, "testbtn", btn.ID())
}

func TestNewButton_CustomOptions(t *testing.T) {
	opt := gpiocdev.WithPullDown
	dev := New("custombtn", 7, opt)
	btn, ok := dev.(*Button)
	require.True(t, ok)

	assert.NotNil(t, btn)
	assert.NotNil(t, btn.DigitalPin)
}
