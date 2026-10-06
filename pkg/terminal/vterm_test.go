package terminal

import (
	"strings"
	"testing"

	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/driver/input"
)

func TestVTerm_ANSI_Decoding(t *testing.T) {
	vt := NewVTerm(80, 24)

	// Write simple text with CR LF
	_, _ = vt.Write([]byte("Hello, World!\r\nNext Line"))
	lines := vt.ContentLines()
	if !strings.HasPrefix(lines[0], "Hello, World!") {
		t.Fatalf("expected line 0 to start with 'Hello, World!', got %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "Next Line") {
		t.Fatalf("expected line 1 to start with 'Next Line', got %q", lines[1])
	}

	// Test Cursor Positioning: CUP \x1b[5;10H
	_, _ = vt.Write([]byte("\x1b[5;10HX"))
	if vt.CursorY != 4 || vt.CursorX != 10 {
		t.Fatalf("expected cursor at (10, 4), got (%d, %d)", vt.CursorX, vt.CursorY)
	}
	if vt.Screen[4][9].Char != 'X' {
		t.Fatalf("expected 'X' at (9, 4), got %c", vt.Screen[4][9].Char)
	}

	// Test SGR TrueColor RGB: \x1b[38;2;255;100;50m
	_, _ = vt.Write([]byte("\x1b[38;2;255;100;50mRGB\x1b[0m"))
	cellR := vt.Screen[4][10]
	if cellR.Char != 'R' {
		t.Fatalf("expected 'R', got %c", cellR.Char)
	}
	expectedFg := cell.RGB(255, 100, 50)
	if cellR.Fg != expectedFg {
		t.Fatalf("expected FG %v, got %v", expectedFg, cellR.Fg)
	}

	// Test SGR 256 colors: \x1b[38;5;208m
	_, _ = vt.Write([]byte("\x1b[38;5;208mC256\x1b[0m"))
	cell256 := vt.Screen[4][13]
	if cell256.Char != 'C' {
		t.Fatalf("expected 'C', got %c", cell256.Char)
	}
	if cell256.Fg.IsDefault() {
		t.Fatalf("expected non-default FG for 256 color")
	}

	// Test SGR Styles: Bold, Underline, Reverse
	_, _ = vt.Write([]byte("\x1b[1;4;7mStyled\x1b[0m"))
	cellStyle := vt.Screen[4][17]
	if cellStyle.Char != 'S' {
		t.Fatalf("expected 'S', got %c", cellStyle.Char)
	}
	if !cellStyle.Attr.Has(cell.AttrBold) || !cellStyle.Attr.Has(cell.AttrUnderline) || !cellStyle.Attr.Has(cell.AttrReverse) {
		t.Fatalf("expected bold+underline+reverse attributes, got %v", cellStyle.Attr)
	}

	// Test Clear Screen: \x1b[2J
	_, _ = vt.Write([]byte("\x1b[2J"))
	for y := 0; y < vt.Rows; y++ {
		for x := 0; x < vt.Cols; x++ {
			if vt.Screen[y][x].Char != ' ' {
				t.Fatalf("expected blank cell at (%d, %d), got %c", x, y, vt.Screen[y][x].Char)
			}
		}
	}
}

func TestVTerm_Scrollback_And_Wrapping(t *testing.T) {
	vt := NewVTerm(20, 5)

	for i := 0; i < 10; i++ {
		_, _ = vt.Write([]byte("Line\r\n"))
	}

	if len(vt.Scrollback) == 0 {
		t.Fatalf("expected scrollback to have saved lines, got %d", len(vt.Scrollback))
	}

	vt.Scroll(5)
	if vt.ScrollOffset != 5 {
		t.Fatalf("expected scroll offset 5, got %d", vt.ScrollOffset)
	}

	vt.Scroll(-10)
	if vt.ScrollOffset != 0 {
		t.Fatalf("expected scroll offset clamped to 0, got %d", vt.ScrollOffset)
	}
}

func TestVTerm_AlternateScreen(t *testing.T) {
	vt := NewVTerm(40, 10)
	_, _ = vt.Write([]byte("Main Screen Content"))

	// Enter Alternate Screen
	_, _ = vt.Write([]byte("\x1b[?1049h"))
	if !vt.inAltScreen {
		t.Fatal("expected to be in alternate screen buffer")
	}
	_, _ = vt.Write([]byte("Alt Screen Content"))
	linesAlt := vt.ContentLines()
	if !strings.HasPrefix(linesAlt[0], "Alt Screen Content") {
		t.Fatalf("expected alt screen content, got %q", linesAlt[0])
	}

	// Exit Alternate Screen
	_, _ = vt.Write([]byte("\x1b[?1049l"))
	if vt.inAltScreen {
		t.Fatal("expected to exit alternate screen buffer")
	}
	linesMain := vt.ContentLines()
	if !strings.HasPrefix(linesMain[0], "Main Screen Content") {
		t.Fatalf("expected main screen content restored, got %q", linesMain[0])
	}
}

func TestKeyToVT(t *testing.T) {
	// Simple Enter
	out := KeyToVT(input.Key{Type: input.KeyEnter})
	if len(out) != 1 || out[0] != '\r' {
		t.Fatalf("expected \\r for enter, got %v", out)
	}

	// Backspace
	out = KeyToVT(input.Key{Type: input.KeyBackspace})
	if len(out) != 1 || out[0] != 0x08 {
		t.Fatalf("expected 0x08 for backspace, got %v", out)
	}

	// Arrow Up
	out = KeyToVT(input.Key{Type: input.KeyUp})
	if string(out) != "\x1b[A" {
		t.Fatalf("expected \\x1b[A for up arrow, got %q", string(out))
	}
}
