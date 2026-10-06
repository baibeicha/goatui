package widgets

import (
	"testing"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/driver/input"
)

func TestColorPicker_HSVConversion(t *testing.T) {
	// Pure Red
	h, s, v := RGBToHSV(255, 0, 0)
	if h != 0 || s != 1.0 || v != 1.0 {
		t.Fatalf("expected (0, 1, 1), got (%f, %f, %f)", h, s, v)
	}

	rgb := HSVToRGB(h, s, v)
	if rgb.R() != 255 || rgb.G() != 0 || rgb.B() != 0 {
		t.Fatalf("expected pure red back, got %v", rgb)
	}

	// Hex parsing
	c, ok := HexToRGBColor("#FF5500")
	if !ok || c.R() != 255 || c.G() != 85 || c.B() != 0 {
		t.Fatalf("expected parsed hex #FF5500, got %v", c)
	}
}

func TestColorPicker_Interaction(t *testing.T) {
	cp := NewColorPickerModal()
	applied := false
	cp.OpenForColor("editor.bg", "Background", "#1E1E2E", func(key, hexVal string) {
		applied = true
	})

	if !cp.Open {
		t.Fatal("expected color picker to be open")
	}

	// Hit Enter to apply
	handled, shouldClose := cp.HandleKey(input.Key{Type: input.KeyEnter})
	if !handled || !shouldClose || !applied {
		t.Fatal("expected enter to apply and close")
	}

	// Render check
	buf := buffer.NewBuffer(80, 24)
	cp.Open = true
	cp.Render(buf, 80, 24, DefaultColorPickerTheme())
	if buf.Cell(15, 5) == nil {
		t.Fatal("expected rendered cell in color picker")
	}
}
