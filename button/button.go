package button

// type Button struct {
// 	*devices.DeviceBase
// }

// func New(name string, index int) *Button {
// 	return &Button{
// 		name:  name,
// 		index: index,
// 	}
// }

// func (b *Button) Name() string {
// 	return b.name
// }

// func (b *Button) Index() int {
// 	return b.index
// }

// func (b *Button) Open() error {

// 	v := GetVPIO[bool]()
// 	_, err := v.Pin(b.name, uint(b.index), DirectionInput)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (b *Button) Close() error {
// 	v := GetVPIO[bool]()
// 	return v.Close(b.index)
// }
