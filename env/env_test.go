package env

import (
	"testing"

	"github.com/maciej/bme280"
	"github.com/rustyeddy/devices"
	device "github.com/rustyeddy/devices"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test constants
const (
	TestI2CBus     = "/dev/i2c-1"
	TestI2CAddress = 0x77
)

func init() {
	devices.SetMock(true)
}

func TestBME280Creation(t *testing.T) {
	tests := []struct {
		name    string
		devName string
		bus     string
		addr    int
		wantErr bool
	}{
		{
			name:    "valid configuration",
			devName: "bme-test",
			bus:     "/dev/i2c-fake",
			addr:    0x76,
			wantErr: false,
		},
		{
			name:    "default configuration",
			devName: "bme-default",
			bus:     TestI2CBus,
			addr:    TestI2CAddress,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bme := New(tt.devName, tt.bus, tt.addr)
			if bme == nil {
				t.Fatal("Failed to create BME280 device")
			}

			if bme.ID() != tt.devName {
				t.Errorf("Name() = %v, want %v", bme.ID(), tt.devName)
			}

			err := bme.Open()
			if tt.wantErr && err == nil {
				t.Error("Open() expected error but got none")
			} else if !tt.wantErr && err != nil {
				t.Errorf("Open() unexpected error = %v", err)
			}
		})
	}
}

// func TestBME280ValidReadings(t *testing.T) {
// 	device.Mock(true)
// 	defer device.Mock(false)

// 	bme := New("bme-test", "/dev/i2c-fake", 0x76)
// 	bme.MockReader = func() (*bme280.Response, error) {
// 		return &bme280.Response{
// 			Temperature: 50.33,
// 			Humidity:    74.33,
// 			Pressure:    1000.33,
// 		}, nil
// 	}

// 	resp, err := bme.Read()
// 	assert.NoError(t, err)
// 	assert.NotEqual(t, &bme280.Response{}, resp)
// 	assert.Equal(t, resp.Temperature, 50.33)
// 	assert.Equal(t, resp.Humidity, 74.33)
// 	assert.Equal(t, resp.Pressure, 1000.33)
// }

func TestBME280Reading(t *testing.T) {
	bme := New("bme-test", "/dev/i2c-fake", 0x76)
	if err := bme.Open(); err != nil {
		t.Fatalf("Open() failed: %v", err)
	}

	tests := []struct {
		name           string
		minTemperature float64
		maxTemperature float64
		minHumidity    float64
		maxHumidity    float64
		minPressure    float64
		maxPressure    float64
	}{
		{
			name:           "valid ranges",
			minTemperature: -40.0,
			maxTemperature: 85.0,
			minHumidity:    0.0,
			maxHumidity:    100.0,
			minPressure:    300.0,
			maxPressure:    1100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := bme.Get()
			assert.NoError(t, err)
			assert.NotEmpty(t, resp)

			if resp.Temperature < tt.minTemperature || resp.Temperature > tt.maxTemperature {
				t.Errorf("Temperature %f outside valid range [%f, %f]",
					resp.Temperature, tt.minTemperature, tt.maxTemperature)
			}

			if resp.Humidity < tt.minHumidity || resp.Humidity > tt.maxHumidity {
				t.Errorf("Humidity %f outside valid range [%f, %f]",
					resp.Humidity, tt.minHumidity, tt.maxHumidity)
			}

			if resp.Pressure < tt.minPressure || resp.Pressure > tt.maxPressure {
				t.Errorf("Pressure %f outside valid range [%f, %f]",
					resp.Pressure, tt.minPressure, tt.maxPressure)
			}
		})
	}
}

func TestBME280ReadPub(t *testing.T) {
	bme := New("bme-test", "/dev/i2c-fake", 0x76)
	if err := bme.Open(); err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
}

func TestBME280JSON(t *testing.T) {
	bme := New("bme-test", "/dev/i2c-fake", 0x76)
	if err := bme.Open(); err != nil {
		t.Fatalf("Open() failed: %v", err)
	}

	// Test JSON method if it exists
	// data, err := bme.JSON()
	// if err != nil {
	// 	t.Errorf("JSON() error = %v", err)
	// }

	// if len(data) == 0 {
	// 	t.Error("JSON() returned empty data")
	// }

	// // Verify it's valid JSON
	// var result map[string]interface{}
	// if err := json.Unmarshal(data, &result); err != nil {
	// 	t.Errorf("JSON() returned invalid JSON: %v", err)
	// }
}

func TestBME280String(t *testing.T) {
	bme := New("bme-test", "/dev/i2c-fake", 0x76)
	str := bme.String()
	require.NotEmpty(t, str)
	expected := bme.ID() + " [0]"
	assert.Equal(t, expected, str)
}

func TestBME280MockReader(t *testing.T) {
	bme := New("bme-test", "/dev/i2c-fake", 0x76)
	err := bme.Open()
	assert.NoError(t, err)

	// Test the mock reader directly
	resp, err := bme.MockRead()
	assert.NoError(t, err)
	assert.NotEmpty(t, resp)

	// Verify expected mock values
	expectedTemp := 50.12
	expectedPressure := 900.34
	expectedHumidity := 77.56

	assert.Equal(t, resp.Temperature, expectedTemp)
	assert.Equal(t, resp.Pressure, expectedPressure)
	assert.Equal(t, resp.Humidity, expectedHumidity)
}

