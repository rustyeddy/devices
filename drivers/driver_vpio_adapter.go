package drivers

// VPIOAdapter adapts VPIO[T] to implement GPIO[T] interface
type VPIOAdapter[T Value] struct {
	vpio *VPIO[T]
}

// NewVPIOAdapter creates a new adapter for VPIO
func NewVPIOAdapter[T Value]() GPIO[T] {
	return &VPIOAdapter[T]{
		vpio: NewVPIO[T](),
	}
}

func (v *VPIOAdapter[T]) Pin(name string, pin int, options PinOptions) (Pin[T], error) {
	// Convert PinOptions to Direction
	var dir Direction
	const (
		PinInput  PinOptions = 1 << 0
		PinOutput PinOptions = 1 << 1
	)
	
	if options&PinOutput != 0 {
		dir = DirectionOutput
	} else {
		dir = DirectionInput
	}
	
	p, err := v.vpio.Pin(name, uint(pin), dir)
	if err != nil {
		return nil, err
	}
	return &VPinAdapter[T]{p}, nil
}

func (v *VPIOAdapter[T]) Get(pin int) (T, error) {
	return v.vpio.Get(uint(pin))
}

func (v *VPIOAdapter[T]) Set(pin int, val T) error {
	return v.vpio.Set(uint(pin), val)
}

func (v *VPIOAdapter[T]) Close() error {
	return v.vpio.Close()
}

// VPinAdapter adapts VPin[T] to implement Pin[T] interface
type VPinAdapter[T Value] struct {
	*VPin[T]
}

func (p *VPinAdapter[T]) Get() (T, error) {
	return p.VPin.Get()
}

func (p *VPinAdapter[T]) Set(v T) error {
	return p.VPin.Set(v)
}
