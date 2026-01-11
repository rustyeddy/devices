package main

import (
	"fmt"

	"github.com/rustyeddy/devices/gtu7"
)

func main() {
	g := gtu7.New("/dev/ttyS0")
	gpsQ := g.StartReading()

	for gps := range gpsQ {
		fmt.Printf("GPS %+v\n", gps)
	}
}
