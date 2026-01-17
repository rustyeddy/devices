package bme280

import (
	"context"
	"testing"
	"time"

	"github.com/rustyeddy/devices"
	"github.com/stretchr/testify/require"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/physic"
)

type fakeBus struct{}

func (b *fakeBus) String() string                    { return "fake-i2c" }
func (b *fakeBus) Tx(addr uint16, w, r []byte) error { return nil }
func (b *fakeBus) SetSpeed(f physic.Frequency) error { return nil }
func (b *fakeBus) Close() error                      { return nil }

// fakeSensor lets us control what Sense() returns.
type fakeSensor struct {
	sense func(e *physic.Env) error
	halt  func() error
}

func (s *fakeSensor) Sense(e *physic.Env) error {
	if s.sense != nil {
		return s.sense(e)
	}
	return nil
}
func (s *fakeSensor) Halt() error {
	if s.halt != nil {
		return s.halt()
	}
	return nil
}

func TestBME280_Run_EmitsSamples(t *testing.T) {
	t.Parallel()

	// Set up: fake sensor returns deterministic env.
	fs := &fakeSensor{
		sense: func(e *physic.Env) error {
			// 25°C = 298.15K
			e.Temperature = physic.Temperature(298.15 * float64(physic.Kelvin))
			e.Pressure = physic.Pressure(101325 * float64(physic.Pascal))
			// 50%RH = 0.5 RH = 500000 microRH? (microRH is 1e-6 RH)
			e.Humidity = physic.RelativeHumidity(0.5 * 1e6 * float64(physic.MicroRH))
			return nil
		},
	}

	cfg := Config{
		Name:        "bme",
		Bus:         "",
		Addr:        0x76,
		Interval:    10 * time.Millisecond,
		EmitInitial: true,
		DropOnFull:  true,

		// avoid real host/i2c
		InitHost: func() error { return nil },
		OpenBus:  func(string) (i2c.BusCloser, error) { return &fakeBus{}, nil },
		NewDev:   func(bus i2c.Bus, addr uint16) (Sensor, error) { return fs, nil },
	}

	dev := New(cfg)
	require.Equal(t, "bme", dev.Name())

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- dev.Run(ctx)
	}()

	// EmitInitial should publish immediately
	got := <-dev.Out()
	require.InEpsilon(t, 25.0, got.Temperature, 0.5)
	require.InEpsilon(t, 101325.0, got.Pressure, 1.0)
	require.InEpsilon(t, 50.0, got.Humidity, 1.0)

	cancel()
	require.NoError(t, <-errCh)

	_, ok := <-dev.Out()
	require.False(t, ok, "out should be closed on stop")
}

func TestDescriptor(t *testing.T) {
	t.Parallel()
	d := New(Config{Name: "bme", Interval: time.Second})
	desc := d.Descriptor()
	require.Equal(t, "bme280", desc.Kind)
	require.Equal(t, devices.ReadOnly, desc.Access)
}
