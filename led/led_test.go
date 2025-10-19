package led

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLEDInitializesFields(t *testing.T) {
	led := New("testled", 12)
	assert.NotNil(t, led)
	assert.Equal(t, "testled", led.ID())
	assert.Nil(t, led.DigitalPin)

	err := led.Open()
	assert.NoError(t, err)
	assert.NotNil(t, led.DigitalPin)
}

func TestLEDCallbackOn(t *testing.T) {
	led := New("led_on", 13)
	called := false
	led.On = func() {
		called = true
	}
	led.Off = func() {}
	led.Callback(true)
	assert.True(t, called, "On should be called when Callback(true)")
}

func TestLEDCallbackOff(t *testing.T) {
	led := New("led_off", 14)

	called := false
	led.On = func() {}
	led.Off = func() {
		called = true
	}
	led.Callback(false)
	assert.True(t, called, "Off should be called when Callback(false)")
}

func TestLEDCallbackNoPanic(t *testing.T) {
	led := New("led_nopanic", 15)
	led.On = func() {}
	led.Off = func() {}
	assert.NotPanics(t, func() { led.Callback(true) })
	assert.NotPanics(t, func() { led.Callback(false) })
}
