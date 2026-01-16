//go:build linux

package drivers

import (
	"fmt"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

// LinuxSerialFactory opens a configured serial port on Linux.
//
// It uses termios to set 8N1 and the requested baud rate.
type LinuxSerialFactory struct{}

func (LinuxSerialFactory) OpenSerial(cfg SerialConfig) (SerialPort, error) {
	if err := validateSerialConfig(cfg); err != nil {
		return nil, err
	}

	// Open non-blocking first, configure termios, then clear O_NONBLOCK.
	fd, err := unix.Open(cfg.Port, unix.O_RDWR|unix.O_NOCTTY|unix.O_NONBLOCK, 0)
	if err != nil {
		return nil, fmt.Errorf("serial: open %s: %w", cfg.Port, err)
	}

	// Ensure we don't leak the fd if configuration fails.
	ok := false
	defer func() {
		if !ok {
			_ = unix.Close(fd)
		}
	}()

	if err := configureTermios(fd, cfg.Baud); err != nil {
		return nil, err
	}

	// Clear O_NONBLOCK so reads behave as expected.
	if err := unix.SetNonblock(fd, false); err != nil {
		return nil, fmt.Errorf("serial: set blocking %s: %w", cfg.Port, err)
	}

	f := os.NewFile(uintptr(fd), cfg.Port)
	if f == nil {
		return nil, fmt.Errorf("serial: os.NewFile returned nil for %s", cfg.Port)
	}

	ok = true
	return &linuxSerialPort{file: f, port: cfg.Port, baud: cfg.Baud}, nil
}

type linuxSerialPort struct {
	file *os.File
	port string
	baud int
}

func (p *linuxSerialPort) Read(b []byte) (int, error)  { return p.file.Read(b) }
func (p *linuxSerialPort) Write(b []byte) (int, error) { return p.file.Write(b) }
func (p *linuxSerialPort) Close() error                { return p.file.Close() }
func (p *linuxSerialPort) String() string              { return fmt.Sprintf("%s@%d", p.port, p.baud) }

func configureTermios(fd int, baud int) error {
	t, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	if err != nil {
		return fmt.Errorf("serial: ioctl TCGETS: %w", err)
	}

	speed, err := baudToUnix(baud)
	if err != nil {
		return err
	}

	// Raw-ish 8N1.
	t.Iflag = 0
	t.Oflag = 0
	t.Lflag = 0
	t.Cflag = unix.CS8 | unix.CREAD | unix.CLOCAL

	// Disable flow control.
	t.Cflag &^= unix.CRTSCTS

	// VMIN/VTIME: block until at least 1 byte.
	t.Cc[unix.VMIN] = 1
	t.Cc[unix.VTIME] = 0

	// Set baud.
	t.Ispeed = uint32(speed)
	t.Ospeed = uint32(speed)

	if err := unix.IoctlSetTermios(fd, unix.TCSETS, t); err != nil {
		return fmt.Errorf("serial: ioctl TCSETS: %w", err)
	}

	// Give the line a moment to settle.
	time.Sleep(10 * time.Millisecond)
	return nil
}

func baudToUnix(baud int) (uint32, error) {
	switch baud {
	case 4800:
		return unix.B4800, nil
	case 9600:
		return unix.B9600, nil
	case 19200:
		return unix.B19200, nil
	case 38400:
		return unix.B38400, nil
	case 57600:
		return unix.B57600, nil
	case 115200:
		return unix.B115200, nil
	default:
		return 0, fmt.Errorf("serial: unsupported baud %d (supported: 4800,9600,19200,38400,57600,115200)", baud)
	}
}
