//go:build linux && (arm || arm64)

package drivers

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/devices/v3/ads1x15"
	"periph.io/x/host/v3"
)

// ADS1115 implements ADC using the periph.io ADS1x15 driver.
//
// It exposes single-ended reads on channels 0-3.
//
// Voltage reference is fixed to 3.3V by default (typical Raspberry Pi use).
type ADS1115 struct {
	mu   sync.Mutex
	bus  i2c.BusCloser
	dev  *ads1x15.Dev
	pins [4]ads1x15.PinADC
}

// ADS1115Opts controls how the ADS1115 is configured.
type ADS1115Opts struct {
	// Bus is the I2C bus name passed to i2creg.Open.
	// Common values: "" (default), "1".
	Bus string
	// Addr is the 7-bit I2C address (typically 0x48).
	Addr uint16
	// RefMV is the reference in millivolts used for scaling (default 3300).
	RefMV int64
	// SampleRate controls the sample frequency (default 1Hz).
	SampleRate physic.Frequency
	// Mode controls power mode (default SaveEnergy).
	Mode ads1x15.Mode
}

func (o *ADS1115Opts) withDefaults() {
	if o.RefMV == 0 {
		o.RefMV = 3300
	}
	if o.SampleRate == 0 {
		o.SampleRate = 1 * physic.Hertz
	}
	if o.Mode == "" {
		o.Mode = ads1x15.SaveEnergy
	}
}

// NewADS1115 opens the ADS1115.
func NewADS1115(opts ADS1115Opts) (*ADS1115, error) {
	opts.withDefaults()

	if _, err := host.Init(); err != nil {
		return nil, fmt.Errorf("ads1115: host init: %w", err)
	}

	bus, err := i2creg.Open(opts.Bus)
	if err != nil {
		return nil, fmt.Errorf("ads1115: open i2c bus %q: %w", opts.Bus, err)
	}

	dev, err := ads1x15.NewADS1115(bus, &ads1x15.DefaultOpts)
	if err != nil {
		_ = bus.Close()
		return nil, fmt.Errorf("ads1115: create device: %w", err)
	}

	a := &ADS1115{bus: bus, dev: dev}

	for ch := 0; ch < 4; ch++ {
		pin, err := a.pinForChannel(ch, opts)
		if err != nil {
			_ = a.Close()
			return nil, err
		}
		a.pins[ch] = pin
	}

	return a, nil
}

func (a *ADS1115) pinForChannel(ch int, opts ADS1115Opts) (ads1x15.PinADC, error) {
	var c ads1x15.Channel
	switch ch {
	case 0:
		c = ads1x15.Channel0
	case 1:
		c = ads1x15.Channel1
	case 2:
		c = ads1x15.Channel2
	case 3:
		c = ads1x15.Channel3
	default:
		return ads1x15.PinADC{}, fmt.Errorf("ads1115: invalid channel %d", ch)
	}

	return a.dev.PinForChannel(c, opts.RefMV*physic.MilliVolt, opts.SampleRate, opts.Mode)
}

// ReadVolts reads a single sample from the given channel.
func (a *ADS1115) ReadVolts(ctx context.Context, channel int) (float64, error) {
	if channel < 0 || channel > 3 {
		return 0, fmt.Errorf("ads1115: invalid channel %d", channel)
	}
	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// Read returns an analog.Sample where V is in physic.ElectricPotential (nanovolts).
	s, err := a.pins[channel].Read()
	if err != nil {
		return 0, err
	}
	return float64(s.V) / float64(physic.Volt), nil
}

func (a *ADS1115) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	var first error
	for i := 0; i < 4; i++ {
		if err := a.pins[i].Halt(); err != nil && first == nil {
			first = err
		}
	}
	if a.bus != nil {
		if err := a.bus.Close(); err != nil && first == nil {
			first = err
		}
		a.bus = nil
	}
	return first
}

// PeriphADCFactory opens periph-backed ADC devices.
type PeriphADCFactory struct{}

func (PeriphADCFactory) OpenADS1115(bus string, addr uint16) (ADC, error) {
	// addr currently unused by periph's constructor (it uses default 0x48),
	// but we keep it in the interface for future extension.
	if addr != 0 && addr != 0x48 {
		return nil, errors.New("ads1115: custom address not supported by this implementation yet")
	}
	return NewADS1115(ADS1115Opts{Bus: bus, Addr: addr})
}

var _ ADC = (*ADS1115)(nil)
