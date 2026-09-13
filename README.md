# GoatUI 🐐

A high-performance, modern, Elm-inspired Terminal User Interface (TUI) framework for Go.

[![Go Report Card](https://goreportcard.com/badge/github.com/baibeicha/goatui)](https://goreportcard.com/report/github.com/baibeicha/goatui)
[![Go Reference](https://pkg.go.dev/badge/github.com/baibeicha/goatui.svg)](https://pkg.go.dev/github.com/baibeicha/goatui)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

GoatUI is built from scratch to provide a desktop-grade application experience inside standard and modern terminal emulators. It features a reactive unidirectional data flow (TEA), double-buffered diff rendering, TrueColor RGB support, physical animation controllers, high-fidelity media rendering (Sixel, Braille, Half-block), an fzf-style Omnibar, and built-in system service screens.

---

## ✨ Features

- **⚡ Reactive TEA Architecture**:
  - Unidirectional architecture (Init, Update, View) inspired by Elm and Bubbletea.
  - Concurrency-safe non-blocking command batching (	ea.Batch, 	ea.Sequence).
  - Automated terminal resize detection via signals and high-frequency polling.
- **🖥️ Dual-Buffer Diff Renderer**:
  - Radix diffing between front and back buffers to minimize ANSI escape sequences.
  - Full TrueColor (24-bit RGB) and 256-color palette support.
  - Accurate Unicode wcwidth calculations and strict rectangular clipping (Buffer.SetClip).
- **📐 Layout & Styling System**:
  - Flexbox-like layout engine with flexible and fixed directional constraints.
  - CSS-like box model with padding, margins, borders, and embedded border titles (─ Title ─).
- **🎨 YAML Theme Engine**:
  - Live runtime theme switching with file watcher hot-reloading (pkg/theme/watcher.go).
  - Ships with built-in dark and light themes (GoatDark, Matrix, Nord, Cyberpunk, Forest, Monokai).
- **🧭 Trie URL Router & Window Manager**:
  - Screen stack navigation with URL routing (goat://...), route parameters, and query parameters.
  - Multi-tier modal dialog overlay system (InputModal, confirmation modals).
  - Integrated **Omnibar** command palette with fzf-like fuzzy matching, score sorting, title highlighting, command history, and Tab autocompletion.
- **🛡️ Role-Based Access Control (RBAC)**:
  - Route guards and role checks (Admin, User, Guest).
  - Component-level security visibility and authorization checks.
- **🧩 Rich Widget Library**:
  - **VirtualTable**: Virtualized scrolling for hundreds of thousands of rows, mouse wheel support, and column sizing.
  - **TreeView**: Hierarchical collapsible tree with cyclic reference protection and mouse navigation.
  - **Tabs**: Responsive tab bar with hotkey navigation ([1..N]) and smart title truncation.
  - **TextInput**: Horizontal scrolling for long inputs, mouse click-to-position, and password masking.
  - **Gauge & Sparkline**: Color-contrasted progress bars and real-time sparkline telemetry graphs.
- **🎬 Physics & Animation Engine**:
  - Spring dynamics with configurable stiffness, damping, and mass.
  - 15+ Easing curves (Bounce, Elastic, EaseInOutQuad, Back, etc.).
  - Sequential and Parallel animation composition pipelines.
  - 2D Particle System for particle effects (confetti, sparks, smoke).
- **🖼️ Terminal Graphics & Media**:
  - Auto-negotiated protocol selector: Sixel (DEC DCS), Braille (2x4 subpixels with ITU-R BT.601 luminance thresholding), Half-block (TrueColor with bilinear and area-averaged box filter resampling), and ASCII fallback.
  - Aspect ratio correction factor for terminal fonts.
- **📁 Built-in Services**:
  - **Interactive File Explorer**: Full directory navigation, breadcrumb location bar, multi-file selection, cut/copy/paste (with cross-device volume fallback), file deletion, in-terminal file preview, and configurable icon modes (ASCII, Unicode, NerdFont, Emoji).
  - **Live Network Viewer**: Active TCP/UDP socket telemetry, process monitoring, and real-time refresh.

---

## 🚀 Quick Start

### Installation

`ash
go get -u github.com/baibeicha/goatui
`

### Minimal TEA Application

`go
package main

import (
	"fmt"
	"os"

	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/driver/input"
	"github.com/baibeicha/goatui/pkg/tea"
)

type model struct {
	counter int
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Key {
		case input.KeyEsc, input.KeyCtrlC:
			return m, tea.Quit
		case input.KeyUp, '+':
			m.counter++
		case input.KeyDown, '-':
			m.counter--
		}
	}
	return m, nil
}

func (m model) View() string {
	return fmt.Sprintf("GoatUI Counter: %d\nPress Up/Down to change, Esc to exit.", m.counter)
}

func main() {
	p := tea.NewProgram(model{})
	if err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
`

---

## 🕹️ Running Examples

### 1. Full Desktop Application
Experience the full power of GoatUI with an interactive multi-tab dashboard, file explorer, media player, live network viewer, and particle animations:

`ash
go run ./examples/desktop_app
`

**Key Controls in Desktop App:**
- 1 .. 6: Switch tabs (Dashboard, Explorer, Media Player, Network, Security, Settings)
- Ctrl+P / Ctrl+K: Open Omnibar command palette and URL navigator
- Tab: Auto-complete paths or navigate fields
- I: Cycle file explorer icon themes (ASCII -> Unicode -> NerdFont -> Emoji)
- Space: Launch particle confetti explosion on Dashboard
- Esc: Cancel modal or exit

### 2. Widget Showcase
Inspect individual widgets and layouts:

`ash
go run ./examples/showcase
`

---

## 🧪 Testing

Run all unit and integration tests with Go race detector:

`ash
go test -race -count=1 ./...
`

---

## 📁 Repository Structure

`
goatui/
├── examples/
│   ├── desktop_app/     # Full-featured multi-screen desktop application
│   └── showcase/        # Interactive widget showcase
├── pkg/
│   ├── animation/       # Physics springs, easings, particles, and sequences
│   ├── core/
│   │   ├── buffer/      # Memory cell buffer, clipping, and rune width
│   │   ├── cell/        # 24-bit TrueColor cells and text styling attributes
│   │   └── renderer/    # Double-buffered terminal diff engine
│   ├── driver/          # Cross-platform raw terminal driver (Windows & Unix)
│   ├── layout/          # Flexbox constraints and box positioning
│   ├── media/           # Sixel, Braille, Half-block media renderers
│   ├── router/          # Trie URL router and route context
│   ├── security/        # RBAC role management and permissions
│   ├── services/
│   │   ├── explorer/    # File explorer with multi-select and icon sets
│   │   └── network/     # Live network socket inspector
│   ├── spatial/         # Rectangles, 2D coordinates, and boundary math
│   ├── style/           # CSS-like styling, borders, and box decoration
│   ├── tea/             # Elm-inspired reactive runtime engine
│   ├── testkit/         # Virtual terminal test driver and assertions
│   ├── theme/           # YAML theme definitions and live file watcher
│   ├── widgets/         # VirtualTable, TreeView, Tabs, TextInput, Gauges
│   └── window/          # WindowManager, Screen life-cycle, Modals, Omnibar
├── themes/              # Built-in YAML color schemes
├── goatui.go            # High-level framework facade
└── README.md
`

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
