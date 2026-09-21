---
name: goatui-app-creator
description: >-
  Build high-performance, modern, Elm-inspired Terminal User Interface (TUI)
  applications in Go using the GoatUI framework. Use when the user asks to
  create, scaffold, design, or implement TUI apps, dashboards, monitoring
  tools, file explorers, or terminal widgets using goatui.
---

# GoatUI TUI Application Creator

## Overview

GoatUI (`github.com/baibeicha/goatui`) is a high-performance, modern, Elm-inspired Terminal User Interface framework for Go. It features:
- **The Elm Architecture (TEA)**: Unidirectional `Init -> Update -> View` data flow.
- **Dual-Buffer Radix Diff Renderer**: Flat contiguous memory layout (zero pointer chasing), Mode 2026 synchronized output, ANSI escape minimization, TrueColor RGB / 256-color / 16-color support.
- **Flexbox Layout Engine**: Proportional division with `Fixed`, `Percent`, `Flex`, `Min`, `Max`, `Ratio` constraints.
- **Rich Interactive Widget Suite**: `VirtualTable` (100k+ virtual rows), `VirtualList` (1M+ items), `TextInput` (with password mode, readline navigation, and bounds clipping), `Tabs`, `TreeView`, `Checkbox`, `Gauge`, `Sparkline`, `BrailleCanvas`, `ImageWidget`, `VideoPlayerWidget`.
- **Full Window & Route Management**: Stack navigation, URL routing (`goat://...`, `file://...`, `http://...`), modals (`AlertModal`, `ConfirmModal`, `InputModal`, `PasswordModal`), and `Omnibar` command palette with fzf fuzzy matching.
- **Theme Engine**: Runtime hot-reloading with built-in presets (`GoatDark`, `Dracula`, `CatppuccinMocha`, `Nord`, `Monokai`, `Cyberpunk`, `Matrix`, `Forest`).
- **Role-Based Access Control (RBAC)**: Route guards (`RequireAuth`, `RequireRole`, `RequirePermission`).
- **Physics & Animations**: Spring dynamics, easing curves, particle effects.

---

## Dependencies

In your application's `go.mod`:
```go
module myapp

go 1.22

require github.com/baibeicha/goatui v0.0.0
```

---

## Quick Start: Minimal TEA Application

Every GoatUI application implements `tea.Model`:

```go
package main

import (
	"fmt"
	"os"

	"github.com/baibeicha/goatui"
)

type model struct {
	counter int
}

func (m model) Init() goatui.Cmd {
	return nil
}

func (m model) Update(msg goatui.Msg) (goatui.Model, goatui.Cmd) {
	switch msg := msg.(type) {
	case goatui.KeyMsg:
		switch msg.Key.Type {
		case goatui.KeyEsc:
			return m, goatui.Quit
		case goatui.KeyUp:
			m.counter++
		case goatui.KeyDown:
			m.counter--
		case goatui.KeyRune:
			switch msg.Key.Rune {
			case 'q', 'Q':
				return m, goatui.Quit
			case '+', '=':
				m.counter++
			case '-':
				m.counter--
			}
		}
	}
	return m, nil
}

func (m model) View(f *goatui.Frame) {
	area := f.Area()
	if area.IsEmpty() {
		return
	}

	curTheme := goatui.DefaultTheme().Current()
	p := curTheme.Colors

	// Background fill
	f.Buffer.Fill(area, goatui.Cell{
		Rune:   ' ',
		Width:  1,
		FgType: goatui.DefaultColor().Type,
		BgType: p.Background.Type,
		Bg:     p.Background.Value,
	})

	// Centered card container
	cardW := min(46, area.Width-4)
	cardH := min(12, area.Height-2)
	cardArea := goatui.NewRect(
		area.X+max(0, (area.Width-cardW)/2),
		area.Y+max(0, (area.Height-cardH)/2),
		cardW,
		cardH,
	)

	cardStyle := goatui.NewStyle().
		Border(curTheme.Borders.Border).
		BorderForeground(p.Primary).
		Background(p.Background).
		AlignHorizontal(goatui.AlignCenter).
		AlignVertical(goatui.AlignMiddle)

	content := fmt.Sprintf(
		"🐐 Counter Value: %d\n\n[↑ / +] Increment\n[↓ / -] Decrement\n[Q / Esc] Quit",
		m.counter,
	)
	cardStyle.Draw(f.Buffer, cardArea, content)
}

func main() {
	p := goatui.NewProgram(model{})
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
```

---

## Layout System (Flexbox Engine)

Divide rectangular screens horizontally or vertically using constraints:

