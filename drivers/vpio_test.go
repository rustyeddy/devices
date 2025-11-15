package drivers

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVPIOBool(t *testing.T) {
	pins := []struct {
		id    string
		index int
		dir   PinOptions
		val   bool
		err   error
	}{
		{id: "input", index: 1, val: true, dir: PinInput, err: nil},
		{id: "output", index: 2, val: false, dir: PinOutput, err: nil},
		{id: "bad-index", index: 75, val: false, dir: PinOutput, err: ErrOutOfRange},
	}

	v := NewVPIO[bool]()
	for _, pin := range pins {
		p, err := v.SetPin(pin.id, pin.index, pin.dir)
		assert.Equal(t, err, pin.err)
		if err == ErrOutOfRange {
			continue
		}

		assert.Equal(t, p.ID(), pin.id)
		assert.Equal(t, p.Index(), pin.index)
		if err != nil {
			continue
		}

		err = v.Set(pin.index, pin.val)
		assert.NoError(t, err)
		val, err := v.Get(pin.index)
		assert.NoError(t, err)
		assert.Equal(t, val, pin.val)
	}

}

func TestVPIOInt(t *testing.T) {
	pins := []struct {
		id    string
		index int
		dir   PinOptions
		val   int
		err   error
	}{
		{id: "input", index: 1, val: 10, dir: PinInput, err: nil},
		{id: "output", index: 2, val: 20, dir: PinOutput, err: nil},
		{id: "bad-index", index: 75, val: -1, dir: PinOutput, err: ErrOutOfRange},
	}

	v := NewVPIO[int]()
	for _, pin := range pins {
		p, err := v.SetPin(pin.id, pin.index, pin.dir)
		assert.Equal(t, err, pin.err)
		if err == ErrOutOfRange {
			continue
		}

		assert.Equal(t, p.ID(), pin.id)
		assert.Equal(t, p.Index(), pin.index)
		if err != nil {
			continue
		}

		err = v.Set(pin.index, pin.val)
		assert.NoError(t, err)
		val, err := v.Get(pin.index)
		assert.NoError(t, err)
		assert.Equal(t, val, pin.val)
	}

}

func TestVPIOFloat(t *testing.T) {
	pins := []struct {
		id    string
		index int
		dir   PinOptions
		val   float64
		err   error
	}{
		{id: "input", index: 1, val: 1.1, dir: PinInput, err: nil},
		{id: "output", index: 2, val: 1.2, dir: PinOutput, err: nil},
		{id: "bad-index", index: 75, val: -1.0, dir: PinOutput, err: ErrOutOfRange},
	}

	v := NewVPIO[float64]()
	for _, pin := range pins {
		p, err := v.SetPin(pin.id, pin.index, pin.dir)
		assert.Equal(t, err, pin.err)
		if err == ErrOutOfRange {
			continue
		}

		assert.Equal(t, p.ID(), pin.id)
		assert.Equal(t, p.Index(), pin.index)
		if err != nil {
			continue
		}

		err = v.Set(pin.index, pin.val)
		assert.NoError(t, err)
		val, err := v.Get(pin.index)
		assert.NoError(t, err)
		assert.Equal(t, val, pin.val)
	}
}

func TestVPIONoTransactions(t *testing.T) {
	v := NewVPIO[float64]()
	v.SetPin("output", 0, PinInput)
	v.SetPin("input", 1, PinOutput)
	v.recording = false

	val := 1.2
	for i := 0; i < 2; i++ {
		val += float64(i)
		var j int
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
	v.SetPin("output", 0, PinInput)
	v.SetPin("input", 1, PinOutput)
	v.recording = true

	val := 1.2
	for i := 0; i < 20; i++ {
		val += float64(i)
		var j int
		for j = 0; j < 2; j++ {
			err := v.Set(j, val)
			assert.NoError(t, err)
			value, err := v.Get(j)
			assert.NoError(t, err)
			assert.Equal(t, val, value)
		}
	}
}
