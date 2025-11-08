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
)

type Direction uint

type Value interface {
	~int | ~float64 | ~bool
}

const (
	DirectionNone	= iota
	DirectionInput
	DirectionOutput
)

var (
	ErrOutOfRange = errors.New("pin out of range")
	ErrPinIsAnOutput = errors.New("can not set an output pin")
)

// PIN_COUNT provides the number of pins per VPIO bank
const PIN_COUNT	= 64

type VPin[T Value] struct {
	id			string
	index		uint
	direction	Direction
	value		T
}

type VPIO[T Value] struct {
	pins	[PIN_COUNT]VPin[T]
}

func NewVPIO[T Value]() *VPIO[T] {
	return &VPIO[T]{}
}

func (v *VPIO[T]) Pin(id string, i uint, dir Direction) (*VPin[T], error) {
	if i > PIN_COUNT {
		return nil, ErrOutOfRange
	}
	p := v.pins[i]
	p.id = id
	p.index = i
	p.direction = dir
	return &p, nil
}

func (v *VPIO[T]) Set(i uint, val T) error {
	if i > PIN_COUNT {
		return ErrOutOfRange
	}
	v.pins[i].value = val
	return nil
}

func (v *VPIO[T]) Get(i uint) (T, error) {
	if i > PIN_COUNT {
		return v.pins[i].value, ErrOutOfRange
	}
	// OK to get the value of an input pin
	return v.pins[i].value, nil
}



