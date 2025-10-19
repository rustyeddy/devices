package drivers

import (
	"log/slog"

	"github.com/rustyeddy/devices"
	"go.bug.st/serial"
)

type Serial struct {
	PortName string
	Baud     int

	*serial.Mode
	serial.Port
	mock bool
}

var (
	serialPorts map[string]*Serial
)

func init() {
	serialPorts = make(map[string]*Serial)
}

func GetSerial(port string) *Serial {
	if s, ex := serialPorts[port]; ex {
		return s
	}
	s, err := NewSerial(port, 115200)
	if err != nil {
		slog.Error("Serial port", "port", port, "error", err)
		return nil
	}
	return s
}

func (s *Serial) ID() string {
	return s.PortName
}

func NewSerial(port string, baud int) (s *Serial, err error) {
	s = &Serial{
		PortName: port,
		Baud:     baud,
	}

	if devices.IsMock() {
		return s, nil
	}
	s.Mode = &serial.Mode{
		BaudRate: baud,
	}
	return s, nil
}

func (s *Serial) Open() (err error) {
	s.Port, err = serial.Open(s.PortName, s.Mode)
	if err != nil {
		return err
	}

	serialPorts[s.ID()] = s
	return nil
}
