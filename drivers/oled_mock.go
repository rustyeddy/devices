package drivers

import "image"

// MockOLEDFactory is a portable OLEDFactory for dev/CI/examples where no hardware exists.
// It returns a no-op OLED that accepts Draw() calls.
type MockOLEDFactory struct{}

func (MockOLEDFactory) OpenSSD1306(bus string, addr uint16, width, height int) (OLED, error) {
	return &mockOLED{}, nil
}

type mockOLED struct{}

func (*mockOLED) Draw(r image.Rectangle, src image.Image, sp image.Point) error {
	// no-op
	return nil
}

func (*mockOLED) Close() error {
	// no-op
	return nil
}

var _ OLEDFactory = MockOLEDFactory{}
var _ OLED = (*mockOLED)(nil)
