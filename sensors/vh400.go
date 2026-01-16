package sensors

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/rustyeddy/devices"
	"github.com/rustyeddy/devices/drivers"
)

// VH400Config configures a VH400 volumetric water content sensor.
//
// The VH400 itself outputs an analog voltage; this device expects that
// voltage to be read via an ADC (e.g. ADS1115).
type VH400Config struct {
	Name string

	// ADC is an already-open analog reader (optional).
	// If nil, Factory will be used to open an ADS1115.
	ADC drivers.ADC

	// Factory opens ADC devices (optional).
	Factory drivers.ADCFactory

	// Bus and Addr are used when opening ADS1115 via Factory.
	Bus  string
	Addr uint16

	// Channel is the ADS1115 single-ended input (0-3).
	Channel int

	// Interval is the polling cadence.
	Interval time.Duration

	// EmitInitial reads once immediately on start.
	EmitInitial bool

	// Buf sizes the out channel. Default 16.
	Buf int

	// NewTicker optionally overrides ticker creation (tests).
	NewTicker func(d time.Duration) Ticker
}

// VH400 reads VWC (volumetric water content) from an analog sensor.
//
// Output units are percent VWC.
type VH400 struct {
	devices.Base
	out chan float64

	cfg VH400Config
	adc drivers.ADC
}

// NewVH400 constructs a VH400 sensor.
func NewVH400(cfg VH400Config) *VH400 {
	if cfg.Buf <= 0 {
		cfg.Buf = 16
	}
	return &VH400{
		Base: devices.NewBase(cfg.Name, cfg.Buf),
		out:  make(chan float64, cfg.Buf),
		cfg:  cfg,
	}
}

// Out returns the VWC sample stream.
func (v *VH400) Out() <-chan float64 { return v.out }

// Descriptor returns sensor metadata.
func (v *VH400) Descriptor() devices.Descriptor {
	min := 0.0
	max := 100.0
	attrs := map[string]string{
		"adc":     "ads1115",
		"bus":     v.cfg.Bus,
		"addr":    fmt.Sprintf("0x%02x", v.cfg.Addr),
		"channel": strconv.Itoa(v.cfg.Channel),
	}
	return devices.Descriptor{
		Name:       v.Name(),
		Kind:       "vh400",
		ValueType:  "float64",
		Access:     devices.ReadOnly,
		Unit:       "%",
		Min:        &min,
		Max:        &max,
		Tags:       []string{"soil", "moisture", "analog"},
		Attributes: attrs,
	}
}

// Run polls the ADC, converts voltage -> VWC, and publishes samples.
func (v *VH400) Run(ctx context.Context) error {
	if v.cfg.Interval <= 0 {
		err := errors.New("vh400: interval must be > 0")
		v.Emit(devices.EventError, "invalid interval", err, nil)
		close(v.out)
		v.CloseEvents()
		return err
	}
	if v.cfg.Channel < 0 || v.cfg.Channel > 3 {
		err := fmt.Errorf("vh400: invalid channel %d", v.cfg.Channel)
		v.Emit(devices.EventError, "invalid channel", err, nil)
		close(v.out)
		v.CloseEvents()
		return err
	}

	adc := v.cfg.ADC
	if adc == nil {
		if v.cfg.Factory == nil {
			err := errors.New("vh400: adc and factory are both nil")
			v.Emit(devices.EventError, "factory missing", err, nil)
			close(v.out)
			v.CloseEvents()
			return err
		}
		opened, err := v.cfg.Factory.OpenADS1115(v.cfg.Bus, v.cfg.Addr)
		if err != nil {
			v.Emit(devices.EventError, "open adc failed", err, nil)
			close(v.out)
			v.CloseEvents()
			return err
		}
		adc = opened
		v.adc = opened
	} else {
		v.adc = adc
	}

	defer func() {
		if v.adc != nil {
			_ = v.adc.Close()
		}
	}()

	read := func(ctx context.Context) (float64, error) {
		volts, err := v.adc.ReadVolts(ctx, v.cfg.Channel)
		if err != nil {
			return 0, err
		}
		return vwcFromVolts(volts), nil
	}

	return RunPoller[float64](ctx, &v.Base, v.out, PollConfig[float64]{
		Interval:       v.cfg.Interval,
		EmitInitial:    v.cfg.EmitInitial,
		DropOnFull:     true,
		Read:           read,
		NewTicker:      v.cfg.NewTicker,
		SampleEventMsg: "sample",
		SampleMeta: func(vwc float64) map[string]string {
			return map[string]string{"vwc": fmt.Sprintf("%.2f", vwc)}
		},
	})
}

// vwcFromVolts converts VH400 output voltage to volumetric water content (percent).
//
// The piecewise linear approximations are based on Vegetronix's published curve:
// https://vegetronix.com/Products/VH400/VH400-Piecewise-Curve
func vwcFromVolts(volts float64) float64 {
	// Each segment is of the form: VWC = m*V - b
	var m, b float64

	switch {
	case volts >= 0.0 && volts <= 1.1:
		m = 10.0
		b = 1.0
	case volts > 1.1 && volts <= 1.3:
		m = 25.0
		b = 17.5
	case volts > 1.3 && volts <= 1.82:
		m = 48.08
		b = 47.5
	case volts > 1.82 && volts <= 2.2:
		m = 26.32
		b = 7.80
	case volts > 2.2 && volts <= 3.0:
		m = 62.5
		b = 7.89
	default:
		// out of spec; return 0 to avoid propagating NaNs.
		return 0.0
	}

	vwc := m*volts - b
	if vwc < 0 {
		return 0
	}
	if vwc > 100 {
		return 100
	}
	return vwc
}
