package drivers

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/rustyeddy/devices"
)

// GPIO is used to initialize the GPIO and pins on a raspberry pi
type GPIO struct {
	Chipname string              `json:"chipname"`
	pins     map[int]*DigitalPin `json:"pins"`
	Mock     bool                `json:"mock"`
}

var (
	gpio *GPIO
)

// GetGPIO returns the GPIO singleton for the Raspberry PI
func GetGPIO() *GPIO {
	if gpio != nil {
		return gpio
	}

	// Chipname: "gpiochip4", // raspberry pi-5
	gpio = &GPIO{
		Chipname: "gpiochip0", // raspberry pi zero
		Mock:     devices.IsMock(),
	}
	gpio.pins = make(map[int]*DigitalPin)
	for _, pin := range gpio.pins {
		if err := pin.Init(); err != nil {
			slog.Error("Error initializing pin ", "offset", pin.offset)
		}
	}
	return gpio
}

// Shutdown resets the GPIO line allowing use by another program
func (gpio *GPIO) Close() error {
	for _, p := range gpio.pins {
		p.Close()
	}
	gpio.pins = nil
	return nil
}

// String returns the string representation of the GPIO
func (gpio *GPIO) String() string {
	str := ""
	for _, pin := range gpio.pins {
		str += pin.String()
	}
	return str
}

// JSON returns the JSON representation of the GPIO
func (gpio *GPIO) JSON() (j []byte, err error) {
	j, err = json.Marshal(gpio)
	if err != nil {
		return nil, fmt.Errorf("Error marshalling GPIO %s", err)
	}
	return j, nil
}

// Line interface is used to emulate a GPIO pin
type Line interface {
	Close() error
	Offset() int
	SetValue(int) error
	Reconfigure(...interface{}) error
	Value() (int, error)
}

type DigitalPin struct {
	name string
	Line

	On  func() error
	Off func() error

	offset int
	val    int
	mock   bool
	
	// Event handling - platform specific implementation
	EvtQ chan interface{}
}

func (p *DigitalPin) PinName() string {
	return p.name
}

func (p *DigitalPin) ID() string {
	return p.name
}

// String returns a string representation of the GPIO pin
func (pin *DigitalPin) String() string {
	v, err := pin.Value()
	if err != nil {
		slog.Error("Failed getting the value of ", "pin", pin.offset, "error", err)
	}
	str := fmt.Sprintf("%d: %d\n", pin.offset, v)
	return str
}

// Get returns the value of the pin, an error is returned if
// the GPIO value fails. Note: you can Get() the value of an
// input pin so no direction checks are done
func (pin *DigitalPin) Get() (int, error) {
	if pin.Line == nil {
		return 0, fmt.Errorf("GPIO not active")
	}
	val, err := pin.Line.Value()
	if err != nil {
		slog.Error("Failed to read PIN", "name", pin.name)
	}
	return val, err
}

// Set the value of the pin. Note: you can NOT set the value
// of an input pin, so we will check it and return an error.
// This maybe worthy of making it a panic!
func (pin *DigitalPin) Set(v int) error {
	if pin.Line == nil {
		return fmt.Errorf("GPIO not active")
	}
	pin.val = v
	return pin.Line.SetValue(v)
}

func (pin *DigitalPin) ON() error {
	return pin.Set(1)
}

// Off sets the value of the pin to 0
func (pin *DigitalPin) OFF() error {
	return pin.Set(0)
}

// Toggle with flip the value of the pin from 1 to 0 or 0 to 1
func (pin *DigitalPin) Toggle() error {
	val, err := pin.Get()
	if err != nil {
		return err
	}

	if val == 0 {
		val = 1
	} else {
		val = 0
	}
	return pin.Set(val)
}

func (d *DigitalPin) Close() error {
	return d.Line.Close()
}
