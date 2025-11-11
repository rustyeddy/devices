/*
Blink sets up pin 6 for an LED and goes into an endless
toggle mode.
*/

package main

import (
	"log/slog"

	"encoding/json"

	"github.com/rustyeddy/devices/drivers"
)

var gpioStr = `
{
    "chipname":"gpiochip4",
    "pins": {
        "6": {
            "name": "led",
            "offset": 6,
            "value": 0,
            "mode": 0
        }
    }
}
`

func main() {
	var g *drivers.GPIOCDev
	if err := json.Unmarshal([]byte(gpioStr), &g); err != nil {
		slog.Error(err.Error())
		return
	}

	defer func() {
		if g != nil {
			g.Close()
		}
	}()

	// TODO
	// led, _ := g.Pin("led", 6, drivers.PinOutput)
	// for {
	// 	led.Toggle()
	// 	time.Sleep(1 * time.Second)
	// }
}
