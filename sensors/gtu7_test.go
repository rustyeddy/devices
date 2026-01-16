package sensors

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGTU7_PrefersRMCOverVTG(t *testing.T) {
	input := `
$GPGGA,160446.00,3340.34121,N,11800.11332,W,2,08,1.20,11.8,M,-33.1,M,,0000*58
$GPVTG,54.70,T,,M,5.50,N,10.19,K,A*00
$GPRMC,160446.00,A,3340.34121,N,11800.11332,W,7.25,123.40,160126,,,A*00
`

	gps := NewGTU7(GTU7Config{
		Name:   "test-gps",
		Reader: strings.NewReader(input),
	})

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

func TestGTU7_EmitsWarningOnDroppedFix(t *testing.T) {
	// Create input with more fixes than buffer size to force drops
	input := `
$GPGGA,160446.00,3340.34121,N,11800.11332,W,2,08,1.20,11.8,M,-33.1,M,,0000*58
$GPGGA,160447.00,3340.34121,N,11800.11332,W,2,08,1.20,11.8,M,-33.1,M,,0000*58
$GPGGA,160448.00,3340.34121,N,11800.11332,W,2,08,1.20,11.8,M,-33.1,M,,0000*58
$GPGGA,160449.00,3340.34121,N,11800.11332,W,2,08,1.20,11.8,M,-33.1,M,,0000*58
$GPGGA,160450.00,3340.34121,N,11800.11332,W,2,08,1.20,11.8,M,-33.1,M,,0000*58
$GPGGA,160451.00,3340.34121,N,11800.11332,W,2,08,1.20,11.8,M,-33.1,M,,0000*58
`

	gps := NewGTU7(GTU7Config{
		Name:   "test-gps",
		Buf:    2, // Small buffer to force drops
		Reader: strings.NewReader(input),
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- gps.Run(ctx) }()

	// Check for drop warning in events while fixes are being sent
	// Don't drain the Out channel to force drops
	var foundDropWarning bool
	eventCheckTimeout := time.After(time.Second)

	for !foundDropWarning {
		select {
		case evt := <-gps.Events():
			if evt.Msg == "gps fix dropped" {
				foundDropWarning = true
				require.Equal(t, "output channel full", evt.Meta["reason"])
			}
		case <-eventCheckTimeout:
			require.FailNow(t, "timeout waiting for drop warning event")
		}
	}

	cancel()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(time.Second):
		require.FailNow(t, "run did not exit")
	}
}
