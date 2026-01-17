package sensors

import "context"

// Sensor is the minimal interface a read-only “data provider” sensor exposes.
// Keep it tiny; wrappers depend on this, not on concrete devices.
type Sensor[T any] interface {
	Run(ctx context.Context) error
	Close() error
	Read() <-chan T
}
