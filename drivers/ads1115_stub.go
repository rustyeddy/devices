//go:build !linux || (!arm && !arm64)

package drivers

import (
	"errors"
)

// PeriphADCFactory is a stub on non-Linux ARM builds.
//
// This keeps the public API building on developer machines and in CI.
type PeriphADCFactory struct{}

func (PeriphADCFactory) OpenADS1115(bus string, addr uint16) (ADC, error) {
	return nil, errors.New("ads1115: supported only on linux/arm or linux/arm64")
}
