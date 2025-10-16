package main

import (
	"context"
	"time"

	"github.com/rustyeddy/devices/bme280"
)

func main() {
	// Set the BME i2c device and address Initialize the bme to use
	// the i2c bus
	bme := bme280.New("bme280", "/dev/i2c-1", 0x76)
	err := bme.Open()
	if err != nil {
		panic(err)
	}

	// start reading in a loop and publish the results via MQTT
	done := make(chan any)
	go bme.TimerLoop(context.TODO(), 5*time.Second, func() error {
		return nil
	})

	<-done
}