func TestBME280ConvertCtoF(t *testing.T) {
	tests := []struct {
		name    string
		celsius float64
		want    float64
	}{
		{"freezing point", 0.0, 32.0},
		{"boiling point", 100.0, 212.0},
		{"negative temperature", -40.0, -40.0},
		{"room temperature", 25.0, 77.0},
		{"body temperature", 37.0, 98.6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ConvertCtoF(tt.celsius)
			if got != tt.want {
				t.Errorf("ConvertCtoF(%f) = %f, want %f", tt.celsius, got, tt.want)
			}
		})
	}
}

func TestBME280DefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Mode != bme280.ModeForced {
		t.Errorf("DefaultConfig() Mode = %v, want %v", config.Mode, bme280.ModeForced)
	}

	if config.Filter != bme280.FilterOff {
		t.Errorf("DefaultConfig() Filter = %v, want %v", config.Filter, bme280.FilterOff)
	}

	if config.Standby != bme280.StandByTime1000ms {
		t.Errorf("DefaultConfig() Standby = %v, want %v", config.Standby, bme280.StandByTime1000ms)
	}

	if config.Oversample.Pressure != bme280.Oversampling16x {
		t.Errorf("DefaultConfig() Pressure oversampling = %v, want %v",
			config.Oversample.Pressure, bme280.Oversampling16x)
	}

	if config.Oversample.Temperature != bme280.Oversampling16x {
		t.Errorf("DefaultConfig() Temperature oversampling = %v, want %v",
			config.Oversample.Temperature, bme280.Oversampling16x)
	}

	if config.Oversample.Humidity != bme280.Oversampling16x {
		t.Errorf("DefaultConfig() Humidity oversampling = %v, want %v",
			config.Oversample.Humidity, bme280.Oversampling16x)
	}
}

func TestBME280CustomMockReader(t *testing.T) {
	bme := New("bme-test", "/dev/i2c-fake", 0x76)

	// Set a custom mock reader
	customTemp := 50.12
	customPressure := 900.34
	customHumidity := 77.56

	oldreader := bme.MockReader
	defer func() { bme.MockReader = oldreader }()
	bme.MockReader = func() (*Env, error) {
		return &Env{
			Temperature: customTemp,
			Pressure:    customPressure,
			Humidity:    customHumidity,
		}, nil
	}

	if err := bme.Open(); err != nil {
		t.Fatalf("Open() failed: %v", err)
	}

	resp, err := bme.Get()
	require.NoError(t, err)
	require.NotEqual(t, &Env{}, resp)
	assert.Equal(t, resp.Temperature, customTemp)
	assert.Equal(t, resp.Pressure, customPressure)
	assert.Equal(t, resp.Humidity, customHumidity)
}

func TestBME280Constants(t *testing.T) {
	tests := []struct {
		name     string
		constant interface{}
		expected interface{}
	}{
		{"DefaultI2CBus", DefaultI2CBus, "/dev/i2c-1"},
		{"DefaultI2CAddress", DefaultI2CAddress, 0x77},
		{"minTemperature", minTemperature, -40.0},
		{"maxTemperature", maxTemperature, 85.0},
		{"minHumidity", minHumidity, 0.0},
		{"maxHumidity", maxHumidity, 100.0},
		{"minPressure", minPressure, 300.0},
		{"maxPressure", maxPressure, 1100.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("Constant %s = %v, want %v", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestBME280ErrorTypes(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"ErrInitFailed", ErrInitFailed, "failed to initialize BME280"},
		{"ErrReadFailed", ErrReadFailed, "failed to read from BME280"},
		{"ErrMarshalFailed", ErrMarshalFailed, "failed to marshal BME280 data"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.want {
				t.Errorf("Error %s = %v, want %v", tt.name, tt.err.Error(), tt.want)
			}
		})
	}
}

func TestBME280NonMockMode(t *testing.T) {
	// Test with mock disabled - this should test the actual I2C path
	device.SetMock(false)
	defer device.SetMock(true) // Reset to mock mode

	bme := New("bme-test", "/dev/i2c-nonexistent", 0x76)

	// This should fail since we're using a nonexistent bus
	err := bme.Open()
	if err == nil {
		t.Error("Open() expected error with nonexistent I2C bus, got nil")
	}
}

func TestBME280ResponseValidation(t *testing.T) {
	bme := New("bme-test", "/dev/i2c-fake", 0x76)

	// Test with out-of-range values
	oldreader := bme.MockReader
	defer func() { bme.MockReader = oldreader }()
	bme.MockReader = func() (*Env, error) {
		return &Env{
			Temperature: -39.0, // Below minTemperature
			Pressure:    200.0, // Below minPressure
			Humidity:    110.0, // Above maxHumidity
		}, nil
	}
	_, err := bme.Get()
	require.Error(t, err)
}

func TestBME280FieldInitialization(t *testing.T) {
	bme := New("test-device", "/dev/i2c-2", 0x76)
	assert.Nil(t, bme.MockReader, "MockReader should be nil before Open() in mock mode")

	err := bme.Open()
	require.NoError(t, err)

	if !device.IsMock() {
		assert.NotNil(t, bme.DeviceBase, "Device field not initialized")
		assert.NotNil(t, bme.driver)
	}

	assert.Equal(t, "test-device", bme.ID(), "Device name = %s, want test-device", bme.ID())
	assert.Equal(t, "/dev/i2c-2", bme.bus)
	assert.Equal(t, 0x76, bme.addr)
}
