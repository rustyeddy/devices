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
	g := drivers.GetGPIO[bool]()
	defer func() {
		g.Close()
	}()

	done := make(chan bool, 0)
	go startSwitchHandler(g, done)
	go startSwitchToggler(g, done)

	<-done
}

func startSwitchToggler(g drivers.GPIO[bool], done chan bool) {
	on := false
	r, err := g.SetPin("reader", 23, drivers.PinOutput)
	if err != nil {
		panic(err)
	}
	for {
		if on {
			r.Set(on)
			on = false
		} else {
			r.Set(on)
			on = true
		}
		time.Sleep(1 * time.Second)
	}
}

func startSwitchHandler(g drivers.GPIO[bool], done chan bool) {
	evtQ := make(chan gpiocdev.LineEvent)
	sw, err := g.SetPin("switch", 24, drivers.PinPullUp, drivers.PinBothEdges)
	// gpiocdev.WithEventHandler(func(evt gpiocdev.LineEvent) {
	//	evtQ <- evt
	//}))
	if err != nil {
		panic(err)
	}

	for {
		select {
		case evt := <-evtQ:
			switch evt.Type {
			case gpiocdev.LineEventFallingEdge:
				slog.Info("GPIO failing edge", "pin", sw.Index())
				fallthrough

			case gpiocdev.LineEventRisingEdge:
				slog.Info("GPIO raising edge", "pin", sw.Index())
				v, err := sw.Get()
				if err != nil {
					slog.Error("Error getting input value: ", "error", err.Error())
					continue
				}
				fmt.Printf("val: %t\n", v)

			default:
				slog.Warn("Switch unknown event type ", "type", evt.Type)
			}

		case <-done:
			return
		}
	}
}
