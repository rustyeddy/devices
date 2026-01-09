package drivers

import (
	"testing"
)

func TestVPIONewMethods(t *testing.T) {
	vpio := NewVPIO[int]()

	pin, err := vpio.SetPin("test-pin", 10, PinOutput)
	if err != nil {
		t.Fatalf("Failed to initialize pin: %v", err)
	}

	if pin.ID() != "test-pin" {
		t.Errorf("Expected pin ID 'test-pin', got '%s'", pin.ID())
	}

	if pin.Index() != 10 {
		t.Errorf("Expected pin index 10, got %d", pin.Index())
	}

	// if pin.Direction() != pinOptionsToGPIOCDev(PinOutput) {
	// 	t.Errorf("Expected direction Output, got %d", pin.Direction())
	// }
}

func TestVPIORecordingAPI(t *testing.T) {
	vpio := NewVPIO[bool]()
	_, err := vpio.SetPin("test", 5, PinOutput)
	if err != nil {
		t.Fatalf("Failed to initialize pin: %v", err)
	}

	vpio.StartRecording()
	vpio.Set(5, true)
	vpio.Set(5, false)
	vpio.Set(5, true)
	vpio.StopRecording()

	transactions := vpio.GetTransactions()
	if len(transactions) != 3 {
		t.Errorf("Expected 3 transactions, got %d", len(transactions))
	}

	vpio.ClearTransactions()
	if len(vpio.GetTransactions()) != 0 {
		t.Error("Transactions should be cleared")
	}
}
