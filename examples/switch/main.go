package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/rustyeddy/devices/drivers"
	"github.com/warthog618/go-gpiocdev"
)

func main() {
	// Get the GPIO driver
	g := drivers.GetGPIO()
	defer func() {
		g.Close()
	}()

	done := make(chan bool, 0)
	go startSwitchHandler(g, done)
	go startSwitchToggler(g, done)

	<-done
}

func startSwitchToggler(g *drivers.GPIO, done chan bool) {
	on := false
	r := g.Pin("reader", 23, gpiocdev.AsOutput(1))
	for {
		if on {
			r.On()
			on = false
		} else {
			r.Off()
			on = true
		}
		time.Sleep(1 * time.Second)
	}
}

func startSwitchHandler(g *drivers.GPIO, done chan bool) {
	evtQ := make(chan gpiocdev.LineEvent)
	sw := g.Pin("switch", 24, gpiocdev.WithPullUp, gpiocdev.WithBothEdges, gpiocdev.WithEventHandler(func(evt gpiocdev.LineEvent) {
		evtQ <- evt
	}))

	for {
		select {
		case evt := <-evtQ:
			switch evt.Type {
			case gpiocdev.LineEventFallingEdge:
				slog.Info("GPIO failing edge", "pin", sw.Offset())
				fallthrough

			case gpiocdev.LineEventRisingEdge:
				slog.Info("GPIO raising edge", "pin", sw.Offset())
				v, err := sw.Get()
				if err != nil {
					slog.Error("Error getting input value: ", "error", err.Error())
					continue
				}
				fmt.Printf("val: %d\n", v)

			default:
				slog.Warn("Switch unknown event type ", "type", evt.Type)
			}

		case <-done:
			return
		}
	}
}
