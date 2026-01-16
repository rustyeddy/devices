package sensors

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/rustyeddy/devices/drivers"
	"github.com/stretchr/testify/require"
)

func TestGTU7_PrefersRMCOverVTG(t *testing.T) {
	input := `
$GPGGA,160446.00,3340.34121,N,11800.11332,W,2,08,1.20,11.8,M,-33.1,M,,0000*58
$GPVTG,54.70,T,,M,5.50,N,10.19,K,A*00
$GPRMC,160446.00,A,3340.34121,N,11800.11332,W,7.25,123.40,160126,,,A*00
`

	gps, err := NewGTU7(GTU7Config{
		Reader: strings.NewReader(input),
	})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- gps.Run(ctx) }()

	// Drain GGA + VTG
	for i := 0; i < 2; i++ {
		select {
		case <-gps.Out():
		case <-time.After(time.Second):
			require.FailNow(t, "timeout")
		}
	}

	// RMC must override VTG
	select {
	case fix := <-gps.Out():
		cancel()
		require.InDelta(t, 7.25, fix.SpeedKnots, 1e-6)
		require.InDelta(t, 7.25*0.514444, fix.SpeedMPS, 1e-6)
		require.InDelta(t, 123.40, fix.CourseDeg, 1e-6)
		require.Equal(t, "A", fix.Status)
		require.Equal(t, "160126", fix.Date)
	case <-time.After(time.Second):
		require.FailNow(t, "timeout waiting for RMC fix")
	}

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(time.Second):
		require.FailNow(t, "run did not exit")
	}
}

// mockSerialFactory is a test helper that returns an error when opening serial.
type mockSerialFactory struct {
	err error
}

func (m mockSerialFactory) OpenSerial(cfg drivers.SerialConfig) (drivers.SerialPort, error) {
	return nil, m.err
}

func TestNewGTU7_OpenSerialError(t *testing.T) {
	expectedErr := errors.New("failed to open serial port")

	gps, err := NewGTU7(GTU7Config{
		Serial:  drivers.SerialConfig{Port: "/dev/ttyUSB0", Baud: 9600},
		Factory: mockSerialFactory{err: expectedErr},
	})

	require.Error(t, err)
	require.Nil(t, gps)
	require.Equal(t, expectedErr, err)
}
