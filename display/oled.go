package display

import (
	"context"
	"errors"
	"fmt"
	"image"
	"image/draw"
	"math"

	"github.com/nfnt/resize"
	"github.com/rustyeddy/devices"
	"github.com/rustyeddy/devices/drivers"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
	"periph.io/x/devices/v3/ssd1306/image1bit"
)

// Pixel controls whether a pixel is on or off.
type Pixel bool

const (
	PixelOn  Pixel = true
	PixelOff Pixel = false
)

// OLEDConfig configures an SSD1306 OLED display.
type OLEDConfig struct {
	Name string

	// Factory opens the underlying OLED driver.
	// Provide drivers.PeriphOLEDFactory on supported platforms or a mock in tests.
	Factory drivers.OLEDFactory

	// Bus and Addr select the I2C bus and device address.
	// Common Raspberry Pi bus is "1" or "" (periph default).
	Bus  string
	Addr uint16

	// Width/Height are the pixel dimensions (e.g. 128x64).
	Width  int
	Height int

	// Buf sizes the command channel. Default 16.
	Buf int
}

// OLEDCommandType identifies the command.
type OLEDCommandType string

const (
	CmdClear     OLEDCommandType = "clear"
	CmdSetPixel  OLEDCommandType = "set_pixel"
	CmdRect      OLEDCommandType = "rect"
	CmdLine      OLEDCommandType = "line"
	CmdDiagonal  OLEDCommandType = "diagonal"
	CmdText      OLEDCommandType = "text"
	CmdFlush     OLEDCommandType = "flush"
	CmdDrawImage OLEDCommandType = "draw_image"
)

// OLEDCommand is a single drawing or control action.
//
// Commands are best-effort; if Done is provided, the device will send
// an error (or nil) after applying the command.
type OLEDCommand struct {
	Type OLEDCommandType

	// Coordinates / dimensions
	X0, Y0, X1, Y1 int
	X, Y           int
	Len, Width     int

	// Payload
	Pixel Pixel
	Text  string
	Img   image.Image

	// Optional ack.
	Done chan error
}

// OLED is a channel-driven SSD1306 display.
//
// It maintains an internal 1-bit frame buffer (Background). Most drawing commands
// mutate the buffer; CmdFlush pushes the buffer to the hardware.
type OLED struct {
	devices.Base
	in chan OLEDCommand

	cfg OLEDConfig
	drv drivers.OLED

	// Publicly readable for convenience (tests / debugging).
	Width      int
	Height     int
	Background *image1bit.VerticalLSB
}

// NewOLED constructs an OLED device.
func NewOLED(cfg OLEDConfig) *OLED {
	if cfg.Buf <= 0 {
		cfg.Buf = 16
	}
	if cfg.Width == 0 {
		cfg.Width = 128
	}
	if cfg.Height == 0 {
		cfg.Height = 64
	}
	o := &OLED{
		Base:   devices.NewBase(cfg.Name, cfg.Buf),
		in:     make(chan OLEDCommand, cfg.Buf),
		cfg:    cfg,
		Width:  cfg.Width,
		Height: cfg.Height,
	}
	o.Background = image1bit.NewVerticalLSB(image.Rect(0, 0, o.Width, o.Height))
	return o
}

// In returns the command channel.
func (o *OLED) In() chan<- OLEDCommand { return o.in }

// Descriptor returns display metadata.
func (o *OLED) Descriptor() devices.Descriptor {
	attrs := map[string]string{
		"bus":    o.cfg.Bus,
		"addr":   fmt.Sprintf("0x%02x", o.cfg.Addr),
		"width":  fmt.Sprintf("%d", o.Width),
		"height": fmt.Sprintf("%d", o.Height),
	}
	return devices.Descriptor{
		Name:       o.Name(),
		Kind:       "oled",
		ValueType:  "command",
		Access:     devices.WriteOnly,
		Tags:       []string{"display", "i2c", "ssd1306"},
		Attributes: attrs,
	}
}

