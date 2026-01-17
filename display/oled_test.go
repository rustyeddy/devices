package display

import (
	"context"
	"image"
	"testing"
	"time"

	"github.com/rustyeddy/devices/drivers"
	"github.com/stretchr/testify/require"
	"periph.io/x/devices/v3/ssd1306/image1bit"
)

type mockOLED struct {
	draws  int
	last   image.Image
	closed bool
}

func (m *mockOLED) Draw(r image.Rectangle, src image.Image, sp image.Point) error {
	m.draws++
	m.last = src
	return nil
}

func (m *mockOLED) Close() error {
	m.closed = true
	return nil
}

type mockOLEDFactory struct{ dev *mockOLED }

func (f *mockOLEDFactory) OpenSSD1306(bus string, addr uint16, width, height int) (drivers.OLED, error) {
	if f.dev == nil {
		f.dev = &mockOLED{}
	}
	return f.dev, nil
}

func TestOLED_ClearAndFlush(t *testing.T) {
	t.Parallel()

	factory := &mockOLEDFactory{}
	o := NewOLED(OLEDConfig{
		Name:    "oled",
		Factory: factory,
		Width:   16,
		Height:  8,
	})

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- o.Run(ctx) }()

	// draw something
	o.In() <- OLEDCommand{Type: CmdRect, X0: 2, Y0: 2, X1: 6, Y1: 6, Pixel: PixelOn}
	o.In() <- OLEDCommand{Type: CmdClear}

	// verify cleared buffer
	for x := 0; x < o.Width; x++ {
		for y := 0; y < o.Height; y++ {
			require.Equal(t, image1bit.Bit(PixelOff), o.Background.BitAt(x, y))
		}
	}

	// flush should call Draw
	done := make(chan error, 1)
	o.In() <- OLEDCommand{Type: CmdFlush, Done: done}
	require.NoError(t, <-done)
	require.NotNil(t, factory.dev)
	require.Equal(t, 1, factory.dev.draws)
	require.NotNil(t, factory.dev.last)

	cancel()
	require.NoError(t, <-errCh)
}

func TestOLED_SetPixel_ClipsOutOfBounds(t *testing.T) {
	t.Parallel()
	factory := &mockOLEDFactory{}
	o := NewOLED(OLEDConfig{Name: "oled", Factory: factory, Width: 4, Height: 4})

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- o.Run(ctx) }()

	o.In() <- OLEDCommand{Type: CmdSetPixel, X: -1, Y: 0, Pixel: PixelOn}
	o.In() <- OLEDCommand{Type: CmdSetPixel, X: 0, Y: -1, Pixel: PixelOn}
	o.In() <- OLEDCommand{Type: CmdSetPixel, X: 99, Y: 99, Pixel: PixelOn}
	o.In() <- OLEDCommand{Type: CmdSetPixel, X: 1, Y: 2, Pixel: PixelOn}

	require.Eventually(t, func() bool {
		return o.Background.BitAt(1, 2) == image1bit.Bit(PixelOn)
	}, 250*time.Millisecond, 5*time.Millisecond)

	cancel()
	require.NoError(t, <-errCh)
}
