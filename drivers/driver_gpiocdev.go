//go:build linux

package drivers

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/warthog618/go-gpiocdev"
)

// GPIOCDev implements a GPIO controller using go-gpiocdev for real hardware
type GPIOCDev struct {
	Chipname string
	pins     map[int]*GPIOCDevPin
	mu       sync.RWMutex
}

// GPIOCDevPin implements the Pin interface for hardware GPIO pins
type GPIOCDevPin struct {
	id        string
	index     int
	direction Direction
	line      *gpiocdev.Line
	opts      []gpiocdev.LineReqOption
	value     int
	evtQ      chan gpiocdev.LineEvent
	mu        sync.RWMutex
}

// NewGPIOCDev creates a new GPIOCDev instance
// chipname should be "gpiochip0" for Raspberry Pi Zero or "gpiochip4" for Raspberry Pi 5
func NewGPIOCDev(chipname string) *GPIOCDev {
	if chipname == "" {
		chipname = "gpiochip0"
	}
	return &GPIOCDev{
		Chipname: chipname,
		pins:     make(map[int]*GPIOCDevPin),
	}
}

// Pin initializes and returns a GPIO pin
func (g *GPIOCDev) Pin(name string, pinIndex int, options PinOptions) (*GPIOCDevPin, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if existing, ok := g.pins[pinIndex]; ok {
		return existing, nil
	}

	opts, dir := pinOptionsToGPIOCDev(options)

	pin := &GPIOCDevPin{
		id:        name,
		index:     pinIndex,
		direction: dir,
		opts:      opts,
		evtQ:      make(chan gpiocdev.LineEvent, 10),
	}

	line, err := gpiocdev.RequestLine(g.Chipname, pinIndex, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to request GPIO line %d: %w", pinIndex, err)
	}
	pin.line = line
	g.pins[pinIndex] = pin

	slog.Info("GPIO pin initialized", "name", name, "index", pinIndex, "direction", dir)
	return pin, nil
}

// Get retrieves the value of a specific pin
func (g *GPIOCDev) Get(pinIndex int) (int, error) {
	g.mu.RLock()
	pin, exists := g.pins[pinIndex]
	g.mu.RUnlock()

	if !exists {
		return 0, fmt.Errorf("pin %d not initialized", pinIndex)
	}
	return pin.Get()
}

// Set sets the value of a specific pin
func (g *GPIOCDev) Set(pinIndex int, value int) error {
	g.mu.RLock()
	pin, exists := g.pins[pinIndex]
	g.mu.RUnlock()

	if !exists {
		return fmt.Errorf("pin %d not initialized", pinIndex)
	}
	return pin.Set(value)
}

// Close releases all GPIO resources
func (g *GPIOCDev) Close() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	for _, pin := range g.pins {
		if err := pin.Close(); err != nil {
			slog.Error("failed to close pin", "index", pin.index, "error", err)
		}
	}
	g.pins = make(map[int]*GPIOCDevPin)
	return nil
}

// ID returns the pin identifier
func (p *GPIOCDevPin) ID() string {
	return p.id
}

// Index returns the pin index/offset
func (p *GPIOCDevPin) Index() int {
	return p.index
}

// Direction returns the pin direction
func (p *GPIOCDevPin) Direction() Direction {
	return p.direction
}

// Get reads the current value of the pin
func (p *GPIOCDevPin) Get() (int, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.line == nil {
		return 0, fmt.Errorf("pin %d not initialized", p.index)
	}

	val, err := p.line.Value()
	if err != nil {
		return 0, fmt.Errorf("failed to read pin %d: %w", p.index, err)
	}

	p.value = val
	return val, nil
}

// Set writes a value to the pin
func (p *GPIOCDevPin) Set(value int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.line == nil {
		return fmt.Errorf("pin %d not initialized", p.index)
	}

	if p.direction != DirectionOutput {
		return fmt.Errorf("cannot set value on input pin %d", p.index)
	}

	if err := p.line.SetValue(value); err != nil {
		return fmt.Errorf("failed to set pin %d: %w", p.index, err)
	}

	p.value = value
	return nil
}

