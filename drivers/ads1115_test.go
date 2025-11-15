//go:build linux

package drivers

import (
	"fmt"
	"testing"
	"time"

	"log"

	"github.com/rustyeddy/devices"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/devices/v3/ads1x15"
	host "periph.io/x/host/v3"
)

func init() {
	devices.SetMock(true)
}

func TestGetADS1115Singleton(t *testing.T) {
	// Reset singleton for testing
	ads1115 = nil

	tests := []struct {
		name string
	}{
		{
			name: "first call creates instance",
		},
		{
			name: "subsequent calls return same instance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adc1 := GetADS1115()
			require.NotNil(t, adc1)

			adc2 := GetADS1115()
			require.NotNil(t, adc2)

			// Should be same instance (singleton pattern)
			assert.Equal(t, adc1, adc2)
		})
	}
}

func TestNewADS1115(t *testing.T) {
	tests := []struct {
		name    string
		devName string
		bus     string
		addr    int
		wantErr bool
	}{
		{
			name:    "valid configuration",
			devName: "ads1115-test",
			bus:     "/dev/i2c-1",
			addr:    0x48,
			wantErr: false,
		},
		{
			name:    "alternative address",
			devName: "ads1115-alt",
			bus:     "/dev/i2c-1",
			addr:    0x49,
			wantErr: false,
		},
		{
			name:    "different bus",
			devName: "ads1115-bus2",
			bus:     "/dev/i2c-2",
			addr:    0x48,
			wantErr: false,
		},
		{
			name:    "empty name",
			devName: "",
			bus:     "/dev/i2c-1",
			addr:    0x48,
			wantErr: false, // Should handle gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adc := NewADS1115(tt.devName, tt.bus, tt.addr)

			if tt.wantErr {
				assert.Nil(t, adc)
			} else {
				assert.NotNil(t, adc)
				assert.True(t, adc.mock, "Device should be in mock mode")
			}
		})
	}
}

func TestADS1115_Open(t *testing.T) {
	tests := []struct {
		name    string
		mock    bool
		wantErr bool
	}{
		{
			name:    "open in mock mode",
			mock:    true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			devices.SetMock(tt.mock)
			defer devices.SetMock(true)

			adc := &ADS1115{mock: tt.mock}
			err := func() (err error) {
				if _, err := host.Init(); err != nil {
					log.Printf("device_ads1115: host init failed: %s", err)
					return err
				}
				adc.bus, err = i2creg.Open("")
				if err != nil {
					log.Printf("device_ads1115: i2c open failed: %s", err)
					return err
				}
				adc.adc, err = ads1x15.NewADS1115(adc.bus, &ads1x15.DefaultOpts)
				if err != nil {
					log.Printf("device_ads1115: new ads failed: %s", err)
					return err
				}
				return nil
			}()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestADS1115_SetPin(t *testing.T) {
	tests := []struct {
		name    string
		pinName string
		channel int
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid channel 0",
			pinName: "pin0",
			channel: 0,
			wantErr: false,
		},
		{
			name:    "valid channel 1",
			pinName: "pin1",
			channel: 1,
			wantErr: false,
		},
		{
			name:    "valid channel 2",
			pinName: "pin2",
			channel: 2,
			wantErr: false,
		},
		{
			name:    "valid channel 3",
			pinName: "pin3",
			channel: 3,
			wantErr: false,
		},
		{
			name:    "invalid channel negative",
			pinName: "pin-neg",
			channel: -1,
			wantErr: true,
			errMsg:  "PinInit Invalid channel",
		},
		{
			name:    "invalid channel too high",
			pinName: "pin-high",
			channel: 4,
			wantErr: true,
			errMsg:  "PinInit Invalid channel",
		},
		{
			name:    "invalid channel much too high",
			pinName: "pin-way-high",
			channel: 10,
			wantErr: true,
			errMsg:  "PinInit Invalid channel",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adc := NewADS1115("test-adc", "/dev/i2c-1", 0x48)
			adc.mock = true
			require.NotNil(t, adc)

			pin, err := adc.SetPin(tt.pinName, tt.channel)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				assert.Nil(t, pin)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, pin)
				assert.Equal(t, tt.pinName, pin.name)
				assert.Equal(t, tt.channel, pin.index)
			}
		})
	}
}

