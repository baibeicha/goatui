package input

import (
	"testing"
)

func TestParseKeys(t *testing.T) {
	p := NewParser()

	var events []Event
	handler := func(ev Event) {
		events = append(events, ev)
	}

	// Test regular text
	p.Parse([]byte("Goat"), handler)
	if len(events) != 4 {
		t.Fatalf("Expected 4 events, got %d", len(events))
	}
	if events[0].Key.Rune != 'G' || events[3].Key.Rune != 't' {
		t.Errorf("Unexpected runes parsed: %+v", events)
	}

	// Test Arrow Up: \x1b[A
	events = events[:0]
	p.Parse([]byte("\x1b[A"), handler)
	if len(events) != 1 || events[0].Key.Type != KeyUp {
		t.Errorf("Expected KeyUp, got %+v", events)
	}

	// Test Delete: \x1b[3~
	events = events[:0]
	p.Parse([]byte("\x1b[3~"), handler)
	if len(events) != 1 || events[0].Key.Type != KeyDelete {
		t.Errorf("Expected KeyDelete, got %+v", events)
	}

	// Test F1: \x1bOP
	events = events[:0]
	p.Parse([]byte("\x1bOP"), handler)
	if len(events) != 1 || events[0].Key.Type != KeyF1 {
		t.Errorf("Expected KeyF1, got %+v", events)
	}
}

func TestParseSGRMouse(t *testing.T) {
	p := NewParser()

	var events []Event
	handler := func(ev Event) {
		events = append(events, ev)
	}

	// Left click at (10, 20): \x1b[<0;11;21M
	p.Parse([]byte("\x1b[<0;11;21M"), handler)
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}
	m := events[0].Mouse
	if m.Button != MouseLeft || m.Action != MousePress || m.X != 10 || m.Y != 20 {
		t.Errorf("Unexpected mouse event: %+v", m)
	}

	// Wheel Up at (5, 5): \x1b[<64;6;6M
	events = events[:0]
	p.Parse([]byte("\x1b[<64;6;6M"), handler)
	if len(events) != 1 || events[0].Mouse.Button != MouseWheelUp {
		t.Errorf("Expected MouseWheelUp, got %+v", events)
	}
}

func TestParseBracketedPaste(t *testing.T) {
	p := NewParser()

	var events []Event
	handler := func(ev Event) {
		events = append(events, ev)
	}

	p.Parse([]byte("\x1b[200~Pasted Content\x1b[201~"), handler)
	if len(events) != 1 || events[0].Type != EventPaste || events[0].PasteText != "Pasted Content" {
		t.Errorf("Expected EventPaste, got %+v", events)
	}
}

func TestParseIncompleteEscapeSequence(t *testing.T) {
	p := NewParser()

	var events []Event
	handler := func(ev Event) {
		events = append(events, ev)
	}

	// Send partial escape sequence
	p.Parse([]byte("\x1b["), handler)
	if len(events) != 0 {
		t.Fatalf("Expected 0 events for incomplete escape sequence, got %d", len(events))
	}

	// Complete sequence with 'B' (Down arrow)
	p.Parse([]byte("B"), handler)
	if len(events) != 1 || events[0].Key.Type != KeyDown {
		t.Errorf("Expected KeyDown after completion, got %+v", events)
	}
}
