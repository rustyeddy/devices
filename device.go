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
	TypeAny
	TypeBME280
	TypeGPS
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
