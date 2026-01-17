//go:build linux && (arm || arm64)

package drivers

import (
	"fmt"
	"image"

	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/devices/v3/ssd1306"
	"periph.io/x/host/v3"
)

// PeriphOLEDFactory opens SSD1306 OLED displays using periph.io.
type PeriphOLEDFactory struct{}

func (PeriphOLEDFactory) OpenSSD1306(bus string, addr uint16, width, height int) (OLED, error) {
	if _, err := host.Init(); err != nil {
		return nil, fmt.Errorf("ssd1306: host init: %w", err)
	}

	b, err := i2creg.Open(bus)
	if err != nil {
		return nil, fmt.Errorf("ssd1306: open i2c bus %q: %w", bus, err)
	}

	opts := &ssd1306.DefaultOpts
	opts.W = width
	opts.H = height

	// NOTE: periph's ssd1306.NewI2C does not expose an address override in all
	// versions. If your panel is not at the default address, use a custom bus
	// wrapper or update to a periph version that supports it.
	_ = addr

	dev, err := ssd1306.NewI2C(b, opts)
	if err != nil {
		_ = b.Close()
		return nil, fmt.Errorf("ssd1306: new i2c: %w", err)
	}

	return &periphOLED{bus: b, dev: dev}, nil
}

type periphOLED struct {
	bus interface{ Close() error }
	dev *ssd1306.Dev
}

func (p *periphOLED) Draw(r image.Rectangle, src image.Image, sp image.Point) error {
	return p.dev.Draw(r, src, sp)
}

func (p *periphOLED) Close() error {
	var err1 error
	if p.bus != nil {
		err1 = p.bus.Close()
		p.bus = nil
	}
	return err1
}

var _ OLEDFactory = PeriphOLEDFactory{}
