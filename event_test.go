package devices

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEventKindConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, EventKind("open"), EventOpen)
	assert.Equal(t, EventKind("close"), EventClose)
	assert.Equal(t, EventKind("error"), EventError)
	assert.Equal(t, EventKind("edge"), EventEdge)
	assert.Equal(t, EventKind("info"), EventInfo)
	assert.Equal(t, EventKind("read"), EventRead)
	assert.Equal(t, EventKind("write"), EventWrite)
}

func TestEventFields(t *testing.T) {
	t.Parallel()

	when := time.Unix(1, 2)
	ev := Event{
		Device: "dev",
		Kind:   EventEdge,
		Time:   when,
		Msg:    "msg",
		Err:    assert.AnError,
		Meta:   map[string]string{"k": "v"},
	}

	assert.Equal(t, "dev", ev.Device)
	assert.Equal(t, EventEdge, ev.Kind)
	assert.Equal(t, when, ev.Time)
	assert.Equal(t, "msg", ev.Msg)
	assert.Equal(t, assert.AnError, ev.Err)
	assert.Equal(t, map[string]string{"k": "v"}, ev.Meta)
}

func TestEventZeroValue(t *testing.T) {
	t.Parallel()

	var ev Event

	assert.Empty(t, ev.Device)
	assert.Empty(t, ev.Kind)
	assert.True(t, ev.Time.IsZero())
	assert.Empty(t, ev.Msg)
	assert.Nil(t, ev.Err)
	assert.Nil(t, ev.Meta)
}

func TestEventMinimalFields(t *testing.T) {
	t.Parallel()

	// Event with only required fields
	ev := Event{
		Device: "sensor1",
		Kind:   EventOpen,
		Time:   time.Now(),
	}

	assert.Equal(t, "sensor1", ev.Device)
	assert.Equal(t, EventOpen, ev.Kind)
	assert.False(t, ev.Time.IsZero())
	assert.Empty(t, ev.Msg)
	assert.Nil(t, ev.Err)
	assert.Nil(t, ev.Meta)
}

func TestEventAllKinds(t *testing.T) {
	t.Parallel()

	kinds := []EventKind{
		EventOpen,
		EventClose,
		EventError,
		EventEdge,
		EventInfo,
		EventRead,
		EventWrite,
	}

	for _, kind := range kinds {
		kind := kind
		t.Run(string(kind), func(t *testing.T) {
			t.Parallel()

			ev := Event{
				Device: "test",
				Kind:   kind,
				Time:   time.Now(),
			}

			assert.Equal(t, kind, ev.Kind)
		})
	}
}

func TestEventWithError(t *testing.T) {
	t.Parallel()

	t.Run("with error and message", func(t *testing.T) {
		t.Parallel()

		ev := Event{
			Device: "sensor",
			Kind:   EventError,
			Time:   time.Now(),
			Msg:    "read failed",
			Err:    assert.AnError,
		}

		assert.Equal(t, EventError, ev.Kind)
		assert.Equal(t, "read failed", ev.Msg)
		assert.Error(t, ev.Err)
	})

	t.Run("error kind without error field", func(t *testing.T) {
		t.Parallel()

		// Valid - Kind is just a label, Err is optional
		ev := Event{
			Device: "sensor",
			Kind:   EventError,
			Time:   time.Now(),
			Msg:    "timeout",
		}

		assert.Equal(t, EventError, ev.Kind)
		assert.Nil(t, ev.Err)
	})
}

func TestEventMetadata(t *testing.T) {
	t.Parallel()

	t.Run("nil metadata", func(t *testing.T) {
		t.Parallel()

		ev := Event{
			Device: "dev",
			Kind:   EventInfo,
			Time:   time.Now(),
			Meta:   nil,
		}

		assert.Nil(t, ev.Meta)
	})

	t.Run("empty metadata", func(t *testing.T) {
		t.Parallel()

		ev := Event{
			Device: "dev",
			Kind:   EventInfo,
			Time:   time.Now(),
			Meta:   make(map[string]string),
		}

		assert.NotNil(t, ev.Meta)
		assert.Empty(t, ev.Meta)
	})

	t.Run("metadata with multiple entries", func(t *testing.T) {
		t.Parallel()

		ev := Event{
			Device: "sensor",
			Kind:   EventEdge,
			Time:   time.Now(),
			Meta: map[string]string{
				"pin":   "GPIO17",
				"state": "high",
				"level": "1",
			},
		}

		assert.Len(t, ev.Meta, 3)
		assert.Equal(t, "GPIO17", ev.Meta["pin"])
		assert.Equal(t, "high", ev.Meta["state"])
		assert.Equal(t, "1", ev.Meta["level"])
	})
}

func TestEventIndependence(t *testing.T) {
	t.Parallel()

	// Verify events don't share state
	ev1 := Event{
		Device: "dev1",
		Kind:   EventOpen,
		Time:   time.Now(),
		Meta:   map[string]string{"id": "1"},
	}

	ev2 := Event{
		Device: "dev2",
		Kind:   EventClose,
		Time:   time.Now(),
		Meta:   map[string]string{"id": "2"},
	}

	// Modify ev1.Meta
	ev1.Meta["modified"] = "true"

	// ev2 should be unaffected
	assert.NotEqual(t, ev1.Device, ev2.Device)
	assert.NotEqual(t, ev1.Kind, ev2.Kind)
	assert.NotContains(t, ev2.Meta, "modified")
	assert.Equal(t, "2", ev2.Meta["id"])
}

func TestEventEmptyStrings(t *testing.T) {
	t.Parallel()

	// Empty strings are valid
	ev := Event{
		Device: "",
		Kind:   EventKind(""),
		Time:   time.Time{},
		Msg:    "",
	}

	assert.Empty(t, ev.Device)
	assert.Empty(t, ev.Kind)
	assert.Empty(t, ev.Msg)
	assert.True(t, ev.Time.IsZero())
}

func TestEventTimestampBehavior(t *testing.T) {
	t.Parallel()

	t.Run("zero time", func(t *testing.T) {
		t.Parallel()

		ev := Event{
			Device: "dev",
			Kind:   EventInfo,
			Time:   time.Time{},
		}

		assert.True(t, ev.Time.IsZero())
	})

	t.Run("specific timestamp", func(t *testing.T) {
		t.Parallel()

		ts := time.Date(2026, 1, 17, 12, 0, 0, 0, time.UTC)
		ev := Event{
			Device: "dev",
			Kind:   EventInfo,
			Time:   ts,
		}

		assert.Equal(t, ts, ev.Time)
		assert.Equal(t, 2026, ev.Time.Year())
	})
}

func TestEventKindAsString(t *testing.T) {
	t.Parallel()

	// EventKind is a string, so it can be used directly
	kinds := map[EventKind]string{
		EventOpen:  "open",
		EventClose: "close",
		EventError: "error",
		EventEdge:  "edge",
		EventInfo:  "info",
		EventRead:  "read",
		EventWrite: "write",
	}

	for kind, expected := range kinds {
		assert.Equal(t, expected, string(kind))
	}
}