func TestADS1115_GetSet(t *testing.T) {
	tests := []struct {
		name      string
		pinIndex  int
		setValue  float64
		expectErr bool
	}{
		{
			name:      "get from valid pin",
			pinIndex:  0,
			expectErr: false,
		},
		{
			name:      "set on read-only pin",
			pinIndex:  0,
			setValue:  3.3,
			expectErr: true, // ADC pins are read-only
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adc := NewADS1115("test-adc", "/dev/i2c-1", 0x48)
			require.NotNil(t, adc)

			// Set up a pin first
			pin, err := adc.SetPin("test-pin", tt.pinIndex)
			require.NoError(t, err)
			require.NotNil(t, pin)

			adc.pins[tt.pinIndex] = pin

			if tt.setValue > 0 {
				err = adc.Set(tt.pinIndex, tt.setValue)
				if tt.expectErr {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), "can not be set")
				} else {
					assert.NoError(t, err)
				}
			} else {
				val, err := adc.Get(tt.pinIndex)
				if tt.expectErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.GreaterOrEqual(t, val, 0.0)
					assert.LessOrEqual(t, val, 5.0) // Reasonable voltage range
				}
			}
		})
	}
}

func TestADS1115Pin_Operations(t *testing.T) {
	tests := []struct {
		name    string
		pinName string
		channel int
	}{
		{
			name:    "pin operations channel 0",
			pinName: "test-pin-0",
			channel: 0,
		},
		{
			name:    "pin operations channel 3",
			pinName: "test-pin-3",
			channel: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adc := NewADS1115("test-adc", "/dev/i2c-1", 0x48)
			require.NotNil(t, adc)

			pin, err := adc.SetPin(tt.pinName, tt.channel)
			require.NoError(t, err)
			require.NotNil(t, pin)

			// Test pin properties
			assert.Equal(t, tt.pinName, pin.ID())
			assert.Equal(t, tt.channel, pin.Index())
			assert.NotEmpty(t, pin.String())

			// Test Open/Close
			err = pin.Open()
			assert.NoError(t, err)

			err = pin.Close()
			assert.NoError(t, err)
		})
	}
}

func TestADS1115Pin_Get(t *testing.T) {
	tests := []struct {
		name    string
		pinName string
		channel int
		wantErr bool
	}{
		{
			name:    "successful read channel 0",
			pinName: "voltage-pin-0",
			channel: 0,
			wantErr: false,
		},
		{
			name:    "successful read channel 1",
			pinName: "voltage-pin-1",
			channel: 1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adc := NewADS1115("test-adc", "/dev/i2c-1", 0x48)
			require.NotNil(t, adc)

			pin, err := adc.SetPin(tt.pinName, tt.channel)
			require.NoError(t, err)
			require.NotNil(t, pin)

			val, err := pin.Get()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, val, 0.0)
				assert.LessOrEqual(t, val, 5.0) // Reasonable voltage range for ADC
				t.Logf("Pin %s reading: %.3fV", tt.pinName, val)
			}
		})
	}
}

func TestADS1115Pin_Set(t *testing.T) {
	tests := []struct {
		name    string
		pinName string
		channel int
		value   float64
	}{
		{
			name:    "attempt set on read-only pin",
			pinName: "readonly-pin",
			channel: 0,
			value:   3.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adc := NewADS1115("test-adc", "/dev/i2c-1", 0x48)
			require.NotNil(t, adc)

			pin, err := adc.SetPin(tt.pinName, tt.channel)
			require.NoError(t, err)
			require.NotNil(t, pin)

			err = pin.Set(tt.value)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "can not be set")
		})
	}
}

