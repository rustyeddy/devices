package oled

import (
	"image"
	"testing"
	"time"

	"github.com/rustyeddy/devices"
	"github.com/stretchr/testify/assert"
	"periph.io/x/devices/v3/ssd1306/image1bit"
)

func init() {
	devices.SetMock(true)
}

func TestNewOLED_MockMode(t *testing.T) {
	oled, err := New("testoled", 128, 64)
	assert.NoError(t, err)
	assert.NotNil(t, oled)
	assert.Equal(t, 128, oled.Width)
	assert.Equal(t, 64, oled.Height)
	assert.Equal(t, "testoled", oled.ID())
}

func TestOLED_Clear_SetsAllBitsOff(t *testing.T) {
	oled, _ := New("clearoled", 10, 10)
	oled.Rectangle(0, 0, 10, 10, On)
	oled.Clear()
	for x := 0; x < oled.Width; x++ {
		for y := 0; y < oled.Height; y++ {
			bit := oled.Background.BitAt(x, y)
			assert.Equal(t, image1bit.Bit(Off), bit)
		}
	}
}

func TestOLED_Rectangle_SetsBits(t *testing.T) {
	oled, _ := New("rectoled", 8, 8)
	oled.Rectangle(2, 2, 6, 6, On)
	for x := 2; x < 6; x++ {
		for y := 2; y < 6; y++ {
			assert.Equal(t, image1bit.Bit(On), oled.Background.BitAt(x, y))
		}
	}
}

func TestOLED_Line_SetsBits(t *testing.T) {
	oled, _ := New("lineoled", 8, 8)
	oled.Line(1, 1, 4, 2, On)
	for x := 1; x < 5; x++ {
		for y := 1; y < 3; y++ {
			assert.Equal(t, image1bit.Bit(On), oled.Background.BitAt(x, y))
		}
	}
}

func TestOLED_Diagonal_SetsBits(t *testing.T) {
	oled, _ := New("diagoled", 8, 8)
	oled.Diagonal(0, 0, 7, 7, On)
	// Diagonal should set bits roughly along the diagonal
	count := 0
	for i := 0; i < 8; i++ {
		if oled.Background.BitAt(i, i) == image1bit.Bit(On) {
			count++
		}
	}
	assert.GreaterOrEqual(t, count, 6)
}

func TestOLED_Clip_ClampsValues(t *testing.T) {
	oled, _ := New("clipoled", 10, 10)
	x0, y0, x1, y1 := -5, -5, 15, 15
	oled.Clip(&x0, &y0, &x1, &y1)
	assert.Equal(t, 0, x0)
	assert.Equal(t, 0, y0)
	assert.Equal(t, 10, x1)
	assert.Equal(t, 10, y1)
}

func TestOLED_SetBit_SetsCorrectBit(t *testing.T) {
	oled, _ := New("setbitoled", 5, 5)
	oled.SetBit(2, 3, On)
	assert.Equal(t, image1bit.Bit(On), oled.Background.BitAt(2, 3))
	oled.SetBit(2, 3, Off)
	assert.Equal(t, image1bit.Bit(Off), oled.Background.BitAt(2, 3))
}

func TestOLED_DrawString_SetsPixels(t *testing.T) {
	oled, _ := New("stringoled", 20, 20)
	oled.DrawString(2, 10, "Hi")
	// Check that some pixels are set
	set := false
	for x := 2; x < 10; x++ {
		for y := 5; y < 15; y++ {
			if oled.Background.BitAt(x, y) == image1bit.Bit(On) {
				set = true
				break
			}
		}
	}
	assert.True(t, set)
}

func TestOLED_AnimatedGIF_ReturnsErrorForMissingFile(t *testing.T) {
	oled, _ := New("gifoled", 32, 32)
	done := make(chan time.Time)
	err := oled.AnimatedGIF("nonexistent.gif", done)
	assert.Error(t, err)
}

func TestOLED_convertAndResizeAndCenter_ResizesAndCenters(t *testing.T) {
	oled, _ := New("resizeoled", 32, 32)
	src := image.NewGray(image.Rect(0, 0, 16, 16))
	img := oled.convertAndResizeAndCenter(src)
	assert.Equal(t, 32, img.Bounds().Dx())
	assert.Equal(t, 32, img.Bounds().Dy())
}
