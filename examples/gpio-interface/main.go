package main

import (
	"fmt"
	"log"
	"time"

	"github.com/rustyeddy/devices/drivers"
)

func main() {
	fmt.Println("=== VPIO Example ===")
	vpioExample()

	// Uncomment when running on actual hardware
	// fmt.Println("\n=== GPIOCDev Example ===")
	// gpiocdevExample()
}

func vpioExample() {
	vpio := drivers.NewVPIO[bool]()
	vpio.StartRecording()

	pin5, err := vpio.Pin("LED", 5, drivers.DirectionOutput)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Pin ID: %s, Index: %d, Direction: %d\n",
		pin5.ID(), pin5.Index(), pin5.Direction())

	vpio.Set(5, true)
	val, _ := vpio.Get(5)
	fmt.Printf("Pin 5 value: %v\n", val)

	vpio.Set(5, false)
	val, _ = vpio.Get(5)
	fmt.Printf("Pin 5 value: %v\n", val)

	vpio.StopRecording()
	transactions := vpio.GetTransactions()
	fmt.Printf("\nRecorded %d transactions:\n", len(transactions))
	for i, tx := range transactions {
		fmt.Printf("  %d: Pin %d = %v at %s\n",
			i+1, tx.Index, tx.Value, tx.Time.Format(time.RFC3339Nano))
	}
}

func gpiocdevExample() {
	gpio := drivers.NewGPIOCDev("gpiochip0")
	defer gpio.Close()

	const (
		PinOutput    drivers.PinOptions = 1 << 1
		PinOutputLow drivers.PinOptions = 1 << 2
	)

	led, err := gpio.Pin("LED", 17, PinOutput|PinOutputLow)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Pin ID: %s, Index: %d, Direction: %d\n",
		led.ID(), led.Index(), led.Direction())

	for i := 0; i < 5; i++ {
		fmt.Printf("LED ON\n")
		led.On()
		time.Sleep(500 * time.Millisecond)

		fmt.Printf("LED OFF\n")
		led.Off()
		time.Sleep(500 * time.Millisecond)
	}

	for i := 0; i < 5; i++ {
		led.Toggle()
		val, _ := led.Get()
		fmt.Printf("LED toggled to: %d\n", val)
		time.Sleep(500 * time.Millisecond)
	}
}































































































}	}		time.Sleep(500 * time.Millisecond)		fmt.Printf("LED toggled to: %d\n", val)		val, _ := led.Get()		led.Toggle()	for i := 0; i < 5; i++ {	// Or use toggle	}		time.Sleep(500 * time.Millisecond)		led.Off()		fmt.Printf("LED OFF\n")		time.Sleep(500 * time.Millisecond)		led.On()		fmt.Printf("LED ON\n")	for i := 0; i < 5; i++ {	// Blink the LED		led.ID(), led.Index(), led.Direction())	fmt.Printf("Pin ID: %s, Index: %d, Direction: %d\n",	}		log.Fatal(err)	if err != nil {	led, err := gpio.Pin("LED", 17, PinOutput|PinOutputLow)	// Configure GPIO 17 as output (LED)	)		PinOutputLow drivers.PinOptions = 1 << 2		PinOutput    drivers.PinOptions = 1 << 1	const (	// Define pin options (output, initially low)	defer gpio.Close()	gpio := drivers.NewGPIOCDev("gpiochip0")	// Create GPIO controller for Raspberry Pi Zerofunc gpiocdevExample() {}	}			i+1, tx.index, tx.value, tx.Time.Format(time.RFC3339Nano))		fmt.Printf("  %d: Pin %d = %v at %s\n", 	for i, tx := range transactions {	fmt.Printf("\nRecorded %d transactions:\n", len(transactions))	transactions := vpio.GetTransactions()	vpio.StopRecording()	// Stop recording and show transactions	fmt.Printf("Pin 5 value: %v\n", val)	val, _ = vpio.Get(5)	vpio.Set(5, false)	fmt.Printf("Pin 5 value: %v\n", val)	val, _ := vpio.Get(5)	vpio.Set(5, true)	// Set and get values		pin5.ID(), pin5.Index(), pin5.Direction())	fmt.Printf("Pin ID: %s, Index: %d, Direction: %d\n", 	}		log.Fatal(err)	if err != nil {	pin5, err := vpio.Pin("LED", 5, drivers.DirectionOutput)	// Configure pin 5 as output	vpio.StartRecording()	// Start recording transactions		vpio := drivers.NewVPIO[bool]()	// Create a virtual GPIO with boolean valuesfunc vpioExample() {}	// gpiocdevExample()	// fmt.Println("\n=== GPIOCDev Example ===")	// Uncomment when running on actual hardware	// Example 2: Using GPIOCDev (real hardware)	vpioExample()	fmt.Println("=== VPIO Example ===")	// Example 1: Using VPIO (virtual/mock GPIO)func main() {)	"github.com/rustyeddy/devices/drivers"	"time"	"log"	"fmt"import (package main