package vh400

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rustyeddy/devices"
	"github.com/rustyeddy/devices/drivers"
	"github.com/rustyeddy/devices/sensors"
	"github.com/stretchr/testify/require"
)

type fakeADC struct {
	volts  float64
	reads  atomic.Int64
	closes atomic.Int64
}

func (f *fakeADC) ReadVolts(ctx context.Context, channel int) (float64, error) {
	f.reads.Add(1)
	return f.volts, nil
}

func (f *fakeADC) Close() error {
	f.closes.Add(1)
	return nil
}

func TestVWCFromVolts_ClampAndSegments(t *testing.T) {
	t.Parallel()

	// out of range -> 0
	require.Equal(t, 0.0, vwcFromVolts(-0.1))
	require.Equal(t, 0.0, vwcFromVolts(3.1))

	// segment checks (spot values)
	require.InDelta(t, 9.0, vwcFromVolts(1.0), 0.0001)     // 10*1.0 - 1
	require.InDelta(t, 15.0, vwcFromVolts(1.3), 0.0001)    // 25*1.3 - 17.5
	require.InDelta(t, 40.0056, vwcFromVolts(1.82), 0.001) // 48.08*1.82 - 47.5
	require.InDelta(t, 50.104, vwcFromVolts(2.2), 0.01)    // 26.32*2.2 - 7.80
	require.Equal(t, 100.0, vwcFromVolts(3.0))
}

func TestVH400_Run_EmitInitialAndTick(t *testing.T) {
	t.Parallel()

	adc := &fakeADC{volts: 1.0}

	ft := &sensors.FakeTicker{Q: make(chan time.Time, 10)}

	v := NewVH400(VH400Config{
		Name:        "soil",
		ADC:         adc,
		Channel:     0,
		Interval:    1 * time.Second,
		EmitInitial: true,
		NewTicker:   func(time.Duration) sensors.Ticker { return ft },
	})

	// basic descriptor sanity
	d := v.Descriptor()
	require.Equal(t, "vh400", d.Kind)
	require.Equal(t, devices.ReadOnly, d.Access)

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() { errCh <- v.Run(ctx) }()

	// initial sample
	require.Equal(t, 9.0, <-v.Out())

	// tick sample
	adc.volts = 1.3
	ft.Q <- time.Now()
	require.Equal(t, 15.0, <-v.Out())

	cancel()
	require.NoError(t, <-errCh)

	// ADC should have been closed exactly once by Run().
	require.Equal(t, int64(1), adc.closes.Load())
}

var _ drivers.ADC = (*fakeADC)(nil)
