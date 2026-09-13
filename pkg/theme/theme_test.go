package theme

import (
	"testing"

	"github.com/baibeicha/goatui/pkg/core/cell"
)

func TestParseThemeYAML(t *testing.T) {
	yamlData := `
name: "Cyberpunk"

colors:
  primary: "#FF007F"
  secondary: "#00F0FF"
  background: "#050505"
  foreground: "#EAEAEA"
  success: "#00FF66"
  danger: "#FF2A2A"

borders:
  style: "rounded"
`
	th, err := ParseThemeYAML(yamlData)
	if err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	if th.Name != "Cyberpunk" {
		t.Errorf("Expected name 'Cyberpunk', got %q", th.Name)
	}

	if th.Colors.Primary != cell.ColorHex("#FF007F") {
		t.Errorf("Primary color mismatch")
	}

	if th.Colors.Secondary != cell.ColorHex("#00F0FF") {
		t.Errorf("Secondary color mismatch")
	}

	if th.Borders.Name != "rounded" {
		t.Errorf("Expected border style 'rounded', got %q", th.Borders.Name)
	}
}

func TestThemeManager(t *testing.T) {
	tm := NewThemeManager()

	// Switch to Nord
	th, ok := tm.SetTheme("nord")
	if !ok {
		t.Fatalf("Expected Nord theme to exist")
	}
	if th.Name != "Nord" {
		t.Errorf("Expected Nord theme name, got %q", th.Name)
	}
	if tm.Current().Name != "Nord" {
		t.Errorf("Expected Current() to be Nord")
	}

	// Switch to Dracula
	th, ok = tm.SetTheme("dracula")
	if !ok || th.Name != "Dracula" {
		t.Errorf("Failed to switch to Dracula")
	}

	// Test GoatDark both with and without space
	th, ok = tm.SetTheme("GoatDark")
	if !ok || th.Name != "Goat Dark" {
		t.Errorf("Failed to switch to GoatDark (normalized)")
	}

	th, ok = tm.SetTheme("Goat Dark")
	if !ok || th.Name != "Goat Dark" {
		t.Errorf("Failed to switch to Goat Dark (with space)")
	}

	th, ok = tm.SetTheme("catppuccinmocha")
	if !ok || th.Name != "Catppuccin Mocha" {
		t.Errorf("Failed to switch to Catppuccin Mocha (normalized)")
	}

	// Non-existent theme
	_, ok = tm.SetTheme("non-existent-theme-xyz")
	if ok {
		t.Errorf("Expected false for non-existent theme")
	}
}
