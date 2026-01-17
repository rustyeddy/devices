package devices

import (
	"context"
	"sync"
	"testing"
	"time"

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

	require.NotPanics(t, b.Close)
	require.NotPanics(t, b.Close) // Safe to call multiple times
	_, ok := <-b.Events()
	assert.False(t, ok)
}

func TestBaseEmitAfterClose(t *testing.T) {
	t.Parallel()

	b := NewBase("dev", 1)
	b.Close()

	// All emit variants should be safe after close
	require.NotPanics(t, func() { b.Emit(EventInfo, "dropped", nil, nil) })
	require.NotPanics(t, func() { b.EmitBlocking(EventInfo, "dropped", nil, nil) })
}

func TestBaseConcurrentEmit(t *testing.T) {
	t.Parallel()

	b := NewBase("dev", 100)
	const goroutines = 10
	const emitsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < emitsPerGoroutine; j++ {
				b.Emit(EventInfo, "concurrent", nil, nil)
			}
		}()
	}

	wg.Wait()
	b.Close()

	count := 0
	for range b.Events() {
		count++
	}

	// Should receive many events (not necessarily all due to non-blocking drops)
	assert.Greater(t, count, 0)
}

func TestBaseEventsSameChannel(t *testing.T) {
	t.Parallel()

	b := NewBase("dev", 1)
	ch1 := b.Events()
	ch2 := b.Events()

	// Channels are compared by value, not pointer, but we want to ensure they're the same channel
	assert.Equal(t, ch1, ch2, "Events() should return the same channel")
}

func TestBaseEmitWithContext(t *testing.T) {
	t.Parallel()

	t.Run("successful delivery", func(t *testing.T) {
		t.Parallel()

		b := NewBase("dev", 1)
		ctx := context.Background()

		err := b.EmitWithContext(ctx, EventInfo, "test", nil, nil)
		require.NoError(t, err)

		ev := <-b.Events()
		assert.Equal(t, "test", ev.Msg)
	})

	t.Run("context canceled", func(t *testing.T) {
		t.Parallel()

		b := NewBase("dev", 1)
		// Fill buffer so next send would block
		b.Emit(EventInfo, "fill", nil, nil)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := b.EmitWithContext(ctx, EventInfo, "test", nil, nil)
		require.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	})

	t.Run("context timeout", func(t *testing.T) {
		t.Parallel()

		b := NewBase("dev", 1)
		// Fill buffer
		b.Emit(EventInfo, "fill", nil, nil)

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		err := b.EmitWithContext(ctx, EventInfo, "blocked", nil, nil)
		require.Error(t, err)
		assert.Equal(t, context.DeadlineExceeded, err)
	})

	t.Run("after close", func(t *testing.T) {
		t.Parallel()

		b := NewBase("dev", 1)
		b.Close()

		err := b.EmitWithContext(context.Background(), EventInfo, "test", nil, nil)
		require.NoError(t, err, "EmitWithContext after Close should return nil")
	})
}
