package devices

import (
	"fmt"
	"sync"
	"testing"
)

// helpers used by tests

type boolDevice struct {
	id  string
	val bool
	mu  sync.Mutex
}

func (d *boolDevice) ID() string { return d.id }
func (d *boolDevice) Type() Type { return TypeBool }
func (d *boolDevice) Get() (bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.val, nil
}
func (d *boolDevice) Set(v bool) error {
	d.mu.Lock()
	d.val = v
	d.mu.Unlock()
	return nil
}

type intDevice struct {
	id  string
	val int
	mu  sync.Mutex
}

func (d *intDevice) ID() string { return d.id }
func (d *intDevice) Type() Type { return TypeInt }
func (d *intDevice) Get() (int, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.val, nil
}
func (d *intDevice) Set(v int) error {
	d.mu.Lock()
	d.val = v
	d.mu.Unlock()
	return nil
}

type readOnlyDevice[T any] struct {
	id  string
	val T
}

func (d *readOnlyDevice[T]) ID() string { return d.id }
func (d *readOnlyDevice[T]) Type() Type { return TypeAny } // arbitrary
func (d *readOnlyDevice[T]) Get() (T, error) {
	return d.val, nil
}
func (d *readOnlyDevice[T]) Set(v T) error {
	return fmt.Errorf("device %s is read-only", d.id)
}

// Tests

func TestBoolDevice_GetSet(t *testing.T) {
	dev := &boolDevice{id: "bdev", val: false}

	v, err := dev.Get()
	if err != nil {
		t.Fatalf("Get returned unexpected error: %v", err)
	}
	if v != false {
		t.Fatalf("expected initial false, got %v", v)
	}

	if dev.ID() != "bdev" {
		t.Fatalf("ID mismatch: want %q got %q", "bdev", dev.ID())
	}
	if dev.Type() != TypeBool {
		t.Fatalf("Type mismatch: want TypeBool got %v", dev.Type())
	}

	if err := dev.Set(true); err != nil {
		t.Fatalf("Set returned unexpected error: %v", err)
	}
	v2, _ := dev.Get()
	if v2 != true {
		t.Fatalf("expected true after Set, got %v", v2)
	}
}

func TestIntDevice_GetSet(t *testing.T) {
	dev := &intDevice{id: "idev", val: 0}

	v, err := dev.Get()
	if err != nil {
		t.Fatalf("Get returned unexpected error: %v", err)
	}
	if v != 0 {
		t.Fatalf("expected initial 0, got %d", v)
	}

	if dev.ID() != "idev" {
		t.Fatalf("ID mismatch: want %q got %q", "idev", dev.ID())
	}
	if dev.Type() != TypeInt {
		t.Fatalf("Type mismatch: want TypeInt got %v", dev.Type())
	}

	if err := dev.Set(42); err != nil {
		t.Fatalf("Set returned unexpected error: %v", err)
	}
	v2, _ := dev.Get()
	if v2 != 42 {
		t.Fatalf("expected 42 after Set, got %d", v2)
	}
}

func TestReadOnlyDevice_SetFails(t *testing.T) {
	dev := &readOnlyDevice[int]{id: "rdev", val: 7}

	v, err := dev.Get()
	if err != nil {
		t.Fatalf("Get returned unexpected error: %v", err)
	}
	if v != 7 {
		t.Fatalf("expected initial 7, got %d", v)
	}

	if err := dev.Set(8); err == nil {
		t.Fatalf("expected error from Set on read-only device, got nil")
	}
	// ensure value unchanged
	v2, _ := dev.Get()
	if v2 != 7 {
		t.Fatalf("expected value to remain 7 after failed Set, got %d", v2)
	}
}

func TestConcurrentAccess_IntDevice(t *testing.T) {
	dev := &intDevice{id: "concurrent", val: 0}
	var wg sync.WaitGroup
	workers := 50
	opsPerWorker := 100

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(seed int) {
			defer wg.Done()
			for j := 0; j < opsPerWorker; j++ {
				_ = dev.Set(seed*j + j)
				_, _ = dev.Get()
			}
		}(i)
	}
	wg.Wait()

	// final value should be an int (non-negative)
	final, err := dev.Get()
	if err != nil {
		t.Fatalf("Get returned unexpected error: %v", err)
	}
	if final < 0 {
		t.Fatalf("final value unexpected: %d", final)
	}
}
