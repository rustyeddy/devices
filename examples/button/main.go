package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rustyeddy/devices/devices/button"
	"github.com/rustyeddy/devices/drivers"
)

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	// Linux GPIO via go-gpiocdev
	f := drivers.NewGPIOCDevFactory()

	cfg := button.ButtonConfig{
		Name:     "Button1",
		Factory:  f,
		Chip:     "gpiochip0", // optional; defaults to gpiochip0 if empty
		Offset:   23,          // CHANGE THIS to match your wiring (GPIO line offset)
		Bias:     drivers.BiasPullUp,
		Edge:     drivers.EdgeBoth,
		Debounce: 30 * time.Millisecond,
	}

	b := button.NewButton(cfg)

	// Consume Events() to show lifecycle + edge notifications.
	go func() {
		for ev := range b.Events() {
			// Keep event logging compact, but show meta for edge events.
			if ev.Err != nil {
				log.Printf("[event] dev=%s kind=%s msg=%q err=%v meta=%v",
					ev.Device, ev.Kind, ev.Msg, ev.Err, ev.Meta)
				continue
			}
			if len(ev.Meta) > 0 {
				log.Printf("[event] dev=%s kind=%s msg=%q meta=%v",
					ev.Device, ev.Kind, ev.Msg, ev.Meta)
				continue
			}
			log.Printf("[event] dev=%s kind=%s msg=%q", ev.Device, ev.Kind, ev.Msg)
		}
	}()

	// Run device
	go func() {
		if err := b.Run(ctx); err != nil {
			log.Printf("button stopped: %v", err)
		}
	}()

	log.Println("Button example started. Ctrl-C to exit.")
	log.Println("Out() emits boolean state. Events() emits lifecycle + edge events.")
	log.Println("With BiasPullUp: released=true (high), pressed=false (low).")

	// Interpret Out() state changes as "pressed"/"released".
	// With pull-up bias, the line is typically HIGH when released, LOW when pressed.
	pressedWhenLow := (cfg.Bias == drivers.BiasPullUp)

	var last *bool
	for {
		select {
		case <-ctx.Done():
			log.Println("shutting down")
			// Run() defers closing Out() and calls Base.Close().
			// Calling Close() here is fine as an extra belt-and-suspenders.
			_ = b.Close()
			return

		case state, ok := <-b.Out():
			if !ok {
				log.Println("button output closed")
				return
			}

			// Drop duplicates (some setups may emit redundant states).
			if last != nil && *last == state {
				continue
			}
			last = ptr(state)

			action := interpretPressed(state, pressedWhenLow)
			fmt.Printf("button state=%v -> %s\n", state, action)
		}
	}
}

func interpretPressed(state bool, pressedWhenLow bool) string {
	// state=true means line high, state=false means line low
	if pressedWhenLow {
		if !state {
			return "PRESSED"
		}
		return "RELEASED"
	}
	// pressed when high (pull-down bias)
	if state {
		return "PRESSED"
	}
	return "RELEASED"
}

func ptr[T any](v T) *T { return &v }
