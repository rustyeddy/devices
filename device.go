// Package device provides a framework for managing hardware devices
// with support for messaging, periodic operations, and state management.
// Devices can be controlled via MQTT messages and can publish their
// state and data periodically.
package devices

import "errors"

type Type uint8

const (
	TypeBool Type = iota
	TypeInt
	TypeFloat
	TypeString
	TypeBytes
	TypeJSON
	TypeAny
)

type DeviceState uint8

const (
	StateUnknown DeviceState = iota
	StateInitializing
	StateRunning
	StateError
	StateStopped
)

var (
	mocking               = false
	ErrNotImplemented     = errors.New("Method is not implemented")
	ErrTypeNotImplemented = errors.New("Type is not implemented")
)

// Device is a type-safe device contract for a single value type T.
// Implementations may be read-only by returning an error from Set.
type Device[T any] interface {
	ID() string
	Index() int
	Type() Type
	Open() error
	Close() error

	Get() (T, error)
	Set(v T) error
}

func SetMock(v bool) {
	mocking = v
}

func IsMock() bool {
	return mocking
}

type DeviceBase[T any] struct {
	name string
	index int
	devtype Type
}

func NewDeviceBase[T any](name string, index int) *DeviceBase[T] {
	return &DeviceBase[T]{
		name: name,
		index: index,
	}
}

func (d *DeviceBase[T]) ID() string {
	return d.name
}

func (d *DeviceBase[T]) Index() int {
	return d.index
}

func (d *DeviceBase[T]) Type() Type {
	return d.devtype
}

func (d *DeviceBase[T]) Open() error {
	return ErrTypeNotImplemented
}

func (d *DeviceBase[T]) Close() error {
	return ErrTypeNotImplemented
}

func (d *DeviceBase[T]) Get() (T, error) {
	var zero T
	return zero, ErrTypeNotImplemented
}

func (d *DeviceBase[T]) Set(v T) error {
	return ErrTypeNotImplemented
}

