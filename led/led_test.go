package led

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLED_InitializesFields(t *testing.T) {
	led := New("testled", 12)
	assert.NotNil(t, led)
	assert.NotNil(t, led.Device)
	assert.Equal(t, "testled", led.Device.Name)
	assert.NotNil(t, led.DigitalPin)
}

func TestLEDCallbackOn(t *testing.T) {
	led := New("led_on", 13)
	called := false
	led.On = func() error {
		called = true
		return nil
	}
	led.Off = func() error { return nil }
	led.Callback(true)
	assert.True(t, called, "On should be called when Callback(true)")
}

func TestLEDCallbackOff(t *testing.T) {
	led := New("led_off", 14)
	called := false
	led.On = func() error { return nil }
	led.Off = func() error {
		called = true
		return nil
	}
	led.Callback(false)
	assert.True(t, called, "Off should be called when Callback(false)")
}

func TestLEDCallbackNoPanic(t *testing.T) {
	led := New("led_nopanic", 15)
	led.On = func() error { return nil }
	led.Off = func() error { return nil }
	assert.NotPanics(t, func() { led.Callback(true) })
	assert.NotPanics(t, func() { led.Callback(false) })
}
