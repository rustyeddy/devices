//go:build !linux

package drivers

import "fmt"

// LinuxSerialFactory is available only on Linux.
type LinuxSerialFactory struct{}

func (LinuxSerialFactory) OpenSerial(cfg SerialConfig) (SerialPort, error) {
	return nil, fmt.Errorf("serial: unsupported platform")
}
