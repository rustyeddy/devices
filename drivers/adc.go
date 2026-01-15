package drivers

import "context"

// ADC reads analog voltages from one or more channels.
//
// ReadVolts returns the current channel voltage in volts (V).
//
// Implementations are expected to be used by a single goroutine. If you need
// concurrent reads, wrap the ADC with a mutex at a higher level.
type ADC interface {
	ReadVolts(ctx context.Context, channel int) (float64, error)
	Close() error
}

// ADCFactory opens ADC devices.
//
// This mirrors the GPIO Factory pattern used elsewhere in the repo.
// For now, we only standardize ADS1115.
type ADCFactory interface {
	OpenADS1115(bus string, addr uint16) (ADC, error)
}
