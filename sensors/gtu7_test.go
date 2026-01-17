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

func TestParseLatLon(t *testing.T) {
	tests := []struct {
		name    string
		lat     string
		ns      string
		lon     string
		ew      string
		wantLat float64
		wantLon float64
		wantErr bool
	}{
		{
			name:    "valid coordinates - northern hemisphere, western hemisphere",
			lat:     "3340.34121",
			ns:      "N",
			lon:     "11800.11332",
			ew:      "W",
			wantLat: 33.6723535,
			wantLon: -118.0018886666667,
			wantErr: false,
		},
		{
			name:    "valid coordinates - southern hemisphere, eastern hemisphere",
			lat:     "3340.34121",
			ns:      "S",
			lon:     "11800.11332",
			ew:      "E",
			wantLat: -33.6723535,
			wantLon: 118.0018886666667,
			wantErr: false,
		},
		{
			name:    "empty latitude",
			lat:     "",
			ns:      "N",
			lon:     "11800.11332",
			ew:      "W",
			wantErr: true,
		},
		{
			name:    "empty longitude",
			lat:     "3340.34121",
			ns:      "N",
			lon:     "",
			ew:      "W",
			wantErr: true,
		},
		{
			name:    "invalid latitude - not a number",
			lat:     "invalid",
			ns:      "N",
			lon:     "11800.11332",
			ew:      "W",
			wantErr: true,
		},
		{
			name:    "invalid longitude - not a number",
			lat:     "3340.34121",
			ns:      "N",
			lon:     "invalid",
			ew:      "W",
			wantErr: true,
		},
		{
			name:    "both coordinates invalid",
			lat:     "not-a-number",
			ns:      "N",
			lon:     "also-invalid",
			ew:      "W",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lat, lon, err := parseLatLon(tt.lat, tt.ns, tt.lon, tt.ew)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			const epsilon = 1e-6
			require.InDelta(t, tt.wantLat, lat, epsilon)
			require.InDelta(t, tt.wantLon, lon, epsilon)
		})
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
func TestGTU7_MalformedSentences(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "empty lines",
			input: `


$GPGGA,160446.00,3340.34121,N,11800.11332,W,2,08,1.20,11.8,M,-33.1,M,,0000*58
`,
		},
		{
			name: "truncated GGA",
			input: `
$GPGGA,160446.00,3340
$GPGGA,160446.00,3340.34121,N,11800.11332,W,2,08,1.20,11.8,M,-33.1,M,,0000*58
`,
		},
		{
			name: "invalid sentence type",
			input: `
$GPXYZ,160446.00,A,3340.34121,N,11800.11332,W,7.25,123.40,160126,,,A*00
$GPGGA,160446.00,3340.34121,N,11800.11332,W,2,08,1.20,11.8,M,-33.1,M,,0000*58
`,
		},
		{
			name: "missing lat/lon fields",
			input: `
$GPGGA,160446.00,,,,,2,08,1.20,11.8,M,-33.1,M,,0000*58
$GPGGA,160446.00,3340.34121,N,11800.11332,W,2,08,1.20,11.8,M,-33.1,M,,0000*58
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gps := NewGTU7(GTU7Config{
				Reader: strings.NewReader(tt.input),
			})

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			done := make(chan error, 1)
			go func() { done <- gps.Run(ctx) }()

			// Should get at least one valid fix
			select {
			case fix := <-gps.Out():
				require.NotZero(t, fix.Lat)
				require.NotZero(t, fix.Lon)
				cancel()
			case <-time.After(time.Second):
				require.FailNow(t, "timeout waiting for valid fix")
			}

			select {
			case err := <-done:
				require.NoError(t, err)
			case <-time.After(time.Second):
				require.FailNow(t, "run did not exit")
			}
		})
	}
}

func TestGTU7_GGAOnly(t *testing.T) {
	input := `
$GPGGA,123519.00,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*47
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
	case fix := <-gps.Out():
		cancel()
		// Verify GGA fields
		require.InDelta(t, 48.1173, fix.Lat, 0.0001)
		require.InDelta(t, 11.5167, fix.Lon, 0.0001)
		require.InDelta(t, 545.4, fix.AltMeters, 1e-6)
		require.InDelta(t, 0.9, fix.HDOP, 1e-6)
		require.Equal(t, 8, fix.Satellites)
		require.Equal(t, 1, fix.Quality)
		// Speed/course should be zero/unset
		require.Zero(t, fix.SpeedKnots)
		require.Zero(t, fix.SpeedMPS)
		require.Zero(t, fix.CourseDeg)
	case <-time.After(time.Second):
		require.FailNow(t, "timeout waiting for GGA fix")
	}

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(time.Second):
		require.FailNow(t, "run did not exit")
	}
}

