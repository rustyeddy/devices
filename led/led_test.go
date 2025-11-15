package led

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLEDInitializesFields(t *testing.T) {
	led, err := New("testled", 12)
	assert.NoError(t, err)
	assert.NotNil(t, led)
	assert.Equal(t, "testled", led.ID())
	assert.NotNil(t, led.Pin)
}
