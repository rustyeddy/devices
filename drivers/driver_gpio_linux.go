//go:build linux

package drivers

import (
	"log/slog"
	"time"

	"github.com/warthog618/go-gpiocdev"
)

// Pin initializes the given GPIO pin, name and mode
func (gpio *GPIO) Pin(name string, offset int, opts ...gpiocdev.LineReqOption) *DigitalPin {

	var dopts []gpiocdev.LineReqOption
	for _, o := range opts {
		dopts = append(dopts, o.(gpiocdev.LineReqOption))
	}
	p := &DigitalPin{
		name:   name,
		offset: offset,
	}
	p.On = p.ON
	p.Off = p.OFF

	if gpio.pins == nil {
		gpio.pins = make(map[int]*DigitalPin)
	}
	gpio.pins[offset] = p
	if err := p.initLinux(gpio.Chipname, dopts); err != nil {
		slog.Error(err.Error(), "name", name, "offset", offset)
	}
	return p
}

func NewDigitalPin(name string, offset int, opts ...gpiocdev.LineReqOption) *DigitalPin {
	gpio := GetGPIO()
	pin := gpio.Pin(name, offset, opts...)
	pin.On = pin.ON
	pin.Off = pin.OFF

	return pin
}

// lineWrapper wraps gpiocdev.Line to implement our Line interface
type lineWrapper struct {
	*gpiocdev.Line
}

func (w *lineWrapper) Reconfigure(opts ...interface{}) error {
	// Convert interface{} to gpiocdev.LineConfigOption
	var configOpts []gpiocdev.LineConfigOption
	for _, opt := range opts {
		if co, ok := opt.(gpiocdev.LineConfigOption); ok {
			configOpts = append(configOpts, co)
		}
	}
	return w.Line.Reconfigure(configOpts...)
}

// Init the pin from the offset and mode
func (p *DigitalPin) Init() error {
	gpio := GetGPIO()
	if gpio.Mock {
		line := GetMockLine(p.offset)
		p.mock = true
		p.Line = line
		return nil
	}

	line, err := gpiocdev.RequestLine(gpio.Chipname, p.offset)
	if err != nil {
		return err
	}
	p.Line = &lineWrapper{Line: line}
	return nil
}

func (p *DigitalPin) initLinux(chipname string, opts []gpiocdev.LineReqOption) error {
	gpio := GetGPIO()
	if gpio.Mock {
		line := GetMockLineWithOpts(p.offset, opts...)
		p.mock = true
		p.Line = line
		return nil
	}

	line, err := gpiocdev.RequestLine(chipname, p.offset, opts...)
	if err != nil {
		return err
	}
	p.Line = &lineWrapper{Line: line}
	return nil
}

func (p *DigitalPin) ReconfigureGPIO(opts ...gpiocdev.LineConfigOption) error {
	if p.Line == nil {
		return nil
	}
	// Type assert to lineWrapper which supports gpiocdev Reconfigure
	if wrapper, ok := p.Line.(*lineWrapper); ok {
		return wrapper.Line.Reconfigure(opts...)
	}
	// For MockLine, call the generic Reconfigure
	return p.Line.Reconfigure(opts)
}

type LineEventHandler struct {
	EvtQ          chan gpiocdev.LineEvent
	EventHandler  gpiocdev.EventHandler
}

func (d *DigitalPin) SetEventHandler(evtQ chan gpiocdev.LineEvent, handler gpiocdev.EventHandler) {
	// Store the event queue and handler for use in EventLoop
	// This is a helper method for button-like devices
}

func (d *DigitalPin) EventLoop(done chan any, readpub func()) {
	// This would need the event queue to be set up properly
	// For now, this is a placeholder
	running := true
	for running {
		select {
		case <-done:
			running = false
		}
	}
}

// MockGPIO fakes the Line interface on computers that don't
// actually have GPIO pins mostly for mocking tests
type MockLine struct {
	offset int `json:"offset"`
	Val    int `json:"val"`
	start  time.Time
}

func GetMockLine(offset int) *MockLine {
	m := &MockLine{
		offset: offset,
		start:  time.Now(),
	}
	return m
}

func GetMockLineWithOpts(offset int, opts ...gpiocdev.LineReqOption) *MockLine {
	m := &MockLine{
		offset: offset,
		start:  time.Now(),
	}
	for _, opt := range opts {
		switch v := opt.(type) {
		case gpiocdev.OutputOption:
			m.Val = v[0]
		default:
			// slog.Debug("MockLine does not record", "optType", v)
		}
	}

	return m
}

func (m MockLine) Close() error {
	return nil
}

func (m MockLine) Offset() int {
	return m.offset
}

func (m *MockLine) SetValue(val int) error {
	m.Val = val
	return nil
}

func (m MockLine) Reconfigure(...interface{}) error {
	return nil
}

func (m MockLine) Value() (int, error) {
	return m.Val, nil
}

var seqno uint32

func getSeqno() uint32 {
	seqno += 1
	return seqno
}

func (d *DigitalPin) MockHWInput(v int) {
	m := d.Line.(*MockLine)
	m.MockHWInput(v)
}

func (m *MockLine) MockHWInput(v int) {
	m.Val = v
	// Event handling would go here if needed
}
