package sensors

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/rustyeddy/devices"
	"github.com/rustyeddy/devices/drivers"
	"github.com/stretchr/testify/require"
)

func TestGTU7_Descriptor(t *testing.T) {
	t.Run("with serial config", func(t *testing.T) {
		gps := NewGTU7(GTU7Config{
			Name: "test-gps",
			Serial: drivers.SerialConfig{
				Port: "/dev/ttyUSB0",
				Baud: 9600,
			},
			Reader: strings.NewReader(""),
		})

		desc := gps.Descriptor()
		require.Equal(t, "test-gps", desc.Name)
		require.Equal(t, "gps", desc.Kind)
		require.Equal(t, "GPSFix", desc.ValueType)
		require.Equal(t, devices.ReadOnly, desc.Access)
		require.Contains(t, desc.Tags, "gps")
		require.Contains(t, desc.Tags, "navigation")
		require.Contains(t, desc.Tags, "location")
		require.Equal(t, "/dev/ttyUSB0", desc.Attributes["port"])
		require.Equal(t, "9600", desc.Attributes["baud"])
	})

	t.Run("without serial config", func(t *testing.T) {
		gps := NewGTU7(GTU7Config{
			Name:   "test-gps-2",
			Reader: strings.NewReader(""),
		})

		desc := gps.Descriptor()
		require.Equal(t, "test-gps-2", desc.Name)
		require.Equal(t, "gps", desc.Kind)
		require.Contains(t, desc.Tags, "gps")
		require.Contains(t, desc.Tags, "navigation")
		require.Contains(t, desc.Tags, "location")
		// Attributes should be empty when no serial config
		require.Empty(t, desc.Attributes)
	})
}

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
