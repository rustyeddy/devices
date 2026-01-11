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
