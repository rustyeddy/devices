package devices

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBaseDefaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		eventBu int
		wantCap int
	}{
		{name: "negative buffer uses default", eventBu: -1, wantCap: 16},
		{name: "zero buffer uses default", eventBu: 0, wantCap: 16},
		{name: "positive buffer preserved", eventBu: 8, wantCap: 8},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			b := NewBase("dev", tt.eventBu)
			assert.Equal(t, "dev", b.Name())
			assert.Equal(t, tt.wantCap, cap(b.Events()))
		})
	}
}

func TestBaseEmitNonBlocking(t *testing.T) {
	t.Parallel()

	b := NewBase("dev", 1)
	b.Emit(EventInfo, "first", nil, map[string]string{"k": "v"})
	b.Emit(EventInfo, "second", nil, nil)

	ev := <-b.Events()
	assert.Equal(t, "dev", ev.Device)
	assert.Equal(t, EventInfo, ev.Kind)
	assert.Equal(t, "first", ev.Msg)
	assert.False(t, ev.Time.IsZero())
	assert.Equal(t, map[string]string{"k": "v"}, ev.Meta)

	select {
	case <-b.Events():
		t.Fatal("expected second event to be dropped")
	default:
	}
}

func TestBaseEmitBlockingAndClose(t *testing.T) {
	t.Parallel()

	b := NewBase("dev", 1)
	b.EmitBlocking(EventOpen, "start", nil, nil)

	ev := <-b.Events()
	require.Equal(t, EventOpen, ev.Kind)
	require.Equal(t, "start", ev.Msg)

	require.NotPanics(t, b.CloseEvents)
	require.NotPanics(t, b.CloseEvents)
	_, ok := <-b.Events()
	assert.False(t, ok)
}
