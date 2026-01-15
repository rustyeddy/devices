package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rustyeddy/devices/bme280"
)

func main() {
	// Root context canceled on SIGINT / SIGTERM
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\nshutting down...")
		cancel()
	}()

	dev := bme280.New(bme280.Config{
		Name:        "env",
		Bus:         "",   // default I²C bus (periph: first available)
		Addr:        0x76, // change to 0x77 if needed
		Interval:    1 * time.Second,
		EmitInitial: true,
		DropOnFull:  true,
	})

	// Run device in background
	errCh := make(chan error, 1)
	go func() {
		errCh <- dev.Run(ctx)
	}()

	fmt.Println("BME280 polling started (Ctrl-C to stop)")

	for {
		select {
		case v, ok := <-dev.Out():
			if !ok {
				// device stopped
				return
			}

			fmt.Printf(
				"Temp: %.2f °C | Humidity: %.1f %%RH | Pressure: %.0f Pa\n",
				v.Temperature,
				v.Humidity,
				v.Pressure,
			)

		case err := <-errCh:
			if err != nil {
				fmt.Println("device error:", err)
			}
			return
		}
	}
}
