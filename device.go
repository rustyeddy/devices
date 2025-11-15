// Package device provides a framework for managing hardware devices
// with support for messaging, periodic operations, and state management.
// Devices can be controlled via MQTT messages and can publish their
// state and data periodically.
package devices

import (
	"errors"
	"fmt"
	"time"

	"github.com/rustyeddy/otto/messanger"
)

type Type uint8

const (
	TypeBool Type = iota
	TypeInt
	TypeFloat
	TypeString
	TypeBytes
	TypeJSON
	TypeAny
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

type DeviceEvent struct {
	Type    string
	Message string
	Time    time.Time
}

const (
	DeviceEventInitialized = "initialized"
	DeviceEventError       = "error"
	DeviceEventDataReady   = "data"
	DeviceEventRisingEdge  = "rising"
	DeviceEventFallingEdge = "falling"
)

var (
	mocking               = false
	ErrNotImplemented     = errors.New("method is not implemented")
	ErrTypeNotImplemented = errors.New("type is not implemented")
)

// Device is a type-safe device contract for a single value type T.
// Implementations may be read-only by returning an error from Set.
type Device[T any] interface {
	Name() string
	Type() Type
	Open() error
	Close() error

	Get() (T, error)
	Set(v T) error

	TickHandler() *func(time.Time)
	StartTicker(period time.Duration, f *func(time.Time))
	RegisterEventHandler(f func(evt *DeviceEvent))
	HandleMsg(msg *messanger.Msg) error

	String() string
}

func SetMock(v bool) {
	mocking = v
}

func IsMock() bool {
	return mocking
}

type TimerHandler func(time.Time)

type DeviceBase[T any] struct {
	name         string
	devtype      Type
	tickHandler  *func(time.Time)
	eventHandler *func(evt *DeviceEvent)
}

func NewDeviceBase[T any](name string) *DeviceBase[T] {
	return &DeviceBase[T]{
		name: name,
	}
}

func (d *DeviceBase[T]) Open() error {
	return ErrTypeNotImplemented
}

func (d *DeviceBase[T]) Close() error {
	return ErrTypeNotImplemented
}

func (d *DeviceBase[T]) Name() string {
	return d.name
}

func (d *DeviceBase[T]) Type() Type {
	return d.devtype
}

func (d *DeviceBase[T]) Get() (T, error) {
	var zero T
	return zero, ErrTypeNotImplemented
}

func (d *DeviceBase[T]) Set(v T) error {
	return ErrTypeNotImplemented
}

func (d *DeviceBase[T]) String() string {
	return fmt.Sprintf("%s [%d]", d.Name(), d.devtype)
}

func (d *DeviceBase[T]) StartTicker(period time.Duration, f *func(time.Time)) {
	d.tickHandler = f
	go func() {
		ticker := time.NewTicker(period)
		for t := range ticker.C {
			if d.tickHandler != nil && *d.tickHandler != nil {
				(*d.tickHandler)(t)
			}
		}
	}()
}

func (d *DeviceBase[T]) RegisterEventHandler(f func(evt *DeviceEvent)) {
	d.eventHandler = &f
	evt := &DeviceEvent{
		Type:    DeviceEventInitialized,
		Message: "event handler registered",
		Time:    time.Now(),
	}
	(*d.eventHandler)(evt)
}

func (d *DeviceBase[T]) TickHandler() *func(time.Time) {
	return d.tickHandler
}

func (d *DeviceBase[T]) HandleMsg(msg *messanger.Msg) error {
	return ErrNotImplemented
}
