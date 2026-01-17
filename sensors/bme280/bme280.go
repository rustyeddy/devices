package bme280

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rustyeddy/devices"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/devices/v3/bmxx80"
	"periph.io/x/host/v3"
)

// Env is a single sample from the BME280.
// Units:
//   - Temperature: °C
//   - Pressure: Pa
//   - Humidity: %RH
type Env struct {
	Temperature float64 `json:"temperature"`
	Pressure    float64 `json:"pressure"`
	Humidity    float64 `json:"humidity"`
}

// Sensor is the small interface we need from periph's bmxx80.Dev.
type Sensor interface {
	Sense(e *physic.Env) error
	Halt() error
}

const (
	DefaultBus  = ""   // i2creg.Open("") picks the "first available" bus
	DefaultAddr = 0x76 // common default for BME280/BMP280
)

// Config configures a BME280 device.
type Config struct {
	Name string

	// I2C bus name for i2creg.Open(). Default "" means first available bus.
	Bus string

	// I2C device address (0x76 or 0x77)
	Addr uint16

	// Poll interval must be > 0.
	Interval time.Duration

	// Poller behavior
	EmitInitial bool
	DropOnFull  bool

	// Optional periph init hook (defaults to host.Init()).
	InitHost func() error

	// Optional overrides for tests
	OpenBus func(bus string) (i2c.BusCloser, error)
	NewDev  func(bus i2c.Bus, addr uint16) (Sensor, error)

	// Optional bmxx80 opts. If nil, uses bmxx80.DefaultOpts.
	Opts *bmxx80.Opts
}

// BME280 is a channels-based BME280 sensor device.
type BME280 struct {
	devices.Base
	cfg Config
	out chan Env
}

// New constructs a new BME280 device.
func New(cfg Config) *BME280 {
	if cfg.Bus == "" {
		cfg.Bus = DefaultBus
	}
	if cfg.Addr == 0 {
		cfg.Addr = DefaultAddr
	}
	return &BME280{
		Base: devices.NewBase(cfg.Name, 16),
		cfg:  cfg,
		out:  make(chan Env, 16),
	}
}

// Out returns the sample stream.
func (b *BME280) Out() <-chan Env { return b.out }

// Descriptor returns metadata for discovery/introspection.
func (b *BME280) Descriptor() devices.Descriptor {
	return devices.Descriptor{
		Name:      b.Name(),
		Kind:      "bme280",
		ValueType: "env",
		Access:    devices.ReadOnly,
		Tags:      []string{"i2c", "sensor", "environment"},
		Attributes: map[string]string{
			"bus":  b.cfg.Bus,
			"addr": fmt.Sprintf("0x%02x", b.cfg.Addr),
		},
	}
}

// Run opens the I2C bus + device and starts the polling loop.
// NOTE: RunPoller owns closing b.out and closing Base events.
func (b *BME280) Run(ctx context.Context) error {
	// Defaults
	initHost := b.cfg.InitHost
	if initHost == nil {
		initHost = func() error {
			_, err := host.Init()
			return err
		}
	}
	openBus := b.cfg.OpenBus
	if openBus == nil {
		openBus = func(name string) (i2c.BusCloser, error) { return i2creg.Open(name) }
	}
	newDev := b.cfg.NewDev
	if newDev == nil {
		newDev = func(bus i2c.Bus, addr uint16) (Sensor, error) {
			opts := b.cfg.Opts
			if opts == nil {
				opts = &bmxx80.DefaultOpts
			}
			return bmxx80.NewI2C(bus, addr, opts)
		}
	}

	if b.cfg.Interval <= 0 {
		err := errors.New("bme280 poll interval must be > 0")
		b.Emit(devices.EventError, "invalid interval", err, nil)
		return err
	}

	if err := initHost(); err != nil {
		b.Emit(devices.EventError, "host init failed", err, nil)
		return err
	}

	bus, err := openBus(b.cfg.Bus)
	if err != nil {
		b.Emit(devices.EventError, "open i2c failed", err, nil)
		return err
	}
	// Close resources BEFORE RunPoller (it will close out/events)
	defer func() { _ = bus.Close() }()

	dev, err := newDev(bus, b.cfg.Addr)
	if err != nil {
		b.Emit(devices.EventError, "init bme280 failed", err, nil)
		return err
	}
	defer func() { _ = dev.Halt() }()

	read := func(ctx context.Context) (Env, error) {
		var e physic.Env
		if err := dev.Sense(&e); err != nil {
			return Env{}, err
		}
		return Env{
			// Per periph conventions, convert via physic base units:
			// Temperature stored as physic.Temperature in Kelvin units.
			Temperature: float64(e.Temperature)/float64(physic.Kelvin) - 273.15,
			Pressure:    float64(e.Pressure) / float64(physic.Pascal),
			// RelativeHumidity stored as physic.RelativeHumidity in MicroRH units.
			Humidity: float64(e.Humidity) / float64(physic.MicroRH) / 10000.0,
		}, nil
	}

	cfg := devices.PollConfig[Env]{
		Interval:       b.cfg.Interval,
		EmitInitial:    b.cfg.EmitInitial,
		DropOnFull:     b.cfg.DropOnFull,
		Read:           read,
		SampleEventMsg: "sample",
		SampleMeta: func(v Env) map[string]string {
			return map[string]string{
				"temp_c":   fmt.Sprintf("%.2f", v.Temperature),
				"pressure": fmt.Sprintf("%.0f", v.Pressure),
				"humidity": fmt.Sprintf("%.2f", v.Humidity),
				"addr":     fmt.Sprintf("0x%02x", b.cfg.Addr),
				"bus":      b.cfg.Bus,
			}
		},
	}

	return devices.RunPoller(ctx, &b.Base, b.out, cfg)
}