// On sets the pin high (1)
func (p *GPIOCDevPin) On() error {
	return p.Set(1)
}

// Off sets the pin low (0)
func (p *GPIOCDevPin) Off() error {
	return p.Set(0)
}

// Toggle flips the pin value
func (p *GPIOCDevPin) Toggle() error {
	val, err := p.Get()
	if err != nil {
		return err
	}

	if val == 0 {
		return p.Set(1)
	}
	return p.Set(0)
}

// Reconfigure changes the pin configuration
func (p *GPIOCDevPin) Reconfigure(opts ...gpiocdev.LineConfigOption) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.line == nil {
		return fmt.Errorf("pin %d not initialized", p.index)
	}
	return p.line.Reconfigure(opts...)
}

// EventLoop processes GPIO events
func (p *GPIOCDevPin) EventLoop(done chan any, callback func(evt gpiocdev.LineEvent)) {
	running := true
	for running {
		select {
		case evt := <-p.evtQ:
			evtype := "unknown"
			switch evt.Type {
			case gpiocdev.LineEventFallingEdge:
				evtype = "falling"
			case gpiocdev.LineEventRisingEdge:
				evtype = "rising"
			}

			slog.Info("GPIO event", "pin", p.id, "direction", evtype, "seqno", evt.Seqno, "lineseq", evt.LineSeqno)

			if callback != nil {
				callback(evt)
			}
		case <-done:
			running = false
		}
	}
}

// Close releases the pin resources
func (p *GPIOCDevPin) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.line != nil {
		_ = p.line.Reconfigure(gpiocdev.AsInput)
		if err := p.line.Close(); err != nil {
			return err
		}
		p.line = nil
	}

	if p.evtQ != nil {
		close(p.evtQ)
		p.evtQ = nil
	}
	return nil
}

// pinOptionsToGPIOCDev converts PinOptions to gpiocdev options
func pinOptionsToGPIOCDev(options PinOptions) ([]gpiocdev.LineReqOption, Direction) {
	var opts []gpiocdev.LineReqOption
	var dir Direction = DirectionNone

	const (
		PinInput       PinOptions = 1 << 0
		PinOutput      PinOptions = 1 << 1
		PinOutputLow   PinOptions = 1 << 2
		PinOutputHigh  PinOptions = 1 << 3
		PinPullUp      PinOptions = 1 << 4
		PinPullDown    PinOptions = 1 << 5
		PinRisingEdge  PinOptions = 1 << 6
		PinFallingEdge PinOptions = 1 << 7
		PinBothEdges   PinOptions = 1 << 8
	)

	if options&PinOutput != 0 {
		if options&PinOutputHigh != 0 {
			opts = append(opts, gpiocdev.AsOutput(1))
		} else {
			opts = append(opts, gpiocdev.AsOutput(0))
		}
		dir = DirectionOutput
	} else if options&PinInput != 0 {
		opts = append(opts, gpiocdev.AsInput)
		dir = DirectionInput
	}

	if options&PinPullUp != 0 {
		opts = append(opts, gpiocdev.WithPullUp)
	}
	if options&PinPullDown != 0 {
		opts = append(opts, gpiocdev.WithPullDown)
	}

	if options&PinBothEdges != 0 {
		opts = append(opts, gpiocdev.WithBothEdges)
	} else {
		if options&PinRisingEdge != 0 {
			opts = append(opts, gpiocdev.WithEventHandler(eventHandler))
		}
		if options&PinFallingEdge != 0 {
			opts = append(opts, gpiocdev.WithEventHandler(eventHandler))
		}
	}

	if dir == DirectionInput && (options&(PinRisingEdge|PinFallingEdge|PinBothEdges)) != 0 {
		opts = append(opts, gpiocdev.WithDebounce(10*time.Millisecond))
	}

	return opts, dir
}

var eventHandler = func(evt gpiocdev.LineEvent) {
	slog.Debug("GPIO event received", "offset", evt.Offset, "type", evt.Type)
}

func SetEventHandler(handler func(gpiocdev.LineEvent)) {
	eventHandler = handler
}
