package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rustyeddy/devices/devices/led"
	"github.com/rustyeddy/devices/drivers"
)

var (
	pin = 21
)

func init() {
	flag.IntVar(&pin, "pin", 21, "GPIO Pin to set")
}

func main() {
	flag.Parse()

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM)
	defer cancel()

	// Factory (Linux GPIO via go-gpiocdev)
	f := drivers.NewGPIOCDevFactory()

	// LED Config
	cfg := led.LEDConfig{
		Name:    "Led1",
		Factory: f,
		Chip:    "gpiochip0", // optional; defaults to gpiochip0 if empty
		Offset:  pin,         // GPIO line offset
		Initial: false,
	}
	l := led.New(cfg)

	// Run device
	go func() {
		if err := l.Run(ctx); err != nil {
			log.Printf("led stopped: %v", err)
		}
	}()

	log.Println("LED example started (toggle every 1s). Ctrl-C to exit.")

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	on := false
	for {
		select {
		case <-ctx.Done():
			log.Println("shutting down")
			l.Close()
			return

		case <-ticker.C:
			on = !on
			select {
			case l.In() <- on:
			default:
				// if the command buffer is full, skip this tick rather than blocking
				log.Println("led command dropped (buffer full)")
			}
		}
	}
}
