package main

// import (
// 	"fmt"
// 	"log"
// 	"time"

// 	"github.com/rustyeddy/devices/drivers"
// )

// func main() {
// 	fmt.Println("=== VPIO Example ===")
// 	vpioExample()

// 	// Uncomment when running on actual hardware
// 	// fmt.Println("\n=== GPIOCDev Example ===")
// 	// gpiocdevExample()
// }

// func vpioExample() {
// 	vpio := drivers.NewVPIO[bool]()
// 	vpio.StartRecording()

// 	pin5, err := vpio.SetPin("LED", 5, drivers.PinOutput)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	fmt.Printf("Pin ID: %s, Index: %d, Direction: %d\n",
// 		pin5.ID(), pin5.Index(), pin5.GetOptions())

// 	vpio.Set(5, true)
// 	val, _ := vpio.Get(5)
// 	fmt.Printf("Pin 5 value: %v\n", val)

// 	vpio.Set(5, false)
// 	val, _ = vpio.Get(5)
// 	fmt.Printf("Pin 5 value: %v\n", val)

// 	vpio.StopRecording()
// 	transactions := vpio.GetTransactions()
// 	fmt.Printf("\nRecorded %d transactions:\n", len(transactions))
// 	for i, tx := range transactions {
// 		fmt.Printf("  %d: Pin %d = %v at %s\n",
// 			i+1, tx.Index(), tx.Value(), tx.Time.Format(time.RFC3339Nano))
// 	}
// }

// func gpiocdevExample() {
// 	gpio := drivers.NewGPIOCDev("gpiochip0")
// 	defer gpio.Close()

// 	const (
// 		PinOutput    drivers.PinOptions = 1 << 1
// 		PinOutputLow drivers.PinOptions = 1 << 2
// 	)

// 	led, err := gpio.SetPin("LED", 17, PinOutput|PinOutputLow)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	fmt.Printf("Pin ID: %s, Index: %d, Direction: %d\n",
// 		led.ID(), led.Index(), led.Direction())

// 	for i := 0; i < 5; i++ {
// 		fmt.Printf("LED ON\n")
// 		led.On()
// 		time.Sleep(500 * time.Millisecond)

// 		fmt.Printf("LED OFF\n")
// 		led.Off()
// 		time.Sleep(500 * time.Millisecond)
// 	}

// 	for i := 0; i < 5; i++ {
// 		led.Toggle()
// 		val, _ := led.Get()
// 		fmt.Printf("LED toggled to: %d\n", val)
// 		time.Sleep(500 * time.Millisecond)
// 	}
// }

