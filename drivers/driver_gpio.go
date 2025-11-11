package drivers

type PinOptions uint
type EventType uint

type GPIO[T any] interface {
	Pin(name string, pin int, options PinOptions) (Pin[T], error)
	Get(pin int) (T, error)
	Set(pin int, v T) error
	Close() error
}

type Pin[T any] interface {
	ID() string
	Index() int
	Direction() Direction
	Get() (T, error)
	Set(v T) error
}

// Singleton instances for supported types
var (
	gpioBool    GPIO[bool]
	gpioInt     GPIO[int]
	gpioFloat64 GPIO[float64]
)

// GetGPIO returns the singleton GPIO instance for the specified type.
// On non-Linux systems or when mocking is enabled, it returns VPIO.
// On Linux systems, it returns GPIOCDev for hardware access.
// Supported types: bool, int, float64
func GetGPIO[T Value]() GPIO[T] {
	var zero T
	switch any(zero).(type) {
	case bool:
		if gpioBool == nil {
			gpioBool = createGPIO[bool]()
		}
		return any(gpioBool).(GPIO[T])
		
	case int:
		if gpioInt == nil {
			gpioInt = createGPIO[int]()
		}
		return any(gpioInt).(GPIO[T])
		
	case float64:
		if gpioFloat64 == nil {
			gpioFloat64 = createGPIO[float64]()
		}
		return any(gpioFloat64).(GPIO[T])
		
	default:
		// For other Value types, create a new VPIO instance (not cached)
		return NewVPIOAdapter[T]()
	}
}

// ResetGPIO clears all singleton instances (useful for testing)
func ResetGPIO() {
	gpioBool = nil
	gpioInt = nil
	gpioFloat64 = nil
}
