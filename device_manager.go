package devices

import (
	"errors"
	"sync"
)

var (
	ErrNotFound           = errors.New("device not found")
	ErrDeviceExists       = errors.New("device already")
	ErrTypeMismatch       = errors.New("device registered with different type")
	ErrTypeNotImplemented = errors.New("method not implemented")
)

type DeviceManager struct {
	mu sync.RWMutex
	// Heterogeneous registry: store typed Device[T] values behind `any`.
	devices map[string]Device[any]
	// Optional deps you may add:
	// log   Logger
	// bus   Bus
	// cfg   Config
}

func NewDeviceManager() *DeviceManager {
	return &DeviceManager{
		devices: make(map[string]Device[any]),
	}
}

// Register a typed device. Fails if name already used with a different T.
func (m *DeviceManager) Add(d Device[any]) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	name := d.ID()
	if _, exists := m.devices[name]; !exists {
		m.devices[name] = d
		return nil
	}
	return ErrDeviceExists
}

func (m *DeviceManager) Get(name string) (Device[any], error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	d, ok := m.devices[name]
	if !ok {
		return nil, ErrNotFound
	}
	return d, nil
}
