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

	require.NotPanics(t, func() { b.Close() })
	require.NotPanics(t, func() { b.Close() }) // Safe to call multiple times
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

func TestBaseEmitBlockingWithSlowConsumer(t *testing.T) {
	t.Parallel()

	b := NewBase("dev", 1)

	// Fill buffer
	b.Emit(EventInfo, "fill", nil, nil)

	var emitDone sync.WaitGroup
	emitDone.Add(1)

	// This should block until we read from channel
	go func() {
		defer emitDone.Done()
		b.EmitBlocking(EventInfo, "blocked", nil, nil)
	}()

	// Give goroutine time to hit blocking send
	time.Sleep(10 * time.Millisecond)

	// Now consume to unblock
	ev1 := <-b.Events()
	assert.Equal(t, "fill", ev1.Msg)

	ev2 := <-b.Events()
	assert.Equal(t, "blocked", ev2.Msg)

	emitDone.Wait()
}

func TestBaseEmitBlockingAfterContextCancel(t *testing.T) {
	t.Parallel()

	b := NewBase("dev", 1)

	// Close Base's context
	b.Close()

	// EmitBlocking should return immediately without blocking
	done := make(chan struct{})
	go func() {
		b.EmitBlocking(EventInfo, "test", nil, nil)
		close(done)
	}()

	select {
	case <-done:
		// Success - returned immediately
	case <-time.After(100 * time.Millisecond):
		t.Fatal("EmitBlocking blocked after Close")
	}
}

func TestBaseEmitWithContextRaceCondition(t *testing.T) {
	t.Parallel()

	b := NewBase("dev", 1)

	// Fill buffer
	b.Emit(EventInfo, "fill", nil, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	// Start blocking emit
	wg.Add(1)
	emitDone := make(chan error, 1)
	go func() {
		defer wg.Done()
		emitDone <- b.EmitWithContext(ctx, EventInfo, "blocked", nil, nil)
	}()

	// Cancel immediately
	cancel()

	wg.Wait()

	// Should return context.Canceled
	err := <-emitDone
	require.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestBaseEmitWithNilMetadata(t *testing.T) {
	t.Parallel()

	b := NewBase("dev", 1)

	b.Emit(EventInfo, "test", nil, nil)

	ev := <-b.Events()
	assert.Nil(t, ev.Meta)
	assert.Nil(t, ev.Err)
}

func TestBaseEmitWithError(t *testing.T) {
	t.Parallel()

	b := NewBase("dev", 1)
	testErr := assert.AnError

	b.Emit(EventError, "error occurred", testErr, map[string]string{"code": "E001"})

	ev := <-b.Events()
	assert.Equal(t, EventError, ev.Kind)
	assert.Equal(t, "error occurred", ev.Msg)
	assert.Equal(t, testErr, ev.Err)
	assert.Equal(t, map[string]string{"code": "E001"}, ev.Meta)
}

func TestBaseEmptyName(t *testing.T) {
	t.Parallel()

	b := NewBase("", 1)

	assert.Equal(t, "", b.Name())

	b.Emit(EventInfo, "test", nil, nil)
	ev := <-b.Events()
	assert.Equal(t, "", ev.Device)
}

func TestBaseEventTimestamp(t *testing.T) {
	t.Parallel()

	b := NewBase("dev", 2)

	before := time.Now()
	b.Emit(EventInfo, "test", nil, nil)
	after := time.Now()

	ev := <-b.Events()
	assert.False(t, ev.Time.IsZero())
	assert.True(t, ev.Time.After(before) || ev.Time.Equal(before))
	assert.True(t, ev.Time.Before(after) || ev.Time.Equal(after))
}

func TestBaseAllEventKinds(t *testing.T) {
	t.Parallel()

	kinds := []EventKind{
		EventInfo,
		EventOpen,
		EventClose,
		EventError,
		EventRead,
		EventWrite,
	}

	b := NewBase("dev", len(kinds))

	for _, kind := range kinds {
		b.Emit(kind, "test", nil, nil)
	}

	received := make(map[EventKind]bool)
	for i := 0; i < len(kinds); i++ {
		ev := <-b.Events()
		received[ev.Kind] = true
	}

	for _, kind := range kinds {
		assert.True(t, received[kind], "EventKind %v not received", kind)
	}
}

// Additional corner case tests

func TestBaseZeroValueBehavior(t *testing.T) {
	t.Parallel()

	var b Base

	// Name should work on zero value
	require.NotPanics(t, func() { b.Name() })
	require.Equal(t, "", b.Name())

	// Close will panic on zero value because cancel func is nil
	// This is acceptable - users should always use NewBase()
	require.Panics(t, func() { b.Close() })

	b = NewBase("device1", 1)
	testErr := assert.AnError
	testMeta := map[string]string{"key1": "val1", "key2": "val2"}

	b.Emit(EventWrite, "write op", testErr, testMeta)

	ev := <-b.Events()
	require.Equal(t, "device1", ev.Device)
	require.Equal(t, EventWrite, ev.Kind)
	require.Equal(t, "write op", ev.Msg)
	require.Equal(t, testErr, ev.Err)
	require.Equal(t, testMeta, ev.Meta)
	require.False(t, ev.Time.IsZero())
}

func TestBaseEmitDropsWhenFull(t *testing.T) {
	t.Parallel()

	b := NewBase("dev", 2)

	// Fill buffer completely
	b.Emit(EventInfo, "msg1", nil, nil)
	b.Emit(EventInfo, "msg2", nil, nil)

	// This should be dropped
	b.Emit(EventInfo, "msg3", nil, nil)

	// Read both buffered events
	ev1 := <-b.Events()
	ev2 := <-b.Events()

	require.Equal(t, "msg1", ev1.Msg)
	require.Equal(t, "msg2", ev2.Msg)

	// No third event should be available
	select {
	case <-b.Events():
		t.Fatal("expected msg3 to be dropped")
	default:
	}
}

func TestBaseCloseIdempotent(t *testing.T) {
	t.Parallel()

	b := NewBase("dev", 1)

	// Close multiple times should be safe
	b.Close()
	b.Close()
	b.Close()

	_, ok := <-b.Events()
	require.False(t, ok)
}

func TestBaseProperShutdownPattern(t *testing.T) {
	t.Parallel()

	b := NewBase("dev", 100)

	var wg sync.WaitGroup

	// Start emitters
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				b.Emit(EventInfo, "concurrent", nil, nil)
			}
		}()
	}

	// Wait for all emitters to finish BEFORE closing
	wg.Wait()

	// Now safe to close
	b.Close()

	// Verify channel is closed
	for range b.Events() {
		// drain
	}
}

