package vh400

// See https://vegetronix.com/Products/VH400/VH400-Piecewise-Curve
// For calculations on the VWC.  Borrowed from above website

import (
	"log/slog"

	"github.com/rustyeddy/devices"
	"github.com/rustyeddy/devices/drivers"
)

type VH400 struct {
	drivers.Pin[float64]
	*devices.DeviceBase[float64]
}

func New(name string, index int) (*VH400, error) {
	gpio := drivers.GetGPIO[float64]()
	p, err := gpio.SetPin(name, index, drivers.PinOutput)
	if err != nil {
		return nil, err
	}
	v := &VH400{
		DeviceBase: devices.NewDeviceBase[float64](name),
		Pin:        p,
	}
	return v, nil
}

func (v *VH400) Get() (float64, error) {
	volts, err := v.Pin.Get()
	if err != nil {
		return volts, err
	}
	vwc := vwc(volts)
	return vwc, nil

}

func (v *VH400) Type() devices.Type {
	return devices.TypeFloat
}

/*
Most curves can be approximated with linear segments of the form:

y= m*x-b,

where m is the slope of the line

The VH400's Voltage to VWC curve can be approximated with 4 segments
of the form:

VWC= m*V-b

where V is voltage.

m = (VWC2 - VWC1) / (V2-V1)

where V1 and V2 are voltages recorded at the respective VWC levels of
VWC1 and VWC2. After m is determined, the y-axis intercept coefficient
b can be found by inserting one of the end points into the equation:

b= m*v-VWC
*/
func vwc(volts float64) float64 {
	var coef float64
	var rem float64

	switch {
	case volts >= 0.0 && volts <= 1.1:
		coef = 10.0
		rem = 1.0

	case volts > 1.1 && volts <= 1.3:
		coef = 25.0
		rem = 17.5

	case volts > 1.3 && volts <= 1.82:
		coef = 48.08
		rem = 47.5

	case volts > 1.82 && volts <= 2.2:
		coef = 26.32
		rem = 7.80

	case volts > 2.2 && volts <= 3.0:
		coef = 62.5
		rem = 7.89

	default:
		slog.Warn("VH400 Invalid voltage: out of range 0.0 -> 3.0", "voltage", volts)
		return 0.0
	}
	vwc := coef*volts - rem
	return vwc
}
