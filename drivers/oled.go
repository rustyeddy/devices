package drivers

import "image"

// OLED is a minimal drawing surface for an SSD1306-style display.
//
// Implementations must be safe to call Draw from a single goroutine.
type OLED interface {
	Draw(r image.Rectangle, src image.Image, sp image.Point) error
	Close() error
}

// OLEDFactory opens OLED drivers.
type OLEDFactory interface {
	// OpenSSD1306 opens an SSD1306 on the given I2C bus/address.
	// Width/Height are pixel dimensions (e.g. 128x64).
	OpenSSD1306(bus string, addr uint16, width, height int) (OLED, error)
}
