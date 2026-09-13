package theme

import (
	"bufio"
	"strings"

	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/style"
	"github.com/baibeicha/goatui/pkg/widgets"
)

// ColorPalette defines the core semantic colors for a theme.
type ColorPalette struct {
	Primary    cell.Color
	Secondary  cell.Color
	Background cell.Color
	Foreground cell.Color
	Success    cell.Color
	Warning    cell.Color
	Danger     cell.Color
	Info       cell.Color
	Muted      cell.Color
	Accent     cell.Color
}

// BorderConfig contains resolved border styles for windows and tables.
type BorderConfig struct {
	Name        string
	Border      style.Border
	TableBorder widgets.TableBorder
}

// Theme encapsulates semantic colors and borders.
type Theme struct {
	Name    string
	Colors  ColorPalette
	Borders BorderConfig
}

// Built-in Theme Presets
var (
	Dracula = Theme{
		Name: "Dracula",
		Colors: ColorPalette{
			Primary:    cell.ColorHex("#BD93F9"),
			Secondary:  cell.ColorHex("#6272A4"),
			Background: cell.ColorHex("#282A36"),
			Foreground: cell.ColorHex("#F8F8F2"),
			Success:    cell.ColorHex("#50FA7B"),
			Warning:    cell.ColorHex("#F1FA8C"),
			Danger:     cell.ColorHex("#FF5555"),
			Info:       cell.ColorHex("#8BE9FD"),
			Muted:      cell.ColorHex("#6272A4"),
			Accent:     cell.ColorHex("#FF79C6"),
		},
		Borders: resolveBorder("rounded"),
	}

	CatppuccinMocha = Theme{
		Name: "Catppuccin Mocha",
		Colors: ColorPalette{
			Primary:    cell.ColorHex("#CBA6F7"),
			Secondary:  cell.ColorHex("#89B4FA"),
			Background: cell.ColorHex("#1E1E2E"),
			Foreground: cell.ColorHex("#CDD6F4"),
			Success:    cell.ColorHex("#A6E3A1"),
			Warning:    cell.ColorHex("#F9E2AF"),
			Danger:     cell.ColorHex("#F38BA8"),
			Info:       cell.ColorHex("#89DCEB"),
			Muted:      cell.ColorHex("#6C7086"),
			Accent:     cell.ColorHex("#F5C2E7"),
		},
		Borders: resolveBorder("rounded"),
	}

	Nord = Theme{
		Name: "Nord",
		Colors: ColorPalette{
			Primary:    cell.ColorHex("#88C0D0"),
			Secondary:  cell.ColorHex("#81A1C1"),
			Background: cell.ColorHex("#2E3440"),
			Foreground: cell.ColorHex("#ECEFF4"),
			Success:    cell.ColorHex("#A3BE8C"),
			Warning:    cell.ColorHex("#EBCB8B"),
			Danger:     cell.ColorHex("#BF616A"),
			Info:       cell.ColorHex("#5E81AC"),
			Muted:      cell.ColorHex("#4C566A"),
			Accent:     cell.ColorHex("#B48EAD"),
		},
		Borders: resolveBorder("normal"),
	}

	Monokai = Theme{
		Name: "Monokai",
		Colors: ColorPalette{
			Primary:    cell.ColorHex("#FD971F"),
			Secondary:  cell.ColorHex("#66D9EF"),
			Background: cell.ColorHex("#272822"),
			Foreground: cell.ColorHex("#F8F8F2"),
			Success:    cell.ColorHex("#A6E22E"),
			Warning:    cell.ColorHex("#E6DB74"),
			Danger:     cell.ColorHex("#F92672"),
			Info:       cell.ColorHex("#AE81FF"),
			Muted:      cell.ColorHex("#75715E"),
			Accent:     cell.ColorHex("#E6DB74"),
		},
		Borders: resolveBorder("normal"),
	}

	GoatDark = Theme{
		Name: "Goat Dark",
		Colors: ColorPalette{
			Primary:    cell.ColorHex("#00D2FF"),
			Secondary:  cell.ColorHex("#7D56F4"),
			Background: cell.ColorHex("#16161E"),
			Foreground: cell.ColorHex("#C0CAF5"),
			Success:    cell.ColorHex("#00FF88"),
			Warning:    cell.ColorHex("#FF9900"),
			Danger:     cell.ColorHex("#FF0055"),
			Info:       cell.ColorHex("#00B4D8"),
			Muted:      cell.ColorHex("#565F89"),
			Accent:     cell.ColorHex("#FF007F"),
		},
		Borders: resolveBorder("rounded"),
	}
)

func resolveBorder(name string) BorderConfig {
	switch strings.ToLower(name) {
	case "normal", "box", "square":
		return BorderConfig{
			Name:        "normal",
			Border:      style.BorderNormal,
			TableBorder: widgets.TableBorderNormal,
		}
	case "thick", "heavy":
		return BorderConfig{
			Name:        "thick",
			Border:      style.BorderThick,
			TableBorder: widgets.TableBorderNormal,
		}
	case "double":
		return BorderConfig{
			Name:        "double",
			Border:      style.BorderDouble,
			TableBorder: widgets.TableBorderNormal,
		}
	case "clean":
		return BorderConfig{
			Name:        "clean",
			Border:      style.BorderNormal,
			TableBorder: widgets.TableBorderClean,
		}
	case "none":
		return BorderConfig{
			Name:        "none",
			Border:      style.Border{},
			TableBorder: widgets.TableBorderNone,
		}
	default: // "rounded"
		return BorderConfig{
			Name:        "rounded",
			Border:      style.BorderRounded,
			TableBorder: widgets.TableBorderRounded,
		}
	}
}

// ParseThemeYAML parses a clean YAML / key-value theme configuration into a Theme struct.
func ParseThemeYAML(content string) (Theme, error) {
	t := GoatDark // Base fallback
	scanner := bufio.NewScanner(strings.NewReader(content))
	currentSection := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasSuffix(line, ":") {
			currentSection = strings.ToLower(strings.TrimSuffix(line, ":"))
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.ToLower(strings.TrimSpace(parts[0]))
		val := strings.Trim(strings.TrimSpace(parts[1]), "\"'")

		switch currentSection {
		case "colors":
			c := cell.ColorHex(val)
			switch key {
			case "primary":
				t.Colors.Primary = c
			case "secondary":
				t.Colors.Secondary = c
			case "background", "bg":
				t.Colors.Background = c
			case "foreground", "fg":
				t.Colors.Foreground = c
			case "success":
				t.Colors.Success = c
			case "warning":
				t.Colors.Warning = c
			case "danger":
				t.Colors.Danger = c
			case "info":
				t.Colors.Info = c
			case "muted":
				t.Colors.Muted = c
			case "accent":
				t.Colors.Accent = c
			}
		case "borders":
			if key == "style" {
				t.Borders = resolveBorder(val)
			}
		default:
			if key == "name" {
				t.Name = val
			} else if key == "border" || key == "borders" {
				t.Borders = resolveBorder(val)
			}
		}
	}

	return t, scanner.Err()
}
