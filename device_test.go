package devices

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testDevice struct {
	name   string
	events chan Event
	out    chan bool
	in     chan bool
}

func (d *testDevice) Name() string              { return d.name }
func (d *testDevice) Run(context.Context) error { return nil }
func (d *testDevice) Events() <-chan Event      { return d.events }
func (d *testDevice) Out() <-chan bool          { return d.out }
func (d *testDevice) In() chan<- bool           { return d.in }

func TestDeviceInterfaces(t *testing.T) {
	t.Parallel()

	d := &testDevice{
		name:   "dev",
		events: make(chan Event, 1),
		out:    make(chan bool, 1),
		in:     make(chan bool, 1),
	}

	require.Implements(t, (*Device)(nil), d)
	require.Implements(t, (*Source[bool])(nil), d)
	require.Implements(t, (*Sink[bool])(nil), d)
	require.Implements(t, (*Duplex[bool])(nil), d)

	assert.Equal(t, "dev", d.Name())
	assert.Equal(t, chanID(d.out), chanID(d.Out()))
	assert.Equal(t, chanID(d.in), chanID(d.In()))
	assert.Equal(t, chanID(d.events), chanID(d.Events()))
}

func chanID(ch any) uintptr {
	return reflect.ValueOf(ch).Pointer()
}
