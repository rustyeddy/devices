// Package bme280 provides a driver for the BME280 temperature, humidity,
// and pressure sensor using I2C communication.
package bme280

import (
	"errors"
	"fmt"

	"github.com/maciej/bme280"
	"github.com/rustyeddy/devices"
	"github.com/rustyeddy/devices/drivers"
)

// BME280 represents an I2C temperature, humidity and pressure sensor.
// It defaults to address 0x77 and implements the device.Device interface.
type BME280 struct {
	id     string
	bus    string
	addr   int
	driver *bme280.Driver

	devices.Device[*bme280.Response]
	MockReader func() (*bme280.Response, error) // Function to mock reading data
}

type Env struct {
	Temperature string `json:"temperature"`
	Humidity    string `json:"humidity"`
	Pressure    string `json:"pressure"`
}

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

// Response returns values read from the sensor containing all three
// values for temperature, humidity and pressure
type Response bme280.Response

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

// Create a new BME280 at the give bus and address. Defaults are
// typically /dev/i2c-1 address 0x99
func New(id, bus string, addr int) *BME280 {
	b := &BME280{
		id:   id,
		bus:  bus,
		addr: addr,
	}
	b.Device = b
	return b
}

func (b *BME280) ID() string {
	return b.id
}

func (b *BME280) Type() devices.Type {
	return devices.TypeBME280
}

// Init opens the i2c bus at the specified address and gets the device
// ready for reading
func (b *BME280) Open() error {
	if devices.IsMock() {
		if b.MockReader == nil {
			b.MockReader = b.MockRead
		}
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
	return errors.New("TODO Need to implement bme280 close")
}

// Read one Response from the sensor. If this device is being mocked
// we will make up some random floating point numbers between 0 and
// 100.
func (b *BME280) Get() (resp *bme280.Response, err error) {
	if devices.IsMock() {
		resp, err = b.MockReader()
	} else {
		val, err := b.driver.Read()
		if err != nil {
			return nil, err
		}
		resp = &val

	}
	if err != nil {
		return resp, err
	}

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

func (b *BME280) Set(v *bme280.Response) error {
	return devices.ErrTypeNotImplemented
}

func (b *BME280) String() string {
	return b.ID()
}

func (b *BME280) MockRead() (*bme280.Response, error) {
	return &bme280.Response{
		Temperature: 50.12,
		Pressure:    900.34,
		Humidity:    77.56,
	}, nil
}

// ConvertCtoF converts Celsius to Fahrenheit
func ConvertCtoF(celsius float64) float64 {
	return (celsius * 9.0 / 5.0) + 32.0
}