func TestBaseEmitBlockingDelivery(t *testing.T) {
	t.Parallel()

	b := NewBase("dev", 1)

	// EmitBlocking should deliver without dropping
	for i := 0; i < 5; i++ {
		delivered := make(chan struct{})
		go func(n int) {
			b.EmitBlocking(EventInfo, "msg", nil, map[string]string{"n": string(rune(n))})
			close(delivered)
		}(i)

		ev := <-b.Events()
		require.NotNil(t, ev.Meta)

		<-delivered
	}
}

func TestBaseEmitWithContextDelivery(t *testing.T) {
	t.Parallel()

	b := NewBase("dev", 1)
	ctx := context.Background()

	// Should deliver successfully
	err := b.EmitWithContext(ctx, EventRead, "data", nil, nil)
	require.NoError(t, err)

	ev := <-b.Events()
	require.Equal(t, EventRead, ev.Kind)
	require.Equal(t, "data", ev.Msg)
}

func TestBaseEventsReadOnly(t *testing.T) {
	t.Parallel()

	b := NewBase("dev", 1)
	ch := b.Events()

	// Verify channel is receive-only
	var _ <-chan Event = ch
}

func TestBaseNamePersistence(t *testing.T) {
	t.Parallel()

	name := "persistent-name"
	b := NewBase(name, 1)

	require.Equal(t, name, b.Name())

	b.Emit(EventInfo, "test", nil, nil)
	ev := <-b.Events()
	require.Equal(t, name, ev.Device)

	b.Close()
	require.Equal(t, name, b.Name())
}