```go
// Split screen into Header (3 lines), Content (Flex), and Footer (1 line)
rows := goatui.SplitVertical(area,
    goatui.Fixed(3), // Exact cell count
    goatui.Flex(1),  // Proportional flex
    goatui.Fixed(1), // Exact cell count
)
headerArea  := rows[0]
contentArea := rows[1]
footerArea  := rows[2]

// Split Content horizontally into Sidebar (30%) and Main Panel (70%)
cols := goatui.SplitHorizontal(contentArea,
    goatui.Percent(30), // 30% of available width
    goatui.Flex(1),     // Remainder
)
sidebarArea := cols[0]
mainArea    := cols[1]
```

### Supported Constraints
- `goatui.Fixed(n)`: Exactly `n` terminal cells.
- `goatui.Percent(p)`: `p`% of available length.
- `goatui.Flex(factor)`: Proportional share of remaining space.
- `goatui.Min(n)`: At least `n` cells.
- `goatui.Max(n)`: At most `n` cells.
- `goatui.Ratio(num, den)`: Fractional length (`num / den`).

---

## Core Widgets Guide

### 1. VirtualTable (Massive Datasets & High Performance)
VirtualTable can effortlessly handle 100,000+ rows with zero allocations on scrolling:

```go
cols := []goatui.TableColumn{
    {Title: "ID", Width: 8, Align: goatui.AlignRight},
    {Title: "Service", Flex: 1, MinWidth: 16, Align: goatui.AlignLeft},
    {Title: "Status", Width: 14, Align: goatui.AlignCenter},
}

table := goatui.NewTable(cols).
    SetTotalRows(50_000).
    SetBorder(goatui.TableBorderClean).
    SetZebra(true).
    SetSelectionPrefix("> ")

// On-demand row provider (called only for visible rows in viewport)
table.SetRowProvider(func(row int) []string {
    status := "RUNNING"
    if row%4 == 0 { status = "IDLE" }
    return []string{fmt.Sprintf("#%05d", row), fmt.Sprintf("worker-%d", row%10), status}
})

// Custom cell badge renderer
table.SetCellRenderer(func(buf *goatui.Buffer, col goatui.TableColumn, cellRect goatui.Rect, row, colIdx int, sel bool) bool {
    if colIdx == 2 { // Status
        status := "● ACTIVE"
        fg := goatui.ColorHex("#00FFAA")
        if row%4 == 0 {
            status = "○ IDLE"
            fg = goatui.ColorHex("#FFB86C")
        }
        goatui.RenderCellText(buf, cellRect, status, goatui.AlignCenter, 1, fg, goatui.DefaultColor(), goatui.AttrBold)
        return true
    }
    return false
})

// Handling table navigation and mouse events in Update:
table.HandleKey(keyMsg.Key, 10) // Supports Up/Down, PgUp/PgDn, Home/End, j/k/g/G
table.HandleMouse(mouseMsg, tableArea) // Mouse wheel scroll & row selection click

// VirtualList has identical navigation parity:
list := goatui.NewVirtualList(1_000_000, renderer)
list.HandleKey(keyMsg.Key, 10)
list.HandleMouse(mouseMsg, listArea)
```

### 2. TextInput (Interactive Text & Password Mode)
TextInput supports readline navigation shortcuts, password masking, placeholder ghost text, clipboard pasting, and click-to-position:

```go
input := goatui.NewTextInput()
input.SetPrompt("Password: ")
input.SetPlaceholder("Enter secret key...")
input.SetPasswordMode(true) // Mask characters as '•'
input.Focus()

// In Update:
switch msg := msg.(type) {
case goatui.KeyMsg:
    if input.HandleKey(msg.Key) {
        // Value changed or cursor moved
    }
case goatui.PasteMsg:
    input.InsertString(msg.Text) // Paste multiline or single-line text cleanly
case goatui.MouseMsg:
    input.HandleMouse(msg, inputArea)
}

// In View:
input.Draw(f.Buffer, inputArea)
```
**Built-in Keyboard Shortcuts**:
- Word left / right: `Ctrl+Left` / `Ctrl+Right` (or `Alt+Left` / `Alt+Right`) — supports full Unicode (Latin, Cyrillic, CJK)
- Word delete: `Ctrl+Backspace` / `Ctrl+W`
- Beginning / end of line: `Ctrl+A` / `Ctrl+E` (or `Home` / `End`)
- Clear before cursor: `Ctrl+U`
- Clear after cursor: `Ctrl+K`
- Paste: `tea.PasteMsg` via `input.InsertString(msg.Text)`
- Modifiers inspection: `key.HasCtrl()`, `key.HasAlt()`, `key.HasShift()`, `mouse.HasCtrl()`

