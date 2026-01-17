//go:build linux && (arm || arm64)

package main

import "github.com/rustyeddy/devices/drivers"

func defaultFactory() drivers.OLEDFactory {
	return drivers.PeriphOLEDFactory{}
}
