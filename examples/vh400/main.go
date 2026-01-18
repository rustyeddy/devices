package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rustyeddy/devices/devices/vh400"
	"github.com/rustyeddy/devices/drivers"
)

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer cancel()

	// ADS1115 factory (periph.io). Note: implemented for linux+arm/arm64 in this repo.
	f := drivers.PeriphADCFactory{}

	cfg := vh400.VH400Config{
		Name:        "VH400",
		Factory:     f,
		Bus:         "1",  // common on Raspberry Pi; "" may also work depending on periph setup
		Addr:        0x48, // typical ADS1115 address
		Channel:     0,    // ADS1115 single-ended A0 (0-3)
		Interval:    2 * time.Second,
		EmitInitial: true,
		Buf:         16,
	}

	s := vh400.NewVH400(cfg)

	go func() {
		if err := s.Run(ctx); err != nil {
			log.Printf("vh400 stopped: %v", err)
		}
	}()

	log.Println("VH400 example started. Ctrl-C to exit.")
	log.Println("Reading VWC (%) from ADS1115...")

	for {
		select {
		case <-ctx.Done():
			log.Println("shutting down")
			s.Close()
			return

		case vwc, ok := <-s.Out():
			if !ok {
				log.Println("vh400 output closed")
				return
			}
			log.Printf("VWC: %.2f%%", vwc)
		}
	}
}
