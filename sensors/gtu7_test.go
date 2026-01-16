package sensors

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGTU7_EmitsFixFromGGA(t *testing.T) {
	input := `
$GPGGA,160446.00,3340.34121,N,11800.11332,W,2,08,1.20,11.8,M,-33.1,M,,0000*58
$GPGSA,A,3,09,16,46,03,07,31,26,04,,,,,3.08,1.20,2.84*0E
$GPGSV,4,1,13,01,02,193,,03,58,181,33,04,64,360,31,06,12,295,*7A
$GPGLL,3340.34121,N,11800.11332,W,160446.00,A,D*74
`

	gps := NewGTU7(GTU7Config{Reader: strings.NewReader(input)})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- gps.Run(ctx) }()

	select {
	case fix := <-gps.Out():
		cancel()
		// lat 33 deg + 40.34121/60
		require.InDelta(t, 33.6723535, fix.Lat, 1e-6)
		// lon -(118 deg + 0.11332/60)
		require.InDelta(t, -118.0018887, fix.Lon, 1e-6)
		require.Equal(t, 2, fix.Quality)
		require.Equal(t, 8, fix.Satellites)
		require.InDelta(t, 1.20, fix.HDOP, 1e-6)
		require.InDelta(t, 11.8, fix.AltitudeM, 1e-6)
		require.Equal(t, "160446.00", fix.UTCTime)
	case err := <-done:
		require.NoError(t, err)
		require.FailNow(t, "expected at least one fix")
	case <-time.After(1 * time.Second):
		require.FailNow(t, "timed out waiting for fix")
	}

	// Ensure Run exits cleanly after cancel.
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(1 * time.Second):
		require.FailNow(t, "timed out waiting for Run to exit")
	}
}
