//go:build !linux

package drivers

import (
	"fmt"
	"log/slog"
	"time"
)

// LineReqOption is a stub type for non-Linux platforms
type LineReqOption interface{}

// Pin initializes the given GPIO pin, name and mode (stub for non-Linux)
func (gpio *GPIO) Pin(name string, offset int, opts ...LineReqOption) *DigitalPin {
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
	if err := p.Init(); err != nil {
		slog.Error(err.Error(), "name", name, "offset", offset)
	}
	return p
}

func NewDigitalPin(name string, offset int, opts ...LineReqOption) *DigitalPin {
	gpio := GetGPIO()
	pin := gpio.Pin(name, offset, opts...)
	pin.On = pin.ON
	pin.Off = pin.OFF

	return pin
}

// Init the pin (stub implementation for non-Linux)
func (p *DigitalPin) Init() error {
	slog.Info("Using stub GPIO implementation (not on Linux)")
	line := GetMockLine(p.offset)
	p.mock = true
	p.Line = line
	return nil
}

func (p *DigitalPin) Reconfigure(opts ...interface{}) error {
	// Stub implementation - does nothing on non-Linux platforms
	return nil
}

func (d *DigitalPin) EventLoop(done chan any, readpub func()) {
	// Stub implementation for non-Linux platforms
	running := true
	for running {
		select {
		case <-done:
			running = false
		}
	}
}

// MockLine is a stub Line implementation for non-Linux platforms
type MockLine struct {
	offset int `json:"offset"`
	Val    int `json:"val"`
	start  time.Time
}

func GetMockLine(offset int, opts ...LineReqOption) *MockLine {
	m := &MockLine{
		offset: offset,
		start:  time.Now(),
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

func (d *DigitalPin) MockHWInput(v int) {
	m, ok := d.Line.(*MockLine)
	if !ok {
		slog.Error("MockHWInput called on non-mock line")
		return
	}
	m.Val = v
}

func (m *MockLine) MockHWInput(v int) {
	m.Val = v
}

// Stub functions and types to provide compatibility

// AsOutput returns a stub option for non-Linux platforms
func AsOutput(value int) LineReqOption {
	return &outputOption{value: value}
}

type outputOption struct {
	value int
}

// AsInput returns a stub option for non-Linux platforms
func AsInput() interface{} {
	return &inputOption{}
}

type inputOption struct{}

// WithPullUp is a stub for non-Linux platforms
var WithPullUp LineReqOption = &pullUpOption{}

type pullUpOption struct{}

// WithDebounce returns a stub option for non-Linux platforms
func WithDebounce(d time.Duration) LineReqOption {
	return &debounceOption{duration: d}
}

type debounceOption struct {
	duration time.Duration
}

// EventHandler is a stub type for non-Linux platforms
type EventHandler func(LineEvent)

// LineEvent is a stub type for non-Linux platforms
type LineEvent struct {
	Offset    int
	Timestamp time.Duration
	Type      LineEventType
	Seqno     uint32
	LineSeqno uint32
}

// LineEventType is a stub type for non-Linux platforms
type LineEventType int

const (
	LineEventRisingEdge  LineEventType = 1
	LineEventFallingEdge LineEventType = 2
)

// WithEventHandler returns a stub option for non-Linux platforms
func WithEventHandler(handler func(LineEvent)) LineReqOption {
	return &eventHandlerOption{handler: handler}
}

type eventHandlerOption struct {
	handler func(LineEvent)
}

// Helper to check if running on unsupported platform
func init() {
	fmt.Println("Warning: GPIO drivers are running in stub mode (non-Linux platform)")
}
