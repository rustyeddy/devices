package drivers

type PinOptions int

const (
	PinInput       PinOptions = 1 << 0
	PinOutput      PinOptions = 1 << 1
	PinOutputLow   PinOptions = 1 << 2
	PinOutputHigh  PinOptions = 1 << 3
	PinPullUp      PinOptions = 1 << 4
	PinPullDown    PinOptions = 1 << 5
	PinRisingEdge  PinOptions = 1 << 6
	PinFallingEdge PinOptions = 1 << 7
	PinBothEdges   PinOptions = 1 << 8
)

type EventType int

type GPIO[T any] interface {
	Open() error
	Close() error
	SetPin(name string, pin int, options ...PinOptions) (Pin[T], error)
	Get(pin int) (T, error)
	Set(pin int, v T) error
}

type Pin[T any] interface {
	ID() string
	Index() int
	Options() []PinOptions
	Get() (T, error)
	Set(v T) error
	ReadContinuous() <-chan T
}

type Value interface {
	~int | ~float64 | ~bool
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
		return NewVPIO[T]()
	}
}

func createGPIO[T Value]() GPIO[T] {
	return NewVPIO[T]()
}

// ResetGPIO clears all singleton instances (useful for testing)
func ResetGPIO() {
	gpioBool = nil
	gpioInt = nil
	gpioFloat64 = nil
}

// DigitalPin
type DigitalPin Pin[bool]
type AnalogPin Pin[float64]
