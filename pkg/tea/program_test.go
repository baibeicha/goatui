package tea

import (
	"strings"
	"testing"
	"time"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/driver/input"
	"github.com/baibeicha/goatui/pkg/style"
	"github.com/baibeicha/goatui/pkg/testkit"
)

type testCounterModel struct {
	count  int
	clicked bool
}

func (m testCounterModel) Init() Cmd {
	return nil
}

func (m testCounterModel) Update(msg Msg) (Model, Cmd) {
	switch msg := msg.(type) {
	case KeyMsg:
		if msg.Rune == '+' {
			m.count++
		}
		if msg.Rune == 'q' {
			return m, Quit
		}
	case HitMsg:
		if msg.Target.ID == "inc-btn" {
			m.clicked = true
			m.count += 10
		}
	}
	return m, nil
}

func (m testCounterModel) View(f *Frame) {
	// Draw counter
	st := style.NewStyle().Bold(true)
	st.Draw(f.Buffer, buffer.NewRect(0, 0, 20, 1), "Counter App")

	// Register interactive button at (5, 2, 10, 1)
	btnRect := buffer.NewRect(5, 2, 10, 1)
	st.Draw(f.Buffer, btnRect, "[+10 Button]")
	f.RegisterHit("inc-btn", btnRect, 1, nil)
}

func TestProgramCounterAndMouseHit(t *testing.T) {
	mock := testkit.NewMockDriver(80, 24)
	initial := testCounterModel{}

	prog := NewProgram(initial, WithDriver(mock))

	done := make(chan Model, 1)
	go func() {
		finalModel, err := prog.Run()
		if err != nil {
			t.Errorf("Program.Run failed: %v", err)
		}
		done <- finalModel
	}()

	// Wait for initial render
	time.Sleep(20 * time.Millisecond)

	// Verify initial frame drawn
	out := mock.OutString()
	if !strings.Contains(out, "Counter App") {
		t.Errorf("Expected 'Counter App' in mock output, got: %q", out)
	}

	// 1. Send Key '+'
	mock.SendRune('+')
	time.Sleep(10 * time.Millisecond)

	// 2. Send Mouse Press AND Release inside registered button at (7, 2)
	mock.SendMouse(7, 2, input.MouseLeft, input.MousePress, cell.AttrNone)
	time.Sleep(10 * time.Millisecond)
	mock.SendMouse(7, 2, input.MouseLeft, input.MouseRelease, cell.AttrNone)
	time.Sleep(10 * time.Millisecond)

	// 3. Send Key 'q' to quit
	mock.SendRune('q')

	select {
	case final := <-done:
		m := final.(testCounterModel)
		if m.count != 11 { // 1 from '+', 10 from button click
			t.Errorf("Expected count = 11, got %d", m.count)
		}
		if !m.clicked {
			t.Errorf("Expected button to be clicked via HitMsg")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Program did not quit in time")
	}
}
