//go:build !linux

package drivers

// createGPIO creates a VPIO instance for non-Linux systems or when mocking.
func createGPIO[T Value]() GPIO[T] {
	return NewVPIOAdapter[T]()
}