### 3. Tabs (Navigation Bar)
Tabs render horizontal tab items with hotkeys and badges:

```go
tabs := goatui.NewTabs(
    goatui.TabItem{ID: "/overview", Title: "Overview", Hotkey: '1'},
    goatui.TabItem{ID: "/metrics",  Title: "Metrics",  Hotkey: '2', Badge: "LIVE"},
    goatui.TabItem{ID: "/settings", Title: "Settings", Hotkey: '3'},
)

// In Update:
tabs.HandleKey(keyMsg.Key)
tabs.HandleMouse(mouseMsg)

// In View:
tabs.Draw(f.Buffer, navArea)
```

### 4. Gauge & Sparkline (Live Telemetry & Metrics)
```go
gauge := goatui.NewGauge().
    SetPercent(0.82).
    SetFg(goatui.ColorHex("#00FFAA")).
    SetBg(goatui.ColorHex("#222233")).
    SetAlign(goatui.AlignCenter)
gauge.Draw(f.Buffer, gaugeArea)

sparkline := goatui.NewSparkline([]float64{10, 24, 45, 80, 75, 90, 60, 40}).
    SetFg(goatui.ColorHex("#00D2FF"))
sparkline.Draw(f.Buffer, sparkArea)
```

### 5. BrailleCanvas (Subpixel Vector Graphics)
Provides 2x4 subpixels per terminal cell with Bresenham line drawing:

```go
canvas := goatui.NewBrailleCanvas(40, 10) // 80x40 virtual subpixels
canvas.DrawLine(0, 0, 79, 39)
canvas.SetPixel(40, 20)
canvas.Draw(f.Buffer, canvasArea)
```

### 6. Select / Dropdown (Overlay Popups)
Select renders a single-choice dropdown that opens a floating popup menu using GoatUI's overlay rendering pass (`buf.AddOverlay`), ensuring dropdown items are never clipped by container boundaries and never overwritten by subsequent sibling widgets:

```go
selector := goatui.NewSelect("env", "Active Environment",
    goatui.SelectItem{ID: "prod", Label: "Production Cluster"},
    goatui.SelectItem{ID: "stage", Label: "Staging Sandbox"},
    goatui.SelectItem{ID: "local", Label: "Local Simulation"},
)

// In Update:
selector.HandleKey(keyMsg.Key)
selector.HandleMouse(mouseMsg)

// In View:
selector.Draw(f.Buffer, selectArea)
```

### 7. Declarative Input Validation & FormField (`pkg/validation` & `ui.FormField`)
GoatUI provides a declarative, reactive validation engine with built-in rules, automatic visual error feedback, and Form integration:

```go
// 1. Direct TextInput validation:
input := goatui.NewTextInput()
input.AddValidation(
    goatui.RuleRequired("Username is required"),
    goatui.RuleMinLength(3, "Must be at least 3 characters"),
    goatui.RuleAlphanumeric("Must contain only letters and digits"),
)
input.SetValidateOnChange(true) // Re-validate on every keystroke

// 2. Declarative FormField combining Label, Input, and Error:
field := goatui.NewFormField("Email Address", input).
    SetRequired(true).
    SetHelperText("We will never share your email.")
field.AddValidation(
    goatui.RuleRequired(),
    goatui.RuleEmail(),
)

// In Update:
field.HandleKey(keyMsg.Key)
field.HandleMouse(mouseMsg, fieldArea)

// In View:
field.Draw(f.Buffer, fieldArea)

// 3. Multi-field Form validation:
form := goatui.NewForm().
    Field("username", goatui.RuleRequired(), goatui.RuleMinLength(3)).
    Field("email", goatui.RuleRequired(), goatui.RuleEmail()).
    Field("age", goatui.RuleOptional(goatui.RuleIntRange(18, 120)))

res := form.Validate(map[string]string{
    "username": "alice",
    "email": "invalid-email",
})
if !res.IsValid() {
    fmt.Println("Error:", res.Error("email"))
}
```
**Built-in Validation Rules**:
- `goatui.RuleRequired(msg...)`
- `goatui.RuleMinLength(min, msg...)` / `goatui.RuleMaxLength(max, msg...)`
- `goatui.RuleLengthRange(min, max, msg...)`
- `goatui.RuleIntRange(min, max, msg...)` / `goatui.RuleFloatRange(min, max, msg...)`
- `goatui.RuleRegex(pattern, msg...)`
- `goatui.RuleEmail(msg...)` / `goatui.RuleURL(msg...)`
- `goatui.RuleNumeric(msg...)` / `goatui.RuleAlpha(msg...)` / `goatui.RuleAlphanumeric(msg...)`
- `goatui.RuleCustom(fn)`
- `goatui.RuleOptional(rule)`

