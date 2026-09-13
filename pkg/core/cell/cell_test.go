package cell

import (
	"testing"
	"unsafe"
)

func TestCellSizeAndAlignment(t *testing.T) {
	var c Cell
	size := unsafe.Sizeof(c)
	if size != 16 {
		t.Fatalf("Expected Cell size to be exactly 16 bytes, got %d", size)
	}

	align := unsafe.Alignof(c)
	if align != 4 {
		t.Fatalf("Expected Cell alignment to be 4 bytes, got %d", align)
	}
}

func TestModifierBitmask(t *testing.T) {
	var m Modifier
	if m.Has(AttrBold) {
		t.Errorf("Expected false for unset AttrBold")
	}
	m.Add(AttrBold)
	m.Add(AttrUnderline)
	if !m.Has(AttrBold) || !m.Has(AttrUnderline) {
		t.Errorf("Expected both Bold and Underline set")
	}
	if m.Has(AttrItalic) {
		t.Errorf("AttrItalic should not be set")
	}
	m.Remove(AttrBold)
	if m.Has(AttrBold) {
		t.Errorf("AttrBold should have been removed")
	}
	if !m.Has(AttrUnderline) {
		t.Errorf("AttrUnderline should still be set")
	}
}

func TestCellEquality(t *testing.T) {
	c1 := NewCell('A', 1)
	c1.FgType = ColorRGB
	c1.Fg = 0x00FFAABB
	c1.Modifier = AttrBold

	c2 := NewCell('A', 1)
	c2.FgType = ColorRGB
	c2.Fg = 0x00FFAABB
	c2.Modifier = AttrBold

	if !c1.Equal(c2) {
		t.Errorf("Expected c1 to equal c2")
	}

	c2.Modifier.Add(AttrItalic)
	if c1.Equal(c2) {
		t.Errorf("Expected c1 != c2 when modifier differs")
	}
}

func TestColorHex(t *testing.T) {
	tests := []struct {
		hex      string
		expected Color
	}{
		{"#FF0000", RGB(255, 0, 0)},
		{"#00FF00", RGB(0, 255, 0)},
		{"#0000FF", RGB(0, 0, 255)},
		{"#FFF", RGB(255, 255, 255)},
		{"#000", RGB(0, 0, 0)},
		{"7D56F4", RGB(0x7D, 0x56, 0xF4)},
	}

	for _, tt := range tests {
		c := ColorHex(tt.hex)
		if c != tt.expected {
			t.Errorf("ColorHex(%q) = %+v; want %+v", tt.hex, c, tt.expected)
		}
	}
}

func TestColorDownsampling(t *testing.T) {
	white := RGB(255, 255, 255)
	c256 := white.To256()
	if c256.Type != ColorANSI256 {
		t.Errorf("Expected ColorANSI256, got %d", c256.Type)
	}

	c16 := white.To16()
	if c16.Type != ColorANSI16 {
		t.Errorf("Expected ColorANSI16, got %d", c16.Type)
	}
	if c16.Value != 15 && c16.Value != 7 { // BrightWhite or White
		t.Errorf("Expected White/BrightWhite ANSI index, got %d", c16.Value)
	}
}

func BenchmarkCellEqual(b *testing.B) {
	c1 := NewCell('A', 1)
	c2 := NewCell('A', 1)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if !c1.Equal(c2) {
			b.Fatal("unexpected inequality")
		}
	}
}

func BenchmarkColorHex(b *testing.B) {
	s := "#7D56F4"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ColorHex(s)
	}
}
