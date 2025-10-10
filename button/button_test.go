package button

import (
	"sync"
	"testing"
	"time"

	"github.com/rustyeddy/devices"
)

var (
	gotit [2]bool
	wg    sync.WaitGroup
)

func TestButton(t *testing.T) {
	devices.Mock(true)
	done := make(chan any)

	b := New("button", 23)
	go b.EventLoop(done, b.ReadPub)

	wg.Add(2)
	b.MockHWInput(0)
	b.MockHWInput(1)

	wg.Wait()
	time.Sleep(10 * time.Millisecond)
	b.Close()
	done <- true

	if !gotit[0] || !gotit[1] {
		t.Errorf("failed to get 0 and 1 got (%t) and (%t)", gotit[0], gotit[1])
	}

}
