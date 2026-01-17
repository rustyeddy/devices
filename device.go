package devices

import "context"

// Device is the minimal lifecycle contract.
// A device owns its goroutines and channels.
//
// Lifecycle:
//   - Create device with constructor (e.g., NewSensor)
//   - Call Run(ctx) in a goroutine
//   - Device runs until ctx is canceled or fatal error
//   - Device closes all output channels before Run returns
//   - Close device to clean up (e.g., device.Close())
//
// Channel semantics:
//   - Events() returns a non-nil receive-only channel
//   - Device must not close Events channel until Run returns
//   - Consumers should stop reading Events when Run returns
type Device interface {
	Name() string

	// Run blocks until ctx is canceled or a fatal error occurs.
	// The device must close all its output channels (Events, Out, etc.) before returning.
	// Implementations should respect ctx cancellation and return promptly.
	Run(ctx context.Context) error

	// Events emits lifecycle, error, and edge notifications.
	// Returns a non-nil receive-only channel.
	Events() <-chan Event
}

// Source emits typed values (sensors, buttons, GPS fixes, etc.).
//
// Example implementing a temperature sensor:
//
//	type TempSensor struct {
//	    *Base
//	    out chan float64
//	}
//
//	func (s *TempSensor) Out() <-chan float64 { return s.out }
//
//	func (s *TempSensor) Run(ctx context.Context) error {
//	    defer close(s.out) // Close output channel before returning
//	    for {
//	        select {
//	        case <-ctx.Done():
//	            return ctx.Err()
//	        case <-time.After(time.Second):
//	            temp := readTemperature()
//	            s.out <- temp
//	        }
//	    }
//	}
//
// Channel semantics:
//   - Out() returns a non-nil receive-only channel
//   - Device closes the Out() channel when Run returns
//   - Type parameter T can be any type (primitives, structs, pointers)
type Source[T any] interface {
	Device
	Out() <-chan T
}

// Sink consumes typed commands (relay, motor, LED, PWM, etc.).
//
// Example implementing a relay controller:
//
//	type Relay struct {
//	    *Base
//	    in chan bool
//	}
//
//	func (r *Relay) In() chan<- bool { return r.in }
//
//	func (r *Relay) Run(ctx context.Context) error {
//	    for {
//	        select {
//	        case <-ctx.Done():
//	            return ctx.Err()
//	        case state := <-r.in:
//	            setRelayState(state)
//	        }
//	    }
//	}
//
// Channel semantics:
//   - In() returns a non-nil send-only channel
//   - Device consumes from In() in its Run loop
//   - Senders should stop writing when Run exits
//   - Type parameter T can be any type (primitives, structs, pointers)
type Sink[T any] interface {
	Device
	In() chan<- T
}

// Duplex is both a Source and a Sink.
//
// Example implementing a bidirectional serial port:
//
//	type SerialPort struct {
//	    *Base
//	    in  chan []byte
//	    out chan []byte
//	}
//
//	func (s *SerialPort) In() chan<- []byte  { return s.in }
//	func (s *SerialPort) Out() <-chan []byte { return s.out }
//
//	func (s *SerialPort) Run(ctx context.Context) error {
//	    defer close(s.out)
//	    for {
//	        select {
//	        case <-ctx.Done():
//	            return ctx.Err()
//	        case data := <-s.in:
//	            writeSerial(data)
//	        case data := <-readSerial():
//	            s.out <- data
//	        }
//	    }
//	}
//
// Channel semantics:
//   - Combines all semantics from Source[T] and Sink[T]
//   - In() and Out() must both be non-nil
//   - Device closes Out() channel when Run returns
type Duplex[T any] interface {
	Source[T]
	Sink[T]
}

func Itoa(n int) string {
	if n == 0 {
		return "0"
	}
	sign := ""
	if n < 0 {
		sign = "-"
		n = -n
	}
	var buf [12]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + (n % 10))
		n /= 10
	}
	return sign + string(buf[i:])
}
