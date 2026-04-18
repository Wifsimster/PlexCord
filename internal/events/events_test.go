package events

import "testing"

func TestRecordingBus_Emit(t *testing.T) {
	bus := NewRecordingBus()

	bus.Emit("Event1", "payload1")
	bus.Emit("Event2", 42, "payload2")
	bus.Emit("Event1", nil)

	if len(bus.Events) != 3 {
		t.Errorf("expected 3 events, got %d", len(bus.Events))
	}
	if bus.Events[0].Name != "Event1" {
		t.Errorf("expected Event1, got %s", bus.Events[0].Name)
	}
	if len(bus.Events[1].Payload) != 2 {
		t.Errorf("expected 2 payload items, got %d", len(bus.Events[1].Payload))
	}
}

func TestRecordingBus_Count(t *testing.T) {
	bus := NewRecordingBus()
	bus.Emit("A", nil)
	bus.Emit("B", nil)
	bus.Emit("A", nil)
	bus.Emit("C", nil)

	if got := bus.Count("A"); got != 2 {
		t.Errorf("expected 2 A events, got %d", got)
	}
	if got := bus.Count("B"); got != 1 {
		t.Errorf("expected 1 B event, got %d", got)
	}
	if got := bus.Count("Missing"); got != 0 {
		t.Errorf("expected 0 Missing events, got %d", got)
	}
}

func TestRecordingBus_Reset(t *testing.T) {
	bus := NewRecordingBus()
	bus.Emit("A", nil)
	bus.Emit("B", nil)

	bus.Reset()

	if len(bus.Events) != 0 {
		t.Errorf("expected empty events after reset, got %d", len(bus.Events))
	}
}

func TestWailsBus_NilContext(t *testing.T) {
	// Emitting with nil context should not panic (it's a no-op)
	bus := NewWailsBus(nil) //nolint:staticcheck // SA1012: intentionally nil to verify defensive behavior
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("emit on nil context should not panic: %v", r)
		}
	}()
	bus.Emit("Test", "payload")
}

func TestWailsBus_NilReceiver(t *testing.T) {
	// Emitting on nil *WailsBus should not panic
	var bus *WailsBus
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("emit on nil bus should not panic: %v", r)
		}
	}()
	bus.Emit("Test", "payload")
}