// Run opens the OLED and applies commands until ctx is canceled.
func (o *OLED) Run(ctx context.Context) error {
	o.Emit(devices.EventOpen, "run", nil, nil)

	if o.cfg.Factory == nil {
		err := errors.New("oled: factory is nil")
		o.Emit(devices.EventError, "factory missing", err, nil)
		o.Close()
		return err
	}
	if o.Width <= 0 || o.Height <= 0 {
		err := fmt.Errorf("oled: invalid dimensions %dx%d", o.Width, o.Height)
		o.Emit(devices.EventError, "invalid dimensions", err, nil)
		o.Close()
		return err
	}

	drv, err := o.cfg.Factory.OpenSSD1306(o.cfg.Bus, o.cfg.Addr, o.Width, o.Height)
	if err != nil {
		o.Emit(devices.EventError, "open failed", err, nil)
		o.Close()
		return err
	}
	o.drv = drv

	defer func() {
		_ = o.drv.Close()
		o.Emit(devices.EventClose, "stop", nil, nil)
		o.Close()
	}()

	for {
		select {
		case cmd := <-o.in:
			err := o.apply(cmd)
			if cmd.Done != nil {
				// Non-blocking ack
				select {
				case cmd.Done <- err:
				default:
				}
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (o *OLED) apply(cmd OLEDCommand) error {
	switch cmd.Type {
	case CmdClear:
		o.rect(0, 0, o.Width, o.Height, PixelOff)
		o.Emit(devices.EventInfo, "clear", nil, nil)
		return nil

	case CmdSetPixel:
		o.setPixel(cmd.X, cmd.Y, cmd.Pixel)
		return nil

	case CmdRect:
		o.rect(cmd.X0, cmd.Y0, cmd.X1, cmd.Y1, cmd.Pixel)
		return nil

	case CmdLine:
		o.line(cmd.X0, cmd.Y0, cmd.Len, cmd.Width, cmd.Pixel)
		return nil

	case CmdDiagonal:
		o.diagonal(cmd.X0, cmd.Y0, cmd.X1, cmd.Y1, cmd.Pixel)
		return nil

	case CmdText:
		o.drawString(cmd.X, cmd.Y, cmd.Text)
		return nil

	case CmdFlush:
		if o.drv == nil {
			return errors.New("oled: driver not open")
		}
		if err := o.drv.Draw(o.Background.Bounds(), o.Background, image.Point{}); err != nil {
			o.Emit(devices.EventError, "draw failed", err, nil)
			return err
		}
		o.Emit(devices.EventInfo, "flush", nil, nil)
		return nil

	case CmdDrawImage:
		if o.drv == nil {
			return errors.New("oled: driver not open")
		}
		if cmd.Img == nil {
			return errors.New("oled: image is nil")
		}
		img := o.convertAndResizeAndCenter(cmd.Img)
		if err := o.drv.Draw(img.Bounds(), img, image.Point{}); err != nil {
			o.Emit(devices.EventError, "draw image failed", err, nil)
			return err
		}
		o.Emit(devices.EventInfo, "draw_image", nil, nil)
		return nil

	default:
		return fmt.Errorf("oled: unknown command %q", cmd.Type)
	}
}

func (o *OLED) setPixel(x, y int, p Pixel) {
	if x < 0 || y < 0 || x >= o.Width || y >= o.Height {
		return
	}
	o.Background.SetBit(x, y, image1bit.Bit(p))
}

func (o *OLED) clip(x0, y0, x1, y1 *int) {
	if x0 != nil {
		if *x0 < 0 {
			*x0 = 0
		}
		if *x0 > o.Width {
			*x0 = o.Width
		}
	}
	if x1 != nil {
		if *x1 < 0 {
			*x1 = 0
		}
		if *x1 > o.Width {
			*x1 = o.Width
		}
	}
	if y0 != nil {
		if *y0 < 0 {
			*y0 = 0
		}
		if *y0 > o.Height {
			*y0 = o.Height
		}
	}
	if y1 != nil {
		if *y1 < 0 {
			*y1 = 0
		}
		if *y1 > o.Height {
			*y1 = o.Height
		}
	}
}

func (o *OLED) rect(x0, y0, x1, y1 int, p Pixel) {
	o.clip(&x0, &y0, &x1, &y1)
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	for x := x0; x < x1; x++ {
		for y := y0; y < y1; y++ {
			o.setPixel(x, y, p)
		}
	}
}

func (o *OLED) line(x0, y0, length, width int, p Pixel) {
	x1 := x0 + length
	y1 := y0 + width
	o.rect(x0, y0, x1, y1, p)
}

func (o *OLED) diagonal(x0, y0, x1, y1 int, p Pixel) {
	o.clip(&x0, &y0, &x1, &y1)

	xf0 := float64(x0)
	xf1 := float64(x1)
	yf0 := float64(y0)
	yf1 := float64(y1)

	l := (xf1 - xf0)
	h := (yf1 - yf0)

	if l == 0 && h == 0 {
		o.setPixel(x0, y0, p)
		return
	}

	var slope float64
	if math.Abs(l) > math.Abs(h) {
		slope = h / l
	} else {
		slope = l / h
	}

	if math.Abs(l) >= math.Abs(h) {
		step := 1.0
		if xf1 < xf0 {
			step = -1.0
		}
		for x := xf0; ; x += step {
			y := slope*(x-xf0) + yf0
			o.setPixel(int(math.Round(x)), int(math.Round(y)), p)
			if (step > 0 && x >= xf1) || (step < 0 && x <= xf1) {
				break
			}
		}
	} else {
		step := 1.0
		if yf1 < yf0 {
			step = -1.0
		}
		for y := yf0; ; y += step {
			x := slope*(y-yf0) + xf0
			o.setPixel(int(math.Round(x)), int(math.Round(y)), p)
			if (step > 0 && y >= yf1) || (step < 0 && y <= yf1) {
				break
			}
		}
	}
}

func (o *OLED) drawString(x, y int, s string) {
	if s == "" {
		return
	}
	face := basicfont.Face7x13
	d := &font.Drawer{
		Dst:  o.Background,
		Src:  image.White,
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)},
	}
	d.DrawString(s)
}

// convertAndResizeAndCenter converts an image, resizes and centers it on a
// *image.Gray of size Width*Height.
func (o *OLED) convertAndResizeAndCenter(src image.Image) *image.Gray {
	w := o.Width
	h := o.Height
	// Resize to fit, keeping aspect ratio.
	src = resize.Thumbnail(uint(w), uint(h), src, resize.Bicubic)
	img := image.NewGray(image.Rect(0, 0, w, h))
	r := src.Bounds()
	r = r.Add(image.Point{X: (w - r.Dx()) / 2, Y: (h - r.Dy()) / 2})
	draw.Draw(img, r, src, image.Point{}, draw.Src)
	return img
}

var _ devices.Sink[OLEDCommand] = (*OLED)(nil)
