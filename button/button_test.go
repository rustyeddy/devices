package button

import (
	"testing"

	"github.com/rustyeddy/devices"
	"github.com/stretchr/testify/assert"
	"github.com/warthog618/go-gpiocdev"
)

func init() {
	devices.SetMock(true)
}
func TestNewButton_DefaultOptions(t *testing.T) {
	btn := New("testbtn", 5)

	assert.NotNil(t, btn)
	assert.NotNil(t, btn.DigitalPin)
	assert.Equal(t, "testbtn", btn.ID())
}

func TestNewButton_CustomOptions(t *testing.T) {
	opt := gpiocdev.WithPullDown
	btn := New("custombtn", 7, opt)
	assert.NotNil(t, btn)
	assert.NotNil(t, btn.DigitalPin)
}
