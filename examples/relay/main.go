/*
Relay sets up pin 6 for a Relay (or LED) and connects to an MQTT
broker waiting for instructions to turn on or off the relay.
*/

package main

import (
	"embed"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/rustyeddy/devices/drivers"
	"github.com/warthog618/go-gpiocdev"
)

//go:embed app
var content embed.FS

type relay struct {
	*drivers.DigitalPin
}

func main() {
	// Get the GPIO driver
	g := drivers.GetGPIO()
	defer func() {
		g.Close()
	}()

	// capture exit signals to ensure pin is reverted to input on exit.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(quit)

	r := g.Pin("relay", 6, gpiocdev.AsOutput(0))
	<-quit

	r.On()
	r.Off()

	slog.Info("Exiting relay")
}
