package drivers

/*
gpiocdev.AsOutput(0)
gpiocdev.AsOutput(1)
gpiocdev.AsInput
gpiocdev.LineReqOption
gpiocdev.LineEvent
gpiocdev.LineEventFallingEdge
gpiocdev.LineEventRisingEdge
gpiocdev.WithEventHandler
gpiocdev.WithDebounce(10 * time.Millisecond)
gpiocdev.WithBothEdges
gpiocdev.WithPullDown
gpiocdev.WithPullUp
*/

import (
	"errors"
	"time"

	"log/slog"
)

var (
	ErrOutOfRange    = errors.New("pin out of range")
	ErrPinIsAnOutput = errors.New("can not set an output pin")
)

// PIN_COUNT provides the number of pins per VPIO bank
const PIN_COUNT = 64

type VPin[T Value] struct {
	id        string
	index     int
	options   []PinOptions
	value     T
}

func (v *VPIO[T]) Open() error {
	return nil
}

// Close releases VPIO resources (no-op for virtual GPIO)
func (v *VPIO[T]) Close() error {
	return nil
}

// ID returns the pin identifier
func (p *VPin[T]) ID() string {
	return p.id
}

// Index returns the pin index
func (p *VPin[T]) Index() int {
	return int(p.index)
}

// Direction returns the pin direction
func (p *VPin[T]) Options() []PinOptions {
	return p.options
}

// Get returns the current value of the pin
func (p *VPin[T]) Get() (T, error) {
	return p.value, nil
}

// Set sets the value of the pin
func (p *VPin[T]) Set(val T) error {
	// if p.direction != DirectionOutput {
	// 	var zero T
	// 	p.value = zero
	// 	return ErrPinIsAnOutput
	// }
	p.value = val
	return nil
}

func (v *VPin[T]) ReadContinuous() <-chan T {
	valq := make(chan T)
	go func() {
		for {
			val, err := v.Get()
			if err != nil {
				slog.Error("Errored to in v.Get() read ", "error", err)
				continue
			}
			valq <- val
		}
	}()
	return valq	
}

type Transaction[T Value] struct {
	index int
	value T
	time.Time
}

type VPIO[T Value] struct {
	pins[PIN_COUNT]VPin[T]
	recording    bool
	transactions []*Transaction[T]
}

func NewVPIO[T Value]() *VPIO[T] {
	return &VPIO[T]{}
}

func (v *VPIO[T]) SetPin(id string, i int, opts ...PinOptions) (Pin[T], error) {
	if i >= PIN_COUNT {
		return nil, ErrOutOfRange
	}
	v.pins[i].id = id
	v.pins[i].index = i
	v.pins[i].options = opts
	return &v.pins[i], nil
}

func (v *VPIO[T]) Set(i int, val T) error {
	if i >= PIN_COUNT {
		return ErrOutOfRange
	}
	v.pins[i].value = val

	if v.recording {
		trans := &Transaction[T]{
			index: i,
			value: val,
			Time:  time.Now(),
		}
		v.transactions = append(v.transactions, trans)
	}

	return nil
}

func (v *VPIO[T]) Get(i int) (T, error) {
	var zero T
	if i >= PIN_COUNT {
		return zero, ErrOutOfRange
	}
	// OK to get the value of an input pin
	return v.pins[i].value, nil
}

// StartRecording enables transaction recording
func (v *VPIO[T]) StartRecording() {
	v.recording = true
	v.transactions = make([]*Transaction[T], 0)
}

// StopRecording disables transaction recording
func (v *VPIO[T]) StopRecording() {
	v.recording = false
}

// GetTransactions returns all recorded transactions
func (v *VPIO[T]) GetTransactions() []*Transaction[T] {
	return v.transactions
}

// ClearTransactions removes all recorded transactions
func (v *VPIO[T]) ClearTransactions() {
	v.transactions = make([]*Transaction[T], 0)
}
