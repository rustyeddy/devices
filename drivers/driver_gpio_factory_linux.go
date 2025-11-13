//go:build linux

package drivers

/*
import "github.com/rustyeddy/devices"

// createGPIO creates a GPIO instance based on the mock setting.
// On Linux, returns GPIOCDev for real hardware unless mocking is enabled.
func createGPIO[T Value]() GPIO[T] {
	if devices.IsMock() {
		return NewVPIO[T]()
	}
	
	// For Linux hardware, we need to use GPIOCDev wrapped to match the interface
	// Since GPIOCDev works with int values, we'll use VPIO for non-int types
	var zero T
	switch any(zero).(type) {
	case int:
		// Return a wrapper that adapts GPIOCDev to GPIO[int]
		return any(newGPIOCDevAdapter()).(GPIO[T])
	default:
		// For bool and float64, use VPIO
		return NewVPIO[T]()
	}
}

// GPIOCDevAdapter adapts GPIOCDev to implement GPIO[int]
type GPIOCDevAdapter struct {
	*GPIOCDev
}

func newGPIOCDevAdapter() GPIO[int] {
	return &GPIOCDevAdapter{
		GPIOCDev: NewGPIOCDev("gpiochip0"),
	}
}

func (g *GPIOCDevAdapter) Pin(name string, pin int, options PinOptions) (Pin[int], error) {
	p, err := g.GPIOCDev.Pin(name, pin, options)
	if err != nil {
		return nil, err
	}
	return &GPIOCDevPinAdapter{p}, nil
}

func (g *GPIOCDevAdapter) Get(pin int) (int, error) {
	return g.GPIOCDev.Get(pin)
}

func (g *GPIOCDevAdapter) Set(pin int, v int) error {
	return g.GPIOCDev.Set(pin, v)
}

// GPIOCDevPinAdapter adapts GPIOCDevPin to implement Pin[int]
type GPIOCDevPinAdapter struct {
	*GPIOCDevPin
}

func (p *GPIOCDevPinAdapter) Get() (int, error) {
	return p.GPIOCDevPin.Get()
}

func (p *GPIOCDevPinAdapter) Set(v int) error {
	return p.GPIOCDevPin.Set(v)
}

*/
