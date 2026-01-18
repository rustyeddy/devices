package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rustyeddy/devices/devices/relay"
	"github.com/rustyeddy/devices/drivers"
)

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	// Factory (Linux GPIO via go-gpiocdev)
	f := drivers.NewGPIOCDevFactory()

	// Relay config
	cfg := relay.RelayConfig{
		Name:    "Relay1",
		Factory: f,
		Chip:    "gpiochip0", // optional; defaults to gpiochip0 if empty
		Offset:  18,          // GPIO line offset (change to match your wiring)
		Initial: false,       // initial relay state (false=off)
		// If your relay board is active-low and the device supports it,
		// you may have a field like Invert/ActiveLow. If so, set it here.
		// Invert: true,
	}

	r := relay.New(cfg)

	// Run device
	go func() {
		if err := r.Run(ctx); err != nil {
			log.Printf("relay stopped: %v", err)
		}
	}()

	log.Println("Relay example started (toggle every 2s). Ctrl-C to exit.")

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	on := false
	for {
		select {
		case <-ctx.Done():
			log.Println("shutting down")
			r.Close()
			return

		case <-ticker.C:
			on = !on
			select {
			case r.In() <- on:
			default:
				// avoid blocking if the command buffer is full
				log.Println("relay command dropped (buffer full)")
			}
		}
	}
}
