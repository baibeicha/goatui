package ui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/driver/input"
	"github.com/baibeicha/goatui/pkg/tea"
	"github.com/baibeicha/goatui/pkg/widgets"
)

func TestSpansAndLine(t *testing.T) {
	line := NewLine(
		Bold("Title: "),
		Color("Value", cell.ColorHex("#00FFAA")),
		BadgeSpan("OK", cell.ColorHex("#000000"), cell.ColorHex("#00FFAA")),
	)

	if line.Width() <= 0 {
		t.Fatal("expected positive width for line")
	}

	// Render to buffer
	buf := buffer.NewBuffer(40, 2)
	line.Render(buf, 0, 0, 40)

	// Export to ANSI
	ansi := line.ANSI()
	if !strings.Contains(ansi, "Title:") || !strings.Contains(ansi, "Value") {
		t.Fatalf("expected ANSI string to contain line text, got: %s", ansi)
	}
}

func TestContainers(t *testing.T) {
	box := VBox(
		Fixed(1, Text("Header")),
		Flex(1, HBox(
			Percent(50, Text("Left")),
			Percent(50, Text("Right")),
		)),
		Fixed(1, Text("Footer")),
	)

	buf := buffer.NewBuffer(40, 10)
	box.Draw(buf, buffer.NewRect(0, 0, 40, 10))

	// Test Padding and Center
	padded := Pad(1, Text("Padded"))
	padded.Draw(buf, buffer.NewRect(0, 0, 10, 5))

	centered := Center(6, 1, Text("Center"))
	centered.Draw(buf, buffer.NewRect(0, 0, 20, 5))
}

func TestComponents(t *testing.T) {
	card := NewCard("System Status", Text("All services operational")).
		SetSubtitle("Production").
		SetFooter(Text("Last checked: Just now"))

	buf := buffer.NewBuffer(40, 8)
	card.Draw(buf, buffer.NewRect(0, 0, 40, 8))

	stat := NewStatCard("LATENCY", "14 ms", "2.1 ms faster").
		SetUp(true).
		SetAccent(cell.ColorHex("#00FFAA"))
	stat.Draw(buf, buffer.NewRect(0, 0, 20, 5))

	badge := NewBadge("ONLINE", cell.ColorHex("#000000"), cell.ColorHex("#00FFAA"))
	badge.Draw(buf, buffer.NewRect(0, 0, 10, 1))

	kh := NewKeyHints(
		KeyHint{Key: "Enter", Desc: "Select"},
		KeyHint{Key: "Esc", Desc: "Back"},
	)
	kh.Draw(buf, buffer.NewRect(0, 0, 40, 1))

	// Test Card header color
	headerFg := cell.ColorHex("#FF00FF")
	card2 := NewCard("Custom Header", Text("Content")).SetHeaderColor(headerFg)
	buf2 := buffer.NewBuffer(30, 5)
	card2.Draw(buf2, buffer.NewRect(0, 0, 30, 5))
	// Cell at (3, 0) should have the header foreground color
	c := buf2.Cell(3, 0)
	if c == nil || c.Fg != headerFg.Value {
		t.Fatalf("expected card title cell to have header color %v, got %v", headerFg.Value, c.Fg)
	}

	// Test ToastManager with zero maxShow and small screen
	tm := &ToastManager{}
	tm.Info("Title", "Message")
	if tm.Count() != 1 {
		t.Fatalf("expected 1 toast with zero maxShow, got %d", tm.Count())
	}
	smallBuf := buffer.NewBuffer(5, 3)
	tm.Draw(smallBuf, buffer.NewRect(0, 0, 5, 3)) // Should not panic
}

func TestRenderStream(t *testing.T) {
	var out bytes.Buffer
	rs := NewRenderStream(&out)

	rs.Println("Hello CLI")
	rs.Progress("Downloading", 0.5)
	rs.Progress("Over 100", 1.5)
	rs.Progress("Negative", -0.5)
	rs.Done("Completed")

	sp := widgets.NewSpinner(widgets.SpinnerDots, "Loading...")
	rs.Spinner("Work in progress", sp)
	rs.ClearLine()

	content := out.String()
	if !strings.Contains(content, "Hello CLI") || !strings.Contains(content, "Downloading") {
		t.Fatalf("expected stream output to contain keywords, got: %s", content)
	}
}

func TestFormField(t *testing.T) {
	field := NewFormField("Username", nil).
		SetRequired(true).
		SetHelperText("Choose a unique username")

	if field.Label() != "Username" {
		t.Fatalf("expected label Username, got %s", field.Label())
	}
	if !field.Required() {
		t.Fatal("expected field to be required")
	}
	if field.HelperText() != "Choose a unique username" {
		t.Fatalf("expected helper text, got %s", field.HelperText())
	}

	// Initially empty -> invalid because Required rule was added
	if field.IsValid() {
		t.Fatal("expected invalid when empty")
	}
	if field.ErrorMessage() != "This field is required" {
		t.Fatalf("expected required error message, got %s", field.ErrorMessage())
	}

	// Set value
	field.SetValue("alice")
	if !field.IsValid() {
		t.Fatalf("expected valid for alice, got %s", field.ErrorMessage())
	}
	if field.Value() != "alice" {
		t.Fatalf("expected alice, got %s", field.Value())
	}

	// Focus and Blur
	field.Focus()
	if !field.Focused() {
		t.Fatal("expected focused")
	}
	field.Blur()
	if field.Focused() {
		t.Fatal("expected blurred")
	}

	// Multiline draw (height 3)
	buf3 := buffer.NewBuffer(40, 3)
	field.Draw(buf3, buffer.NewRect(0, 0, 40, 3))

	// Multiline draw with error (height 3)
	field.SetValue("")
	field.Validate()
	buf3Err := buffer.NewBuffer(40, 3)
	field.Draw(buf3Err, buffer.NewRect(0, 0, 40, 3))

	// Line 2 should have the error icon
	hasErrRune := false
	for x := 0; x < 40; x++ {
		c := buf3Err.Cell(x, 2)
		if c != nil && c.Rune == '✖' {
			hasErrRune = true
			break
		}
	}
	if !hasErrRune {
		t.Fatal("expected error symbol on line 2 of multiline FormField")
	}

	// 2-line draw (height 2)
	buf2 := buffer.NewBuffer(40, 2)
	field.Draw(buf2, buffer.NewRect(0, 0, 40, 2))

	// 1-line draw (height 1)
	buf1 := buffer.NewBuffer(40, 1)
	field.Draw(buf1, buffer.NewRect(0, 0, 40, 1))

	// HandleKey and HandleMouse
	field.Focus()
	field.HandleKey(input.Key{Type: input.KeyRune, Rune: 'X'})
	if field.Value() != "X" {
		t.Fatalf("expected value 'X' after HandleKey, got '%s'", field.Value())
	}
	rect := buffer.NewRect(0, 0, 40, 3)
	handled := field.HandleMouse(tea.MouseMsg{Mouse: input.Mouse{Action: input.MousePress, Button: input.MouseLeft, X: 5, Y: 1}}, rect)
	if !handled {
		t.Fatal("expected HandleMouse to handle mouse press on input")
	}
}

