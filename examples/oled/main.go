package main

import (
	"context"

	"github.com/rustyeddy/devices/display"
)

func main() {
	oled := display.NewOLED(display.OLEDConfig{
		Name:    "status",
		Factory: defaultFactory(),
		Bus:     "1",
		Addr:    0x3c,
		Width:   128,
		Height:  64,
	})

	ctx := context.Background()
	go oled.Run(ctx)

	oled.In() <- display.OLEDCommand{Type: display.CmdClear}
	oled.In() <- display.OLEDCommand{Type: display.CmdText, X: 0, Y: 12, Text: "Hello!"}
	oled.In() <- display.OLEDCommand{Type: display.CmdFlush}
}
