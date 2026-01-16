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
				require.Error(t, err, "expected error but got none")
			} else {
				require.NoError(t, err, "unexpected error")
				require.InDelta(t, tt.wantLat, lat, 1e-6, "latitude mismatch")
				require.InDelta(t, tt.wantLon, lon, 1e-6, "longitude mismatch")
			}
		})
	}
}
