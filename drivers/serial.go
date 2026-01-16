package drivers

import (
	"fmt"
	"io"
	"time"
)

// SerialConfig describes how to open a serial port.
type SerialConfig struct {
	Port string
	Baud int

	// SettleDelay is the time to wait after configuring the serial port
	// before returning. This allows the hardware to stabilize after changing
	// terminal settings. Some serial devices require a brief pause after
	// reconfiguration to avoid data corruption or initialization issues.
	//
	// If zero, a default of 10ms is used. To disable the delay entirely,
	// set to a negative value (e.g., -1).
	SettleDelay time.Duration
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
