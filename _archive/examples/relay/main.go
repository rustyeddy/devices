/*
Relay sets up pin 6 for a Relay (or LED) and connects to an MQTT
broker waiting for instructions to turn on or off the relay.
*/

package main

import (
	"embed"
)

//go:embed app
var content embed.FS

func main() {
}
