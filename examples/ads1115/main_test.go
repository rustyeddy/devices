package main

import (
	"testing"

	"github.com/rustyeddy/devices"
	"github.com/rustyeddy/devices/bme280"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	devices.SetMock(true)
}

// TestableMain wraps main functionality for testing
func TestableMain() error {
	// Set the BME i2c device and address Initialize the bme to use
	// the i2c bus
	bme, err := bme280.New("bme280", "/dev/i2c-1", 0x76)
	if err != nil {
		return err
	}

	// Open the device to prepare it for usage
	err = bme.Open()
	if err != nil {
		return err
	}

	_, err = bme.Get()
	if err != nil {
		return err
	}

	return nil
}

func TestMain(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "successful execution",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := TestableMain()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBME280Integration(t *testing.T) {
	tests := []struct {
		name          string
		deviceName    string
		i2cBus        string
		i2cAddr       int
		expectOpenErr bool
		expectReadErr bool
		wantErr       bool
	}{
		{
			name:          "valid BME280 configuration",
			deviceName:    "bme280-test",
			i2cBus:        "/dev/i2c-1",
			i2cAddr:       0x76,
			expectOpenErr: false,
			expectReadErr: false,
			wantErr:       false,
		},
		{
			name:          "alternative I2C address",
			deviceName:    "bme280-alt",
			i2cBus:        "/dev/i2c-1",
			i2cAddr:       0x77,
			expectOpenErr: false,
			expectReadErr: false,
			wantErr:       false,
		},
		{
			name:          "different I2C bus",
			deviceName:    "bme280-bus2",
			i2cBus:        "/dev/i2c-2",
			i2cAddr:       0x76,
			expectOpenErr: false,
			expectReadErr: false,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bme, err := bme280.New(tt.deviceName, tt.i2cBus, tt.i2cAddr)
			if err != nil {
				t.Fatalf("Failed to create BME280 device: %v", err)
			}

			err = bme.Open()
			if tt.expectOpenErr && err == nil {
				t.Error("Open() expected error but got none")
			} else if !tt.expectOpenErr && err != nil {
				t.Errorf("Open() unexpected error = %v", err)
			}

			if !tt.expectOpenErr {
				val, err := bme.Get()
				if tt.expectReadErr && err == nil {
					t.Error("Get() expected error but got none")
				} else if !tt.expectReadErr && err != nil {
					t.Errorf("Get() unexpected error = %v", err)
				}

				if !tt.expectReadErr {
					assert.NotNil(t, val)
					t.Logf("BME280 reading: Temp=%.2f°C, Humidity=%.2f%%, Pressure=%.2fhPa",
						val.Temperature, val.Humidity, val.Pressure)
				}
			}
		})
	}
}

func TestBME280ErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		deviceName  string
		i2cBus      string
		i2cAddr     int
		shouldPanic bool
	}{
		{
			name:        "invalid device name - empty string",
			deviceName:  "",
			i2cBus:      "/dev/i2c-1",
			i2cAddr:     0x76,
			shouldPanic: false, // bme280.New should handle this gracefully
		},
		{
			name:        "invalid I2C address - negative",
			deviceName:  "bme280-test",
			i2cBus:      "/dev/i2c-1",
			i2cAddr:     -1,
			shouldPanic: false,
		},
		{
			name:        "invalid I2C address - too high",
			deviceName:  "bme280-test",
			i2cBus:      "/dev/i2c-1",
			i2cAddr:     0x80,
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				assert.Panics(t, func() {
					bme, err := bme280.New(tt.deviceName, tt.i2cBus, tt.i2cAddr)
					if err != nil {
						panic(err)
					}
					err = bme.Open()
					if err != nil {
						panic(err)
					}
					_, err = bme.Get()
					if err != nil {
						panic(err)
					}
				})
			} else {
				assert.NotPanics(t, func() {
					bme, err := bme280.New(tt.deviceName, tt.i2cBus, tt.i2cAddr)
					if err != nil {
						t.Logf("Expected error creating device: %v", err)
						return
					}
					err = bme.Open()
					if err != nil {
						t.Logf("Expected error opening device: %v", err)
						return
					}
					_, err = bme.Get()
					if err != nil {
						t.Logf("Expected error reading device: %v", err)
					}
				})
			}
		})
	}
}

func TestBME280MockData(t *testing.T) {
	bme, err := bme280.New("bme280-mock", "/dev/i2c-1", 0x76)
	require.NoError(t, err)

	err = bme.Open()
	require.NoError(t, err)

	// Test multiple readings to ensure consistency
	for i := 0; i < 5; i++ {
		val, err := bme.Get()
		require.NoError(t, err)
		require.NotNil(t, val)

		// Validate readings are within expected ranges
		assert.GreaterOrEqual(t, val.Temperature, -40.0, "Temperature below minimum")
		assert.LessOrEqual(t, val.Temperature, 85.0, "Temperature above maximum")

		assert.GreaterOrEqual(t, val.Humidity, 0.0, "Humidity below minimum")
		assert.LessOrEqual(t, val.Humidity, 100.0, "Humidity above maximum")

		assert.GreaterOrEqual(t, val.Pressure, 300.0, "Pressure below minimum")
		assert.LessOrEqual(t, val.Pressure, 1100.0, "Pressure above maximum")

		t.Logf("Reading %d: Temp=%.2f°C, Humidity=%.2f%%, Pressure=%.2fhPa",
			i+1, val.Temperature, val.Humidity, val.Pressure)
	}
}

func TestBME280DeviceProperties(t *testing.T) {
	tests := []struct {
		name       string
		deviceName string
		i2cBus     string
		i2cAddr    int
	}{
		{
			name:       "standard configuration",
			deviceName: "bme280",
			i2cBus:     "/dev/i2c-1",
			i2cAddr:    0x76,
		},
		{
			name:       "alternative configuration",
			deviceName: "weather-sensor",
			i2cBus:     "/dev/i2c-2",
			i2cAddr:    0x77,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bme, err := bme280.New(tt.deviceName, tt.i2cBus, tt.i2cAddr)
			require.NoError(t, err)

			assert.Equal(t, tt.deviceName, bme.Name(), 0.10)
			assert.Contains(t, bme.String(), tt.deviceName)

			err = bme.Open()
			require.NoError(t, err)

			// Test that the device can be read multiple times
			val1, err := bme.Get()
			require.NoError(t, err)

			val2, err := bme.Get()
			require.NoError(t, err)

			// In mock mode, readings should be consistent
			assert.InDelta(t, val1.Temperature, val2.Temperature, 0.11)
			assert.InDelta(t, val1.Humidity, val2.Humidity, 0.10)
			assert.InDelta(t, val1.Pressure, val2.Pressure, 0.10)
		})
	}
}

func TestBME280NonMockMode(t *testing.T) {
	// Temporarily disable mock mode to test real I2C behavior
	devices.SetMock(false)
	defer devices.SetMock(true)

	bme, err := bme280.New("bme280-real", "/dev/i2c-nonexistent", 0x76)
	require.NoError(t, err)

	// This should fail since we're using a nonexistent bus
	err = bme.Open()
	assert.Error(t, err, "Open() should fail with nonexistent I2C bus")
}

func BenchmarkBME280Reading(b *testing.B) {
	bme, err := bme280.New("bme280-bench", "/dev/i2c-1", 0x76)
	if err != nil {
		b.Fatalf("Failed to create BME280 device: %v", err)
	}

	err = bme.Open()
	if err != nil {
		b.Fatalf("Failed to open BME280 device: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := bme.Get()
		if err != nil {
			b.Fatalf("Failed to read BME280: %v", err)
		}
	}
}
