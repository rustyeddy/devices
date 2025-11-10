package drivers

type PinOptions uint
type EventType  uint

type GPIO[T any] interface {
	Pin(name string, pin int, options PinOptions) (*Pin[T], error)
	Get() T
	Set(v T)

	EventHandler(done chan any, event EventType)
}

type Pin[T any] interface {
	ID()		string
	Index()		int
	Direction() Direction
}
