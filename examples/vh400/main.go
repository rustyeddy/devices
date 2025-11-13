package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/rustyeddy/devices/vh400"
)

func main() {
	var err error
	pin := 0

	if len(os.Args) > 1 {
		pin, err = strconv.Atoi(os.Args[1])
		if err != nil {
			fmt.Printf("Bad argument %s - expected integers for adc\n", os.Args[1])
		}
	}

	readQ := make(<-chan float64)
	s, err := vh400.New("vh400", pin)
	if err != nil {
		panic(err)
	}
	readQ = s.ReadContinuous()

	for {
		select {
		case val := <-readQ:
			fmt.Printf("adc: %d: %5.2f\n", pin, val)
		}
	}
}
