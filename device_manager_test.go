package devices

import (
	"fmt"
	"sync"
	"testing"
)

// mockDevice implements the minimal interface expected by Device[T]
type mockDevice struct {
	id   string
	data any
	kind Type
}

func (m *mockDevice) ID() string {
	return m.id
}

func (m *mockDevice) Type() Type {
	return m.kind
}

func (m *mockDevice) Get() (any, error) {
	return m.data, nil
}

func (m *mockDevice) Set(d any) error {
	m.data = d
	return nil
}

func (m *mockDevice) Open() error {
	return nil
}
func (m *mockDevice) Close() error {
	return nil
}

func TestNewDeviceManager_InitialState(t *testing.T) {
	dm := NewDeviceManager()
	if dm == nil {
		t.Fatal("NewDeviceManager returned nil")
	}
	if dm.devices == nil {
		t.Fatal("devices map should be initialized")
	}
	if len(dm.devices) != 0 {
		t.Fatalf("expected empty devices map, got %d entries", len(dm.devices))
	}
}

func TestAddAndGet(t *testing.T) {
	dm := NewDeviceManager()
	dev := &mockDevice{id: "dev1"}

	if err := dm.Add(dev); err != nil {
		t.Fatalf("Add returned unexpected error: %v", err)
	}

	got, err := dm.Get("dev1")
	if err != nil {
		t.Fatalf("Get returned unexpected error: %v", err)
	}
	if got == nil {
		t.Fatalf("expected non-nil device from Get, got nil")
	}
	if got.ID() != dev.ID() {
		t.Fatalf("Get returned device with ID %q, want %q", got.ID(), dev.ID())
	}
}

func TestAddDuplicate(t *testing.T) {
	dm := NewDeviceManager()
	dev := &mockDevice{id: "dup1"}

	if err := dm.Add(dev); err != nil {
		t.Fatalf("first Add returned unexpected error: %v", err)
	}

	if err := dm.Add(dev); err != ErrDeviceExists {
		t.Fatalf("second Add on same name: got error %v, want ErrDeviceExists", err)
	}
}

func TestGet_NotFound(t *testing.T) {
	dm := NewDeviceManager()

	_, err := dm.Get("missing")
	if err != ErrNotFound {
		t.Fatalf("Get on missing device: got error %v, want ErrNotFound", err)
	}
}

func TestConcurrentAdds(t *testing.T) {
	dm := NewDeviceManager()
	var wg sync.WaitGroup
	count := 200

	wg.Add(count)
	for i := 0; i < count; i++ {
		i := i
		go func() {
			defer wg.Done()
			id := fmt.Sprintf("concurrent-%d", i)
			_ = dm.Add(&mockDevice{id: id})
		}()
	}
	wg.Wait()

	if len(dm.devices) != count {
		t.Fatalf("after concurrent adds: got %d devices, want %d", len(dm.devices), count)
	}
}