func TestGTU7_ContextCancellation(t *testing.T) {
	// Test that context cancellation is checked between sentences
	input := `$GPGGA,160446.00,3340.34121,N,11800.11332,W,2,08,1.20,11.8,M,-33.1,M,,0000*58
$GPGGA,160447.00,3340.34122,N,11800.11333,W,2,08,1.20,11.8,M,-33.1,M,,0000*58
`

	gps := NewGTU7(GTU7Config{
		Reader: strings.NewReader(input),
	})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- gps.Run(ctx) }()

	// Wait for first fix
	select {
	case <-gps.Out():
	case <-time.After(time.Second):
		require.FailNow(t, "timeout waiting for first fix")
	}

	// Cancel context before second fix can be processed
	cancel()

	// Run should exit cleanly after processing completes
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(time.Second):
		require.FailNow(t, "run did not exit")
	}
}
		require.FailNow(t, "run did not exit after cancellation")
	}

	// Drain any remaining messages and verify channel closes
	for range gps.Out() {
		// drain
	}
}

func TestGTU7_MultiConstellationVariants(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantLat  float64
		wantLon  float64
		wantAlt  float64
		wantSats int
	}{
		{
			name:     "GNGGA - multi-constellation",
			input:    "$GNGGA,123519.00,4807.038,N,01131.000,E,1,12,0.9,545.4,M,46.9,M,,*4E\n",
			wantLat:  48.1173,
			wantLon:  11.5167,
			wantAlt:  545.4,
			wantSats: 12,
		},
		{
			name:    "GNRMC - multi-constellation",
			input:   "$GNRMC,123519.00,A,4807.038,N,01131.000,E,5.5,123.4,230394,,,A*57\n",
			wantLat: 48.1173,
			wantLon: 11.5167,
		},
		{
			name:    "GNVTG with GPGGA - multi-constellation",
			input:   "$GNVTG,54.7,T,,M,5.5,N,10.2,K,A*2F\n$GPGGA,123519.00,4807.038,N,01131.000,E,1,08,0.9,545.4,M,46.9,M,,*47\n",
			wantLat: 48.1173,
			wantLon: 11.5167,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gps := NewGTU7(GTU7Config{
				Reader: strings.NewReader(tt.input),
			})

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			done := make(chan error, 1)
			go func() { done <- gps.Run(ctx) }()

			// Drain any intermediate fixes
			var lastFix GPSFix
			timeout := time.After(time.Second)
		loop:
			for {
				select {
				case fix, ok := <-gps.Out():
					if !ok {
						break loop
					}
					lastFix = fix
				case <-timeout:
					break loop
				}
			}

			cancel()

			// Verify we got data
			require.NotZero(t, lastFix.Lat, "should have received at least one fix")
			require.InDelta(t, tt.wantLat, lastFix.Lat, 0.0001)
			require.InDelta(t, tt.wantLon, lastFix.Lon, 0.0001)
			if tt.wantAlt != 0 {
				require.InDelta(t, tt.wantAlt, lastFix.AltMeters, 1e-6)
			}
			if tt.wantSats != 0 {
				require.Equal(t, tt.wantSats, lastFix.Satellites)
			}

			select {
			case err := <-done:
				require.NoError(t, err)
			case <-time.After(time.Second):
				require.FailNow(t, "run did not exit")
			}
		})
	}
func TestGTU7_BufferSize(t *testing.T) {
	t.Run("default buffer size is 16", func(t *testing.T) {
		gps := NewGTU7(GTU7Config{
			Reader: strings.NewReader(""),
		})
		require.Equal(t, 16, cap(gps.out))
	})

	t.Run("custom buffer size", func(t *testing.T) {
		gps := NewGTU7(GTU7Config{
			Reader: strings.NewReader(""),
			Buf:    32,
		})
		require.Equal(t, 32, cap(gps.out))
	})

	t.Run("zero buffer size defaults to 16", func(t *testing.T) {
		gps := NewGTU7(GTU7Config{
			Reader: strings.NewReader(""),
			Buf:    0,
		})
		require.Equal(t, 16, cap(gps.out))
	})

	t.Run("negative buffer size defaults to 16", func(t *testing.T) {
		gps := NewGTU7(GTU7Config{
			Reader: strings.NewReader(""),
			Buf:    -5,
		})
		require.Equal(t, 16, cap(gps.out))
	})
}
