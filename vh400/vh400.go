package vh400

// See https://vegetronix.com/Products/VH400/VH400-Piecewise-Curve
// For calculations on the VWC.  Borrowed from above website

import (
	"fmt"
	"log"

	"github.com/rustyeddy/devices"
	"github.com/rustyeddy/devices/drivers"
)

type VH400 struct {
	id  string
	pin int
	drivers.AnalogPin
	devices.Device[*VH400]
}

func New(id string, pin int) *VH400 {
	v := &VH400{
		id:  id,
		pin: pin,
	}
	return v
}

func (v *VH400) Open() error {
	if devices.IsMock() {
		v.AnalogPin = drivers.NewMockAnalogPin(v.id, v.pin, nil)
		return nil
	}

	ads := drivers.GetADS1115()
	p, err := ads.Pin(v.id, v.pin, nil)
	if err != nil {
		return fmt.Errorf("VH400 ERROR opening %s pin: %d - error: %s",
			v.id, v.pin, err)
	}
	v.AnalogPin = p
	return nil
}

func (v *VH400) ID() string {
	return v.id
}

func (v *VH400) Get() (float64, error) {
	volts, err := v.AnalogPin.Read()
	if err != nil {
		return volts, err
	}
	vwc := vwc(volts)
	return vwc, nil

}

func (v *VH400) Set(val float64) error {
	return devices.ErrNotImplemented
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
		log.Printf("VH400 Invalid voltage: out of range 0.0 -> 3.0 %5.2f", volts)
		return 0.0
	}
	vwc := coef*volts - rem
	return vwc
}
