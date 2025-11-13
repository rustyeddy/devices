package oled

import (
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/gif"
	"math"
	"os"
	"time"

	"github.com/nfnt/resize"
	"github.com/rustyeddy/devices"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/devices/v3/ssd1306"
	"periph.io/x/devices/v3/ssd1306/image1bit"
	"periph.io/x/host/v3"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// Bit is used to turn on or off a Bit on the ssd1306 OLED display
type Bit bool

const (
	On  Bit = true
	Off Bit = false
)

type OLED struct {
	id         string
	Dev        *ssd1306.Dev
	Height     int
	Width      int
	Font       *basicfont.Face
	Background *image1bit.VerticalLSB

	// devices.Device[OLED]

	bus  string
	addr int
}

func NewDevice(id string) (devices.Device[any], error) {
	o, err := New(id, 128, 64)
	return o, err
}

func New(id string, width, height int) (*OLED, error) {
	o := &OLED{
		id:     id,
		Height: height,
		Width:  width,
		bus:    "/dev/i2c-1",
		addr:   0x3c,
	}
	o.Device = o
	o.Background = image1bit.NewVerticalLSB(image.Rect(0, 0, width, height))
	if devices.IsMock() {
		return o, nil
	}
	return o, nil
}

func (o *OLED) ID() string {
	return o.id
}

func (o *OLED) Type() devices.Type {
	return devices.TypeAny
}

func (o *OLED) Open() error {

	// Load all the drivers:
	if _, err := host.Init(); err != nil {
		return err
	}

	// Open a handle to the first available I²C bus:
	bus, err := i2creg.Open(o.bus)
	if err != nil {
		return err
	}

	// Open a handle to a ssd1306 connected on the I²C bus:
	opts := &ssd1306.DefaultOpts
	opts.H = o.Height
	opts.W = o.Width

	o.Dev, err = ssd1306.NewI2C(bus, opts)
	if err != nil {
		return err
	}

	return nil
}

func (o *OLED) Close() error {
	return errors.New("TODO Need to implement bme280 close")
}

func (d *OLED) Clear() {
	// got to be a better way
	d.Rectangle(0, 0, d.Width, d.Height, Off)
}

func (d *OLED) Get() (devices.Value, error) {
	return nil, devices.ErrNotImplemented
}

func (d *OLED) Set(v devices.Value) error {

	// what to do with set?
	return nil
}

func (d *OLED) Draw() error {
	err := d.Dev.Draw(d.Background.Bounds(), d.Background, image.Point{})
	if err != nil {
		fmt.Println("ERROR - OLED Draw: ", err)
		return err
	}
	return nil
}
func (d *OLED) Rectangle(x0, y0, x1, y1 int, value Bit) {
	d.Clip(&x0, &y0, &x1, &y1)

	if x0 > x1 {
		x0, x1 = x1, x0
	}
	if y0 > y1 {
		y0, y1 = y1, y0
	}

	for x := x0; x < x1; x++ {
		for y := y0; y < y1; y++ {
			d.SetBit(x, y, value)
		}
	}
}

func (d *OLED) Line(x0, y0, len, width int, value Bit) {
	x1 := x0 + len
	y1 := y0 + width
	d.Rectangle(x0, y0, x1, y1, value)
}

func (d *OLED) Diagonal(x0, y0, x1, y1 int, value Bit) {
	d.Clip(&x0, &y0, &x1, &y1)

	xf0 := float64(x0)
	xf1 := float64(x1)
	yf0 := float64(y0)
	yf1 := float64(y1)

	l := (xf1 - xf0)
	h := (yf1 - yf0)

	var slope float64
	if l > h {
		slope = h / l
	} else {
		slope = l / h
	}

	if l >= h {
		for x := xf0; x < xf1; x++ {
			y := slope*(x-xf0) + yf0
			d.SetBit(int(math.Round(x)), int(math.Round(y)), value)
		}
	} else {
		for y := yf0; y < yf1; y++ {
			x := slope*(y-yf0) + xf0
			d.SetBit(int(math.Round(x)), int(math.Round(y)), value)
		}
	}
}

func (d *OLED) Scroll(o ssd1306.Orientation, rate ssd1306.FrameRate, startLine, endLine int) error {
	return d.Dev.Scroll(o, rate, startLine, endLine)
}

func (d *OLED) StopScroll() error {
	return d.Dev.StopScroll()
}

func (d *OLED) SetBit(x, y int, value Bit) {
	d.Background.SetBit(x, y, image1bit.Bit(value))
}

func (d *OLED) Clip(x0, y0, x1, y1 *int) {
	if x0 != nil {
		if *x0 < 0 {
			*x0 = 0
		}
		if *x0 > d.Width {
			*x0 = d.Width
		}
	}
	if x1 != nil {
		if *x1 < 0 {
			*x1 = 0
		}
		if *x1 > d.Width {
			*x1 = d.Width
		}
	}

	if y0 != nil {
		if *y0 < 0 {
			*y0 = 0
		}
		if *y0 > d.Height {
			*y0 = d.Height
		}
	}

	if y1 != nil {
		if *y1 < 0 {
			*y1 = 0
		}
		if *y1 > d.Height {
			*y1 = d.Height
		}
	}
}

func (d *OLED) DrawString(x, y int, str string) {
	d.Font = basicfont.Face7x13
	drawer := &font.Drawer{
		Dst:  d.Background,
		Src:  image.White,
		Face: d.Font,
		Dot:  fixed.Point26_6{X: fixed.I(x), Y: fixed.I(y)},
	}
	drawer.DrawString(str)
}

func (d *OLED) AnimatedGIF(fname string, done <-chan time.Time) error {
	f, err := os.Open(fname)
	if err != nil {
		return err
	}

	g, err := gif.DecodeAll(f)
	f.Close()
	if err != nil {
		return err
	}

	// Converts every frame to image.Gray and resize them:
	imgs := make([]*image.Gray, len(g.Image))
	for i := range g.Image {
		imgs[i] = d.convertAndResizeAndCenter(g.Image[i])
	}

	// OLED the frames in a loop:
	for i := 0; ; i++ {
		select {
		case <-done:
			return nil

		default:
			index := i % len(imgs)
			c := time.After(time.Duration(10*g.Delay[index]) * time.Millisecond)
			img := imgs[index]
			d.Dev.Draw(img.Bounds(), img, image.Point{})
			<-c
		}
	}
	return nil
}

// convertAndResizeAndCenter takes an image, resizes and centers it on a
// image.Gray of size w*h.
func (d *OLED) convertAndResizeAndCenter(src image.Image) *image.Gray {
	w := d.Width
	h := d.Height
	src = resize.Thumbnail(uint(w), uint(h), src, resize.Bicubic)
	img := image.NewGray(image.Rect(0, 0, w, h))
	r := src.Bounds()
	r = r.Add(image.Point{(w - r.Max.X) / 2, (h - r.Max.Y) / 2})
	draw.Draw(img, r, src, image.Point{}, draw.Src)
	return img
}
