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

func TestGTU7_FallbackToVTGWhenRMCStopsProvidingData(t *testing.T) {
	// Scenario: RMC initially provides speed/course, then stops (empty fields).
	// VTG should be used for speed/course after RMC stops providing it.
	input := `
$GPGGA,160446.00,3340.34121,N,11800.11332,W,2,08,1.20,11.8,M,-33.1,M,,0000*58
$GPRMC,160446.00,A,3340.34121,N,11800.11332,W,7.25,123.40,160126,,,A*00
$GPGGA,160447.00,3340.34121,N,11800.11332,W,2,08,1.20,11.8,M,-33.1,M,,0000*58
$GPRMC,160447.00,A,3340.34121,N,11800.11332,W,,,160126,,,A*00
$GPVTG,54.70,T,,M,5.50,N,10.19,K,A*00
`

	gps := NewGTU7(GTU7Config{
		Reader: strings.NewReader(input),
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- gps.Run(ctx) }()

	// First GGA
	var fix GPSFix
	select {
	case fix = <-gps.Out():
	case <-time.After(time.Second):
		require.FailNow(t, "timeout on first GGA")
	}

	// First RMC with speed/course
	select {
	case fix = <-gps.Out():
		require.InDelta(t, 7.25, fix.SpeedKnots, 1e-6)
		require.InDelta(t, 123.40, fix.CourseDeg, 1e-6)
	case <-time.After(time.Second):
		require.FailNow(t, "timeout on first RMC")
	}

	// Second GGA
	select {
	case fix = <-gps.Out():
	case <-time.After(time.Second):
		require.FailNow(t, "timeout on second GGA")
	}

	// Second RMC without speed/course (empty fields)
	select {
	case fix = <-gps.Out():
		// Speed/course from first RMC should still be in the state
		require.InDelta(t, 7.25, fix.SpeedKnots, 1e-6)
	case <-time.After(time.Second):
		require.FailNow(t, "timeout on second RMC")
	}

	// VTG should now update speed/course since RMC stopped providing it
	select {
	case fix = <-gps.Out():
		cancel()
		require.InDelta(t, 5.50, fix.SpeedKnots, 1e-6)
		require.InDelta(t, 5.50*0.514444, fix.SpeedMPS, 1e-6)
		require.InDelta(t, 54.70, fix.CourseDeg, 1e-6)
	case <-time.After(time.Second):
		require.FailNow(t, "timeout waiting for VTG to override")
	}

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(time.Second):
		require.FailNow(t, "run did not exit")
	}
}
