// Package bme280 provides a driver for the BME280 temperature, humidity,
// and pressure sensor using I2C communication.
package bme280

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/maciej/bme280"
	"github.com/rustyeddy/devices"
	"github.com/rustyeddy/devices/drivers"
)

// Response returns values read from the sensor containing all three
// values for temperature, humidity and pressure
//     type Response bme280.Response

type Env bme280.Response

func (e *Env) JSON() ([]byte, error) {
	jbuf, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	return jbuf, err
}

// BME280 represents an I2C temperature, humidity and pressure sensor.
// It defaults to address 0x77 and implements the device.Device interface.
type BME280 struct {
	*devices.DeviceBase[Env]
	bus    string
	addr   int
	driver *bme280.Driver
	isMock bool
	mu     sync.Mutex // protects Env fields from concurrent access
	Env
}

var (
	ErrInitFailed    = errors.New("failed to initialize BME280")
	ErrReadFailed    = errors.New("failed to read from BME280")
	ErrMarshalFailed = errors.New("failed to marshal BME280 data")
)

const (
	DefaultI2CBus     = "/dev/i2c-1"
	DefaultI2CAddress = 0x77

	minTemperature = -40.0
	maxTemperature = 85.0
	minHumidity    = 0.0
	maxHumidity    = 100.0
	minPressure    = 300.0
	maxPressure    = 1100.0
)

// BME280Config holds the configuration for the BME280 sensor
type BME280Config struct {
	Mode       bme280.Mode
	Filter     bme280.Filter
	Standby    bme280.StandByTime
	Oversample struct {
		Pressure    bme280.Oversampling
		Temperature bme280.Oversampling
		Humidity    bme280.Oversampling
	}
}

// defaultMockEnv returns the default mock environment values.
// This helper centralizes mock defaults for consistency and
// will support timeseries changes in the future.
func defaultMockEnv() Env {
	return Env{
		Temperature: 50.12,
		Pressure:    900.34,
		Humidity:    77.56,
	}
}

// DefaultConfig returns the default configuration
func DefaultConfig() BME280Config {
	return BME280Config{
		Mode:    bme280.ModeForced,
		Filter:  bme280.FilterOff,
		Standby: bme280.StandByTime1000ms,
		Oversample: struct {
			Pressure    bme280.Oversampling
			Temperature bme280.Oversampling
			Humidity    bme280.Oversampling
		}{
			Pressure:    bme280.Oversampling16x,
			Temperature: bme280.Oversampling16x,
			Humidity:    bme280.Oversampling16x,
		},
	}
}

// Create a new BME280 at the give bus and address. Defaults are
// typically /dev/i2c-1 address 0x99
// func New(id, bus string, addr int) (*BME280, error) {
func New(id, bus string, addr int) (b *BME280, err error) {
	b = &BME280{
		bus:        bus,
		addr:       addr,
		DeviceBase: devices.NewDeviceBase[Env](id),
		isMock:     devices.IsMock(),
		env:        Env{},
	}

	// initialize default mock values (will be used in Open/Get when isMock)
	if b.isMock {
		b.Env = defaultMockEnv()
	}

	return b, nil
}

func (b *BME280) Name() string {
	return b.DeviceBase.Name()
}

// Init opens the i2c bus at the specified address and gets the device
// ready for reading
func (b *BME280) Open() error {
	if b.isMock {
		return nil
	}

	i2c, err := drivers.GetI2CDriver(b.bus, b.addr)
	if err != nil {
		return err
	}
	b.driver = bme280.New(i2c)
	err = b.driver.InitWith(bme280.ModeForced, bme280.Settings{
		Filter:                  bme280.FilterOff,
		Standby:                 bme280.StandByTime1000ms,
		PressureOversampling:    bme280.Oversampling16x,
		TemperatureOversampling: bme280.Oversampling16x,
		HumidityOversampling:    bme280.Oversampling16x,
	})
	if err != nil {
		return err
	}
	return nil
}

func (b *BME280) Close() error {
	if b.isMock {
		return nil
	}
	return errors.New("TODO Need to implement bme280 close")
}

func (b *BME280) Set(v Env) error {
	if b.isMock {
		b.mu.Lock()
		b.Env = v
		b.mu.Unlock()
		return nil
	}
	return errors.New("BME280 is read-only")
}

// Read one Response from the sensor. If this device is being mocked
// we will make up some random floating point numbers between 0 and
// 100.
func (b *BME280) Get() (resp Env, err error) {
	if b.isMock {
		b.mu.Lock()
		// mutate stored values slightly to simulate readings
		b.Env.Temperature += 0.1
		b.Env.Humidity += 0.02
		b.Env.Pressure += 0.001
		resp = b.Env
		b.mu.Unlock()
		return resp, nil
	}

	val, err := b.driver.Read()
	if err != nil {
		return Env{}, err
	}
	resp.Temperature = val.Temperature
	resp.Humidity = val.Humidity
	resp.Pressure = val.Pressure

	inrange := func(val float64, min float64, max float64) error {
		if val < min || val > max {
			return fmt.Errorf("%7.2f out of range: min(%7.2f) max(%7.2f)", val, min, max)
		}
		return nil
	}

	if err = inrange(resp.Temperature, minTemperature, maxTemperature); err != nil {
		return resp, err
	}
	if err = inrange(resp.Humidity, minHumidity, maxHumidity); err != nil {
		return resp, err
	}
	if err = inrange(resp.Pressure, minPressure, maxPressure); err != nil {
		return resp, err
	}
	return resp, err
}

// BME280Mock provides a mocked BME280 for testing purposes
type BME280Mock struct {
	*devices.DeviceBase[Env]
	mu sync.Mutex // protects Env fields from concurrent access
	Env
}

func (b *BME280Mock) Open() error {
	b.mu.Lock()
	// Initialize with default mock values if not set
	if b.Env.Temperature == 0 && b.Env.Pressure == 0 && b.Env.Humidity == 0 {
		b.Env = defaultMockEnv()
	}
	b.mu.Unlock()
	return nil
}

func (b *BME280Mock) Close() error {
	return nil
}

func (b *BME280Mock) Get() (Env, error) {
	b.mu.Lock()
	// Return stored values (set via Set() or defaults from Open())
	b.Env.Temperature += 0.1
	b.Env.Humidity += 0.02
	b.Env.Pressure += 0.001
	// Return an immutable copy to avoid race conditions
	result := Env{
		Temperature: b.Env.Temperature,
		Humidity:    b.Env.Humidity,
		Pressure:    b.Env.Pressure,
	}
	b.mu.Unlock()
	return result, nil
}

func (b *BME280Mock) Set(v Env) error {
	b.mu.Lock()
	b.Env = v
	b.mu.Unlock()
	return nil
}

// ConvertCtoF converts Celsius to Fahrenheit
func ConvertCtoF(celsius float64) float64 {
	return (celsius * 9.0 / 5.0) + 32.0
}
