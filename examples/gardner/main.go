package main

// import "github.com/rustyeddy/devices/button"
// import "github.com/rustyeddy/devices/led"
//import "github.com/rustyeddy/devices/relay"

import "github.com/rustyeddy/devices/drivers"
import "github.com/rustyeddy/devices/vh400"
import "github.com/rustyeddy/devices/bme280"
import "github.com/rustyeddy/devices"

type controller struct {
	green, red	drivers.DigitalPin
	on, off		drivers.DigitalPin
	pump		drivers.DigitalPin
	soil		vh400.VH400
	env			bme280.BME280
}

const (
	lowThreshold = 1.0
	highThreshold = 2.0
)

func main() {
	green, red := ledInit()
	on, off := buttonInit()
	pump := pumpInit()
	soil := soilInit()
	env := envInit()

	done := make(chan any)
	ctl := &controller{green, red, on, off, pump, soil, env}
	go ctl.Start()
	<-done
	ctl.Stop()
}

func (c *controller) Start() {

	// set red light on to indicate the unit is operational
	c.red.Set(1)

	// subscribe to published 
	c.Subscribe("soil", func(val float64) {
		ison := c.pump.Get()
		switch val {
		case val < lowThreshold && !ison:
			c.pump.Set(true)
			c.green.Set(true)

		case val > highThreshold && ison:
			c.pump.Set(false)
			c.green.Set(false)
		}
		// stats.Save("pump", val)
	})

	c.Subscribe("env", func(val *bme280.Environment) {
		// TODO set alarms if any value is too high or too low
		// stats.Save("env", val)
	})

	on.EventHandler = func(val bool) {
		c.Publish("pump", "on")
	}

	off.EventHandler = func(val bool) {
		c.Publish("pump", "off")
	}

	go c.pump.TimerLoop.Start()
	go c.env.TimerLoop.Start()

	<-done
}

func (c controller) Stop() {
	// Stop timers

	// Stop all devices and have them unsubscribe

	red.Set(0)
}


// LEDs are straight GPIO Outputs that accepts false/0 == off,
// true/1 == on 
func ledInit() (*devices.LED, *devices.LED) {
	green := digital.Pin("green", 8, drivers.PinOptionOutput)
	green.Subscribe("green", func(val bool) {
		green.Set(val)
	})

	red := digital.Pin("red", 9, drivers.PinOptionOutput)
	red.Subscribe("red", func(val bool) {
		red.Set(val)
	})
	return green, red
}

// Buttons are straight GPIO Inputs where incoming false/0 is not
// pressed, true/1 is pressed. The buttons also support events,
// when button is pressed and/or released an event is sent to
// the buttons event handler
func buttonInit() (*devices.Button, *devices.Button) {
	on := digital.Pin("on", 4, PinOptionInput)
	on.EventHandler = func(val bool) {
		on.Publish("on", val)
	}

	off := digital.Pin("off", 5, PinOptionInput)
	off.EventHandler = func(val bool) {
		off.Publish("off", val)
	}
	return on, off
}

// Pump is a PWM GPIO where the input can range from 0.0 - 1.0
// determining the PWM duty cycle.
func pumpInit() *devices.Pump {
	pump := digital.Pin("pump", 11, PinOptionOutput)
	pump.Subscribe("pump", func(val float64) {
		if val > 1.0 || val < 0.0 {
			errQ <- fmt.Printf("pump error: value out of range %5.2f", val)
			return
		}
		pump.Set(val)
	})
}

// VH400 is a meter representing voltage ranging from 0.0 to 3.3v.
// The voltage meter uses an ADC with the ads1112 chip.
func soilInit() *devices.VH400 {
	soil := vh400.New("soil", 22)
	soil.TimerLoop = new.Timer(15 * time.Second, func(t time.Time) {
		val, err := soil.Get()
		if err != nil {
			errQ <- err 
		}
		soil.Publish("soil", val)
	})
	return soil
}

// bme280 is an i2c device that can periodically read the triple
// Temperature, Humidity and Pressure
func envInit() *devices.bme280 {
	env := bme280.New("env", "/dev/i2c-1", 0x4c)
	// val = { temp, humidity, pressure }
	env.TimerLoop = new.Timer(15 * time.Second, func(t time.Time) {
		val, err := env.Get()
		if err != nil {
			errQ <- err 
		}
		env.Publish("env", val)
	})
	return env
}

