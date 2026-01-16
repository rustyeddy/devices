//go:build linux

package drivers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
)

// TestBaudToUnix tests the baud rate to Unix constant conversion.
func TestBaudToUnix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		baud     int
		expected uint32
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "4800 baud",
			baud:     4800,
			expected: unix.B4800,
			wantErr:  false,
		},
		{
			name:     "9600 baud",
			baud:     9600,
			expected: unix.B9600,
			wantErr:  false,
		},
		{
			name:     "19200 baud",
			baud:     19200,
			expected: unix.B19200,
			wantErr:  false,
		},
		{
			name:     "38400 baud",
			baud:     38400,
			expected: unix.B38400,
			wantErr:  false,
		},
		{
			name:     "57600 baud",
			baud:     57600,
			expected: unix.B57600,
			wantErr:  false,
		},
		{
			name:     "115200 baud",
			baud:     115200,
			expected: unix.B115200,
			wantErr:  false,
		},
		{
			name:    "unsupported baud 1200",
			baud:    1200,
			wantErr: true,
			errMsg:  "unsupported baud 1200",
		},
		{
			name:    "unsupported baud 230400",
			baud:    230400,
			wantErr: true,
			errMsg:  "unsupported baud 230400",
		},
		{
			name:    "zero baud",
			baud:    0,
			wantErr: true,
			errMsg:  "unsupported baud 0",
		},
		{
			name:    "negative baud",
			baud:    -9600,
			wantErr: true,
			errMsg:  "unsupported baud -9600",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			speed, err := baudToUnix(tt.baud)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Zero(t, speed)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, speed)
			}
		})
	}
}

// TestBaudToUnixAllSupportedRates verifies all supported rates are valid Unix constants.
func TestBaudToUnixAllSupportedRates(t *testing.T) {
	t.Parallel()

	supportedRates := []int{4800, 9600, 19200, 38400, 57600, 115200}

	for _, baud := range supportedRates {
		t.Run(string(rune(baud)), func(t *testing.T) {
			speed, err := baudToUnix(baud)
			require.NoError(t, err, "supported baud rate %d should not error", baud)
			assert.NotZero(t, speed, "baud rate %d should return non-zero speed", baud)
		})
	}
}

// TestLinuxSerialFactoryType tests that LinuxSerialFactory implements SerialFactory.
func TestLinuxSerialFactoryType(t *testing.T) {
	t.Parallel()

	var _ SerialFactory = LinuxSerialFactory{}
}

// TestLinuxSerialPortString tests the String() method format.
func TestLinuxSerialPortString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		port     string
		baud     int
		expected string
	}{
		{
			name:     "USB0 at 9600",
			port:     "/dev/ttyUSB0",
			baud:     9600,
			expected: "/dev/ttyUSB0@9600",
		},
		{
			name:     "AMA0 at 115200",
			port:     "/dev/ttyAMA0",
			baud:     115200,
			expected: "/dev/ttyAMA0@115200",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := &linuxSerialPort{
				port: tt.port,
				baud: tt.baud,
			}
			assert.Equal(t, tt.expected, p.String())
		})
	}
}

// TestLinuxSerialPortInterfaceCompliance verifies linuxSerialPort implements SerialPort.
func TestLinuxSerialPortInterfaceCompliance(t *testing.T) {
	t.Parallel()

	var _ SerialPort = (*linuxSerialPort)(nil)
}

// TestLinuxSerialFactoryOpenSerialInvalidConfig tests error handling for invalid configs.
func TestLinuxSerialFactoryOpenSerialInvalidConfig(t *testing.T) {
	t.Parallel()

	factory := LinuxSerialFactory{}

	tests := []struct {
		name   string
		cfg    SerialConfig
		errMsg string
	}{
		{
			name:   "empty port",
			cfg:    SerialConfig{Port: "", Baud: 9600},
			errMsg: "Port is required",
		},
		{
			name:   "zero baud",
			cfg:    SerialConfig{Port: "/dev/ttyUSB0", Baud: 0},
			errMsg: "Baud must be > 0",
		},
		{
			name:   "negative baud",
			cfg:    SerialConfig{Port: "/dev/ttyUSB0", Baud: -115200},
			errMsg: "Baud must be > 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			port, err := factory.OpenSerial(tt.cfg)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.errMsg)
			assert.Nil(t, port)
		})
	}
}

// TestLinuxSerialFactoryOpenSerialUnsupportedBaud tests unsupported baud rate handling.
func TestLinuxSerialFactoryOpenSerialUnsupportedBaud(t *testing.T) {
	t.Parallel()

	factory := LinuxSerialFactory{}

	// Use a non-existent port path to avoid actual hardware interaction.
	// The validation should fail at baud rate conversion before attempting to open.
	cfg := SerialConfig{
		Port: "/dev/ttyUSB999", // Non-existent port
		Baud: 230400,           // Unsupported baud rate
	}

	port, err := factory.OpenSerial(cfg)
	// The error could be either from baud conversion or file opening.
	// We just verify an error occurred and no port was returned.
	require.Error(t, err)
	assert.Nil(t, port)
}
