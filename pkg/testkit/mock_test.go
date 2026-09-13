package testkit

import (
	"testing"

	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/driver/input"
)

func TestMockDriverEvents(t *testing.T) {
	mock := NewMockDriver(80, 24)

	mock.SendRune('A')
	mock.SendKey(input.KeyEnter, 0, cell.AttrNone)
	mock.SendMouse(10, 5, input.MouseLeft, input.MousePress, cell.AttrNone)
	mock.SendResize(100, 30)

	ev1 := <-mock.Events()
	if ev1.Type != input.EventKey || ev1.Key.Rune != 'A' {
		t.Errorf("Expected rune 'A', got %+v", ev1)
	}

	ev2 := <-mock.Events()
	if ev2.Type != input.EventKey || ev2.Key.Type != input.KeyEnter {
		t.Errorf("Expected KeyEnter, got %+v", ev2)
	}

	ev3 := <-mock.Events()
	if ev3.Type != input.EventMouse || ev3.Mouse.X != 10 || ev3.Mouse.Y != 5 {
		t.Errorf("Expected Mouse at (10, 5), got %+v", ev3)
	}

	ev4 := <-mock.Events()
	if ev4.Type != input.EventResize || ev4.Width != 100 || ev4.Height != 30 {
		t.Errorf("Expected Resize to 100x30, got %+v", ev4)
	}

	w, h, _ := mock.Size()
	if w != 100 || h != 30 {
		t.Errorf("Expected updated Size() to be 100x30, got %dx%d", w, h)
	}
}
