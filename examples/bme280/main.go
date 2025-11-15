package main

import (
	"fmt"

	"github.com/rustyeddy/devices/env"
)

func main() {
	// Set the BME i2c device and address Initialize the bme to use
	// the i2c bus
	bme, err := env.New("bme280", "/dev/i2c-1", 0x76)
	if err != nil {
		panic(err)
	}

	// Open the device to prepare it for usage
	err = bme.Open()
	if err != nil {
		panic(err)
	}

	val, err := bme.Get()
	if err != nil {
		panic(err)
	}
	fmt.Printf("BME values %+v\n", val)

}