---

## Desktop-Grade Architecture: WindowManager, Router & Omnibar

For complex multi-screen applications, use `WindowManager`:

```go
package main

import (
	"github.com/baibeicha/goatui"
)

func main() {
	r := goatui.NewRouter()

	// Register Routes
	r.Handle("/dashboard", func(ctx *goatui.RouteContext) any {
		return NewDashboardScreen()
	})
	r.Handle("/settings", func(ctx *goatui.RouteContext) any {
		return NewSettingsScreen()
	}, goatui.RequireAuth("/login")) // Protected route

	wm := goatui.NewWindowManager(r)
	wm.Push(NewDashboardScreen(), nil)

	// Ctrl+K automatically toggles the fzf Omnibar command palette!
	prog := goatui.NewProgram(wm)
	prog.Run()
}
```

### Modals & Dialogs
```go
// Show an input modal
wm.ShowModal(goatui.NewInputModal(
    "Rename File",
    "Enter new filename:",
    "current.txt",
    func(val string) goatui.Cmd {
        // Confirmed
        return nil
    },
    func() goatui.Cmd {
        // Cancelled
        return nil
    },
))

// Show password dialog
wm.ShowModal(goatui.NewPasswordModal(
    "Authentication",
    "Enter Root Password:",
    func(pass string) goatui.Cmd {
        // Authenticate
        return nil
    },
    nil,
))
```

---

## Theming & Live Hot-Reloading

GoatUI comes with rich TrueColor themes out of the box:
- `goatui.GoatDarkTheme`
- `goatui.DraculaTheme`
- `goatui.CatppuccinTheme`
- `goatui.NordTheme`
- `goatui.MonokaiTheme`
- `goatui.CyberpunkTheme`
- `goatui.MatrixTheme`
- `goatui.ForestTheme`

### Switching Themes at Runtime:
```go
goatui.SwitchTheme("Cyberpunk")
```

### Loading External YAML Themes:
```yaml
name: HackerNeon
borders: rounded
colors:
  primary: "#00FF66"
  secondary: "#008833"
  background: "#050A05"
  foreground: "#CCFFCC"
  success: "#00FF88"
  warning: "#FFCC00"
  danger: "#FF0033"
  info: "#00CCFF"
  muted: "#336633"
  accent: "#55FF99"
```

```go
goatui.DefaultTheme().LoadFile("themes/custom.yaml")
```

---

## Headless Testing with `testkit.MockDriver`

Write automated regression tests for your UI without opening a terminal:

```go
package main_test

import (
	"testing"
	"github.com/baibeicha/goatui"
	"github.com/baibeicha/goatui/pkg/testkit"
)

func TestAppNavigation(t *testing.T) {
	mockDrv := testkit.NewMockDriver(80, 24)
	app := initialModel()

	prog := goatui.NewProgram(app, goatui.WithDriver(mockDrv))

	// Send key events programmatically
	mockDrv.SendRune('+')
	mockDrv.SendKey(goatui.KeyEnter, 0, 0)

	// Send quit
	mockDrv.SendKey(goatui.KeyEsc, 0, 0)

	finalModel, err := prog.Run()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	m := finalModel.(myModel)
	if m.counter != 1 {
		t.Errorf("expected counter 1, got %d", m.counter)
	}
}
```

---

## Common Mistakes & Best Practices

1. **Unclipped Strings**: Never call `buf.SetString(x, y, s, ...)` with long or untrusted strings without boundary checking. Use `buf.SetStringAligned(rect, s, align, ...)` or calculate visual rune width with `goatui.StringWidth(s)` to avoid overwriting borders.
2. **Byte Slicing vs Rune Slicing**: In Go, `s[:n]` slices BYTES, not characters! Truncating UTF-8 strings by byte count will slice in the middle of Cyrillic/CJK characters and panic. Always convert to `[]rune` or use `SetStringAligned` for truncation.
3. **Double Redraws**: GoatUI uses double-buffered radix diffing. Do NOT manually emit ANSI escape codes or cursor sequences inside `View()`; simply modify the `f.Buffer`.
4. **Blocking Commands in `Update`**: Never execute HTTP requests, file I/O, or sleeps directly in `Update()`. Wrap them in `tea.Cmd` (`func() tea.Msg`) so they run in background goroutines without freezing the UI.
5. **Mouse Click Handling**: Check `msg.IsLeftClick()` on `HitMsg` or `msg.Action == goatui.MousePress && msg.Button == goatui.MouseLeft` on `MouseMsg` to avoid handling both button press and release as duplicate clicks.
