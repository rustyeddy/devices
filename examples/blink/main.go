/*
Blink sets up pin 6 for an LED and goes into an endless
toggle mode.
*/

package main

import (
	"flag"
	"fmt"
	"log/slog"
	"time"

	"github.com/rustyeddy/devices/led"
)

var (
	useMQTT bool
	pinid   int
	mock    string
	count   int
)

func init() {
	flag.BoolVar(&useMQTT, "mqtt", false, "Use mqtt or a timer")
	flag.IntVar(&pinid, "pin", 6, "The GPIO pin the LED is attached to")
	flag.StringVar(&mock, "mock", "", "mock gpio and/or mqtt")
}

// TODO add a timerloop to this device (from otto)
func main() {
	flag.Parse()

	// Create the led, name it "led" and add a publish topic
	led, done := initLED("led", pinid)
	dotimer(led, 1*time.Second, done)
	fmt.Println("LED will blink every second")
	<-done
}

func initLED(name string, pin int) (*led.LED, chan any) {
	led := led.New(name, pin)
	done := make(chan any)
	return led, done
}

func dotimer(led *led.LED, period time.Duration, done chan any) {
	count = 0
	// led.TimerLoop(context.TODO(), period, func() error {
	// 	count++
	// 	return nil
	// })
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for {
			select {
			case t := <-ticker.C:
				count++
				slog.Info("led blink at ", "time", t, "count", count)

			case <-done:
				break
			}
		}
	}()

}
