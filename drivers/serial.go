package drivers

import (
	"fmt"
	"io"
)

// SerialConfig describes how to open a serial port.
type SerialConfig struct {
	Port string
	Baud int
}

// SerialPort is an opened serial port.
//
// Implementations should be safe to read from with bufio.Scanner.
type SerialPort interface {
	io.ReadWriteCloser
	String() string
}

// SerialFactory opens serial ports.
type SerialFactory interface {
	OpenSerial(cfg SerialConfig) (SerialPort, error)
}

func validateSerialConfig(cfg SerialConfig) error {
	if cfg.Port == "" {
		return fmt.Errorf("serial: Port is required")
	}
	if cfg.Baud <= 0 {
		return fmt.Errorf("serial: Baud must be > 0")
	}
	return nil
}
