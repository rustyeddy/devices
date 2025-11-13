package drivers

import (
	"testing"
	"github.com/stretchr/testify/assert"
	
)

func TestVPIOBool(t *testing.T) {
	pins := []struct{
		id string
		index uint
		dir Direction
		val bool
		err error
	} {
		{ id: "input", index: 1, val: true,  dir: DirectionInput, err: nil},
			{ id: "output", index: 2, val: false, dir: DirectionOutput, err: nil },
			{ id: "bad-index", index: 75, val: false, dir: DirectionOutput, err: ErrOutOfRange },
	}

	v := NewVPIO[bool]()
	for _, pin := range pins {
		p, err := v.Pin(pin.id, pin.index, pin.dir)
		assert.Equal(t, err, pin.err)
		if err == ErrOutOfRange {
			continue
		}

		assert.Equal(t, p.id, pin.id)
		assert.Equal(t, p.index, pin.index)
		if err != nil {
			continue
		}

		err =  v.Set(pin.index, pin.val)
		assert.NoError(t, err)
		val, err := v.Get(pin.index)
		assert.NoError(t, err)
		assert.Equal(t, val, pin.val)
	}

}

func TestVPIOInt(t *testing.T) {
	pins := []struct{
		id string
		index uint
		dir Direction
		val int
		err error
	} {
		{ id: "input", index: 1, val: 10,  dir: DirectionInput, err: nil},
		{ id: "output", index: 2, val: 20, dir: DirectionOutput, err: nil },
		{ id: "bad-index", index: 75, val: -1, dir: DirectionOutput, err: ErrOutOfRange },
	}

	v := NewVPIO[int]()
	for _, pin := range pins {
		p, err := v.Pin(pin.id, pin.index, pin.dir)
		assert.Equal(t, err, pin.err)
		if err == ErrOutOfRange {
			continue
		}

		assert.Equal(t, p.id, pin.id)
		assert.Equal(t, p.index, pin.index)
		if err != nil {
			continue
		}

		err =  v.Set(pin.index, pin.val)
		assert.NoError(t, err)
		val, err := v.Get(pin.index)
		assert.NoError(t, err)
		assert.Equal(t, val, pin.val)
	}

}

func TestVPIOFloat(t *testing.T) {
	pins := []struct{
		id string
		index uint
		dir Direction
		val float64
		err error
	} {
		{ id: "input", index: 1, val: 1.1,  dir: DirectionInput, err: nil},
		{ id: "output", index: 2, val: 1.2, dir: DirectionOutput, err: nil },
		{ id: "bad-index", index: 75, val: -1.0, dir: DirectionOutput, err: ErrOutOfRange },
	}

	v := NewVPIO[float64]()
	for _, pin := range pins {
		p, err := v.Pin(pin.id, pin.index, pin.dir)
		assert.Equal(t, err, pin.err)
		if err == ErrOutOfRange {
			continue
		}

		assert.Equal(t, p.id, pin.id)
		assert.Equal(t, p.index, pin.index)
		if err != nil {
			continue
		}

		err =  v.Set(pin.index, pin.val)
		assert.NoError(t, err)
		val, err := v.Get(pin.index)
		assert.NoError(t, err)
		assert.Equal(t, val, pin.val)
	}
}

func TestVPIONoTransactions(t *testing.T) {
	v := NewVPIO[float64]()
	v.Pin("output", 0, DirectionInput)
	v.Pin("input", 1, DirectionOutput)
	v.recording = false

	val := 1.2
	for i := 0; i < 2; i++ {
		val += float64(i)
		var j uint
		for j = 0; j < 2; j++ {
			err := v.Set(j, val)
			assert.NoError(t, err)
			value, err := v.Get(j)
			assert.NoError(t, err)
			assert.Equal(t, val, value)
		}
	}

	assert.Equal(t, 0, len(v.transactions))
}

func TestVPIOTransactions(t *testing.T) {
	v := NewVPIO[float64]()
	v.Pin("output", 0, DirectionInput)
	v.Pin("input", 1, DirectionOutput)
	v.recording = true

	val := 1.2
	for i := 0; i < 20; i++ {
		val += float64(i)
		var j uint
		for j = 0; j < 2; j++ {
			err := v.Set(j, val)
			assert.NoError(t, err)
			value, err := v.Get(j)
			assert.NoError(t, err)
			assert.Equal(t, val, value)
		}
	}
}
