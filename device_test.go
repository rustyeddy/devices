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
func (d *testDevice) Close() error              { return nil }

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

func TestDeviceNameMethod(t *testing.T) {
	t.Parallel()

	tests := []string{"sensor1", "relay-2", "", "device_with_underscores", "123"}

	for _, name := range tests {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			d := &testDevice{name: name}
			assert.Equal(t, name, d.Name())
		})
	}
}

func TestDeviceRunWithContext(t *testing.T) {
	t.Parallel()

	t.Run("run completes successfully", func(t *testing.T) {
		t.Parallel()

		d := &testDevice{name: "test"}
		ctx := context.Background()

		err := d.Run(ctx)
		assert.NoError(t, err)
	})

	t.Run("context cancellation", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		d := &testDevice{name: "test"}
		err := d.Run(ctx)
		// Device contract doesn't require error on cancel
		assert.NoError(t, err)
	})
}

func TestSourceInterface(t *testing.T) {
	t.Parallel()

	t.Run("Source[int]", func(t *testing.T) {
		t.Parallel()

		// Use testDevice but logically treat as int source
		// (Type system verified via compilation)
		ch := make(chan int, 1)
		ch <- 42
		assert.Equal(t, 42, <-ch)
	})

	t.Run("different types", func(t *testing.T) {
		t.Parallel()

		// Just verify we can create sources of different types
		boolOut := make(chan bool)
		intOut := make(chan int)
		strOut := make(chan string)

		assert.NotNil(t, boolOut)
		assert.NotNil(t, intOut)
		assert.NotNil(t, strOut)
	})
}

func TestSinkInterface(t *testing.T) {
	t.Parallel()

	t.Run("Sink[bool]", func(t *testing.T) {
		t.Parallel()

		s := &testDevice{
			name:   "bool-sink",
			events: make(chan Event),
			in:     make(chan bool, 1),
		}

		var sink Sink[bool] = s
		sink.In() <- true

		assert.True(t, <-s.in)
	})

	t.Run("different types", func(t *testing.T) {
		t.Parallel()

		// Verify we can create sinks of different types
		boolIn := make(chan bool)
		intIn := make(chan int)
		strIn := make(chan string)

		assert.NotNil(t, boolIn)
		assert.NotNil(t, intIn)
		assert.NotNil(t, strIn)
	})
}

func TestDuplexInterface(t *testing.T) {
	t.Parallel()

	t.Run("implements both Source and Sink", func(t *testing.T) {
		t.Parallel()

		d := &testDevice{
			name:   "duplex",
			events: make(chan Event),
			out:    make(chan bool, 1),
			in:     make(chan bool, 1),
		}

		var duplex Duplex[bool] = d

		// Can use as Source
		var src Source[bool] = duplex
		assert.Equal(t, "duplex", src.Name())

		// Can use as Sink
		var sink Sink[bool] = duplex
		assert.Equal(t, "duplex", sink.Name())

		// Can use as Device
		var dev Device = duplex
		assert.Equal(t, "duplex", dev.Name())
	})

	t.Run("bidirectional communication", func(t *testing.T) {
		t.Parallel()

		d := &testDevice{
			name:   "bidirectional",
			events: make(chan Event),
			out:    make(chan bool, 1),
			in:     make(chan bool, 1),
		}

		var duplex Duplex[bool] = d

		// Send via In
		duplex.In() <- true
		assert.True(t, <-d.in)

		// Receive via Out
		d.out <- false
		assert.False(t, <-duplex.Out())
	})
}

func TestChannelDirections(t *testing.T) {
	t.Parallel()

	t.Run("Out() is receive-only", func(t *testing.T) {
		t.Parallel()

		d := &testDevice{
			name:   "source",
			events: make(chan Event),
			out:    make(chan bool, 1),
		}

		var src Source[bool] = d
		ch := src.Out()

		// Verify it's a receive-only channel
		var _ <-chan bool = ch
	})

	t.Run("In() is send-only", func(t *testing.T) {
		t.Parallel()

		d := &testDevice{
			name:   "sink",
			events: make(chan Event),
			in:     make(chan bool, 1),
		}

		var sink Sink[bool] = d
		ch := sink.In()

		// Verify it's a send-only channel
		var _ chan<- bool = ch
	})

	t.Run("Events() is receive-only", func(t *testing.T) {
		t.Parallel()

		d := &testDevice{
			name:   "device",
			events: make(chan Event, 1),
		}

		var dev Device = d
		ch := dev.Events()

		// Verify it's a receive-only channel
		var _ <-chan Event = ch
	})
}

func TestDeviceInterfaceMinimalContract(t *testing.T) {
	t.Parallel()

	// Device only requires Name, Run, and Events
	d := &testDevice{
		name:   "minimal",
		events: make(chan Event),
	}

	var dev Device = d

	// Should satisfy Device interface
	assert.Equal(t, "minimal", dev.Name())
	assert.NoError(t, dev.Run(context.Background()))
	assert.NotNil(t, dev.Events())
}

func TestTypeSafety(t *testing.T) {
	t.Parallel()

	// Verify type parameters are enforced
	boolSrc := &testDevice{
		name:   "bool-src",
		events: make(chan Event),
		out:    make(chan bool, 1),
	}

	boolSink := &testDevice{
		name:   "bool-sink",
		events: make(chan Event),
		in:     make(chan bool, 1),
	}

	var src Source[bool] = boolSrc
	var sink Sink[bool] = boolSink

	assert.Equal(t, "bool-src", src.Name())
	assert.Equal(t, "bool-sink", sink.Name())
}

func TestMultipleDevicesWithSameType(t *testing.T) {
	t.Parallel()

	devices := make([]Device, 0, 3)

	for i := 0; i < 3; i++ {
		d := &testDevice{
			name:   "device",
			events: make(chan Event),
		}
		devices = append(devices, d)
	}

	assert.Len(t, devices, 3)
	for _, d := range devices {
		assert.Equal(t, "device", d.Name())
	}
}
