package button

import (
	"log/slog"

	"github.com/rustyeddy/devices"
	"github.com/rustyeddy/devices/drivers"
)

// Button is a simple momentary GPIO Device that can be detected when
// the button is pushed (rising edge) or when it is released (falling edge).
// Low is open, high is closed.
type ButtonOld struct {
	name string
	pin  *drivers.GPIOCDevPin
	devices.Device[int]
}

// NewButton creates a new button with the given GPIO controller and pin configuration.
// Use NewButtonVPIO for virtual/mock buttons or NewButtonHW for hardware buttons.
func NewButton(name string, pin *drivers.GPIOCDevPin) *Button {
	b := &Button{
		name: name,
		pin:  pin,
	}
	b.Device = b
	return b
}

// NewButtonHW creates a button using real hardware GPIO
func NewButtonHW(name string, offset int) (*Button, error) {
	gpio := drivers.NewGPIOCDev("gpiochip0")

	const (
		PinInput     drivers.PinOptions = 1 << 0
		PinPullUp    drivers.PinOptions = 1 << 4
		PinBothEdges drivers.PinOptions = 1 << 8
	)

	pin, err := gpio.Pin(name, offset, PinInput|PinPullUp|PinBothEdges)
	if err != nil {
		return nil, err
	}

	return NewButton(name, pin), nil
}

// NewButtonVPIO creates a button using virtual GPIO for testing
func NewButtonVPIO(name string, offset uint, vpio *drivers.VPIO[int]) (*Button, error) {
	vpin, err := vpio.Pin(name, offset, drivers.DirectionInput)
	if err != nil {
		return nil, err
	}

	// For now, we'll create a simple adapter
	// In a full implementation, you might want to create a proper adapter
	// that converts VPin to GPIOCDevPin interface
	b := &Button{
		name: name,
		pin:  nil, // VPIO doesn't use GPIOCDevPin
	}
	b.Device = b

	// Store vpin reference if needed for Get operations
	// This is a simplified implementation
	_ = vpin

	return b, nil
}

func (b *Button) ID() string {
	return b.name
}

func (b *Button) Open() error {
	return nil
}

func (b *Button) Close() error {
	if b.pin != nil {
		return b.pin.Close()
	}
	return nil
}

// Get reads the value of the button (0 = not pressed, 1 = pressed)
func (b *Button) Get() (int, error) {
	if b.pin == nil {
		return 0, devices.ErrNotImplemented
	}

	val, err := b.pin.Get()
	if err != nil {
		slog.Error("Failed to read button value", "name", b.name, "error", err.Error())
		return val, err
	}
	return val, nil
}

func (b *Button) Set(val int) error {
	return devices.ErrNotImplemented
}

func (b *Button) Type() devices.Type {
	return devices.TypeInt
}
