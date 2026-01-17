package sensors

import (
	"context"
)

// Sensor is the minimal interface a read-only “data provider” sensor exposes.
// Keep it tiny; wrappers depend on this, not on concrete devices.
type Sensor[T any] interface {
	Run(ctx context.Context) error
	Close() error
	Read() <-chan T
}

func Itoa(n int) string {
	if n == 0 {
		return "0"
	}
	sign := ""
	if n < 0 {
		sign = "-"
		n = -n
	}
	var buf [12]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + (n % 10))
		n /= 10
	}
	return sign + string(buf[i:])
}