func TestADS1115Pin_ReadContinuous(t *testing.T) {
	tests := []struct {
		name      string
		pinName   string
		channel   int
		readCount int
		timeout   time.Duration
	}{
		{
			name:      "continuous reading channel 0",
			pinName:   "continuous-pin-0",
			channel:   0,
			readCount: 3,
			timeout:   time.Second * 2,
		},
		{
			name:      "continuous reading channel 2",
			pinName:   "continuous-pin-2",
			channel:   2,
			readCount: 5,
			timeout:   time.Second * 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adc := NewADS1115("test-adc", "/dev/i2c-1", 0x48)
			require.NotNil(t, adc)

			pin, err := adc.SetPin(tt.pinName, tt.channel)
			require.NoError(t, err)
			require.NotNil(t, pin)

			// Start continuous reading
			readChan := pin.ReadContinuous()
			require.NotNil(t, readChan)

			// Read specified number of values with timeout
			readings := make([]float64, 0, tt.readCount)
			timeout := time.After(tt.timeout)

			for len(readings) < tt.readCount {
				select {
				case val := <-readChan:
					readings = append(readings, val)
					assert.GreaterOrEqual(t, val, 0.0)
					assert.LessOrEqual(t, val, 5.0)
					t.Logf("Continuous reading %d: %.3fV", len(readings), val)

				case <-timeout:
					t.Fatalf("Timeout waiting for continuous readings after %v", tt.timeout)
				}
			}

			assert.Len(t, readings, tt.readCount)
		})
	}
}

func TestADS1115_Close(t *testing.T) {
	tests := []struct {
		name      string
		setupPins int
	}{
		{
			name:      "close with no pins",
			setupPins: 0,
		},
		{
			name:      "close with one pin",
			setupPins: 1,
		},
		{
			name:      "close with all pins",
			setupPins: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adc := NewADS1115("test-adc", "/dev/i2c-1", 0x48)
			require.NotNil(t, adc)

			// Set up specified number of pins
			for i := 0; i < tt.setupPins; i++ {
				pin, err := adc.SetPin(fmt.Sprintf("pin-%d", i), i)
				require.NoError(t, err)
				adc.pins[i] = pin
			}

			// Close should not error
			err := adc.Close()
			assert.NoError(t, err)
		})
	}
}

func TestSample2Volts(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected float64
	}{
		{
			name:     "zero volts",
			input:    0,
			expected: 0.0,
		},
		{
			name:     "3.3 volts",
			input:    3300000000, // 3.3V in nanovolts
			expected: 3.3,
		},
		{
			name:     "5.0 volts",
			input:    5000000000, // 5.0V in nanovolts
			expected: 5.0,
		},
		{
			name:     "1.65 volts",
			input:    1650000000, // 1.65V in nanovolts
			expected: 1.65,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sample2Volts(tt.input)
			assert.InDelta(t, tt.expected, result, 0.001,
				"Expected %.3fV, got %.3fV", tt.expected, result)
		})
	}
}

func TestADS1115MockBehavior(t *testing.T) {
	// Ensure we're in mock mode
	devices.SetMock(true)

	tests := []struct {
		name string
	}{
		{
			name: "mock device creation and basic operations",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adc := NewADS1115("mock-adc", "/dev/i2c-1", 0x48)
			require.NotNil(t, adc)
			assert.True(t, adc.mock)

			// In mock mode, we should be able to create pins without hardware
			pin, err := adc.SetPin("mock-pin", 0)
			require.NoError(t, err)
			require.NotNil(t, pin)

			// Mock pin should provide reasonable values
			val, err := pin.Get()
			require.NoError(t, err)
			assert.GreaterOrEqual(t, val, 0.0)
			assert.LessOrEqual(t, val, 5.0)
		})
	}
}

func BenchmarkADS1115Pin_Get(b *testing.B) {
	adc := NewADS1115("bench-adc", "/dev/i2c-1", 0x48)
	pin, err := adc.SetPin("bench-pin", 0)
	if err != nil {
		b.Fatalf("Failed to create pin: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := pin.Get()
		if err != nil {
			b.Fatalf("Failed to read pin: %v", err)
		}
	}
}

func BenchmarkSample2Volts(b *testing.B) {
	testVal := int64(3300000000) // 3.3V in nanovolts

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sample2Volts(testVal)
	}
}
