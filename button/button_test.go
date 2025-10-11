package button

import (
	"fmt"
	"testing"
	"time"

	"github.com/rustyeddy/devices"
	"github.com/stretchr/testify/assert"
	"github.com/warthog618/go-gpiocdev"
)

func init() {
	devices.Mock(true)
}
func TestNewButton_DefaultOptions(t *testing.T) {
	btn := New("testbtn", 5)
	assert.NotNil(t, btn)
	assert.NotNil(t, btn.Device)
	assert.NotNil(t, btn.DigitalPin)
	assert.Equal(t, "testbtn", btn.Device.Name)
	assert.NotNil(t, btn.EvtQ)
}

func TestNewButton_CustomOptions(t *testing.T) {
	opt := gpiocdev.WithPullDown
	btn := New("custombtn", 7, opt)
	assert.NotNil(t, btn)
	assert.NotNil(t, btn.DigitalPin)
}

func TestButton_ReadPub_Success(t *testing.T) {
	btn := New("readpubbtn", 8)
	btn.Device.Get = func() (int, error) { return 1, nil }
	assert.NotPanics(t, func() { btn.ReadPub() })
}

func TestButton_ReadPub_Error(t *testing.T) {
	btn := New("errbtn", 9)
	btn.Device.Get = func() (int, error) { return 0, fmt.Errorf("fail") }
	assert.NotPanics(t, func() { btn.ReadPub() })
}

func TestButton_EventHandler(t *testing.T) {
	btn := New("evtbtn", 10)
	event := gpiocdev.LineEvent{Offset: 10, Type: gpiocdev.LineEventRisingEdge}
	go func() { btn.EvtQ <- event }()
	select {
	case e := <-btn.EvtQ:
		assert.Equal(t, event.Offset, e.Offset)
		assert.Equal(t, event.Type, e.Type)
	case <-time.After(100 * time.Millisecond):
		t.Error("event not received")
	}
}
