package main

import (
	"fmt"
	"os"

	"github.com/baibeicha/goatui"
	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/driver/input"
	"github.com/baibeicha/goatui/pkg/tea"
	"github.com/baibeicha/goatui/pkg/ui"
	"github.com/baibeicha/goatui/pkg/widgets"
)

func main() {
	app := goatui.NewApp("GOATUI QUICK DASHBOARD")

	// Global status bar
	app.SetStatus("ONLINE • Cluster EU-Central", "v2.4.0-boost")
	app.SetKeyHints(
		ui.KeyHint{Key: "Tab", Desc: "Switch Tab"},
		ui.KeyHint{Key: "F", Desc: "Flatten"},
		ui.KeyHint{Key: "A", Desc: "Alert"},
		ui.KeyHint{Key: "P", Desc: "Prompt"},
		ui.KeyHint{Key: "T", Desc: "Toast"},
		ui.KeyHint{Key: "Q", Desc: "Quit"},
	)

	// State for Controls tab
	checkboxStyleGroup := widgets.NewCheckboxGroup(
		widgets.CheckboxItem{ID: "c1", Label: "Neural Engine", Checked: true},
		widgets.CheckboxItem{ID: "c2", Label: "High Freq Mode", Checked: false},
		widgets.CheckboxItem{ID: "c3", Label: "Audit Tracing", Checked: true, Indeterminate: true},
	).SetStyle(widgets.CheckboxCircle).SetTriState(true)

	radioGroup := widgets.NewRadioGroup(
		widgets.RadioItem{ID: "m1", Label: "Quasar-2 (Macro Gate)"},
		widgets.RadioItem{ID: "m2", Label: "Solaris-1 (Balanced)"},
		widgets.RadioItem{ID: "m3", Label: "Syzygy-1 (High Conviction)"},
	).SetStyle(widgets.RadioCircleFilled)

	slider := widgets.NewSlider("risk", "Risk Per Trade", 0.1, 5.0, 1.25).
		SetFormat("%.2f%%")

	selector := widgets.NewSelect("env", "Active Environment",
		widgets.SelectItem{ID: "prod", Label: "Production Cluster"},
		widgets.SelectItem{ID: "stage", Label: "Staging Sandbox"},
		widgets.SelectItem{ID: "local", Label: "Local Simulation"},
	)

	spinner := widgets.NewSpinner(widgets.SpinnerDots, "Synthesizing market orderbook...")

	// Action Buttons
	btn1 := widgets.NewButton("btn-save", "Sync Config", func() {
		app.ToastSuccess("SYNCHRONIZED", "Parameters applied to live engine")
	}).SetVariant(widgets.ButtonVariantPrimary).SetHotkey('s')

	btn2 := widgets.NewButton("btn-flatten", "Emergency Flatten", func() {
		app.Confirm("EMERGENCY FLATTEN", "Liquidate all positions at market price?", func() {
			app.ToastError("FLATTENED", "All positions liquidated")
		}, nil)
	}).SetVariant(widgets.ButtonVariantDanger).SetHotkey('f')

	// Histogram data
	hist := widgets.NewHistogram(
		widgets.HistogramBar{Label: "09:00", Value: 42, Color: cell.ColorHex("#00D2FF")},
		widgets.HistogramBar{Label: "10:00", Value: 85, Color: cell.ColorHex("#00FFAA")},
		widgets.HistogramBar{Label: "11:00", Value: 63, Color: cell.ColorHex("#FFB86C")},
		widgets.HistogramBar{Label: "12:00", Value: 95, Color: cell.ColorHex("#FF5555")},
		widgets.HistogramBar{Label: "13:00", Value: 71, Color: cell.ColorHex("#00D2FF")},
		widgets.HistogramBar{Label: "14:00", Value: 54, Color: cell.ColorHex("#BD93F9")},
	).SetBarWidth(5)

	// Tab 1: Telemetry & Overview
	app.AddTab("overview", "Telemetry Overview", func(f *tea.Frame, area buffer.Rect) {
		ui.VBox(
			// Row 1: KPI Stat Cards
			ui.Fixed(4, ui.HBox(
				ui.Percent(25, ui.NewStatCard("NET PnL (24h)", "+$12,450.80", "18.4% vs benchmark").SetUp(true)),
				ui.Percent(25, ui.NewStatCard("SATELLITE EXPOSURE", "$48,200.00", "Leverage: 2.1x").SetAccent(cell.ColorHex("#00D2FF"))),
				ui.Percent(25, ui.NewStatCard("SYSTEM LATENCY", "1.24 ms", "P99: 2.80 ms").SetUp(true).SetAccent(cell.ColorHex("#FFB86C"))),
				ui.Percent(25, ui.NewStatCard("CIRCUIT BREAKER", "ARMED / NORMAL", "0 breaches").SetAccent(cell.ColorHex("#00FFAA"))),
			)),
			// Row 2: Throughput Chart & Activity
			ui.Flex(1, ui.HBox(
				ui.Percent(60, ui.NewCard("Hourly Order Execution Volume (k req/s)", hist)),
				ui.Percent(40, ui.NewCard("Live Status", ui.ViewFunc(func(b *buffer.Buffer, r buffer.Rect) {
					spinner.Tick()
					spinner.Draw(b, buffer.NewRect(r.X, r.Y+1, r.Width, 1))

					div := widgets.NewHorizontalDivider("Recent Transactions")
					div.Draw(b, buffer.NewRect(r.X, r.Y+3, r.Width, 1))

					l1 := ui.NewLine(
						ui.BadgeSpan("FILL", cell.ColorHex("#000000"), cell.ColorHex("#00FFAA")),
						ui.Text(" BUY  2.40 BTC @ $68,450.00"),
					)
					l1.Render(b, r.X, r.Y+5, r.Width)

					l2 := ui.NewLine(
						ui.BadgeSpan("WARN", cell.ColorHex("#000000"), cell.ColorHex("#FFB86C")),
						ui.Text(" Basis carry spread tightened: 4.8 bps"),
					)
					l2.Render(b, r.X, r.Y+7, r.Width)
				}))),
			)),
		).Draw(f.Buffer, area)
	})

	// Tab 2: Interactive Controls
	app.AddTab("controls", "Controls & Tuning", func(f *tea.Frame, area buffer.Rect) {
		ui.VBox(
			ui.Flex(1, ui.HBox(
				ui.Percent(50, ui.NewCard("Execution Parameters", ui.ViewFunc(func(b *buffer.Buffer, r buffer.Rect) {
					// Checkboxes
					ui.Text("Engine Flags (Space to cycle, Click to toggle):").Draw(b, buffer.NewRect(r.X, r.Y, r.Width, 1))
					checkboxStyleGroup.Draw(b, buffer.NewRect(r.X, r.Y+2, r.Width, 4))

					// Radios
					ui.Text("Active Decision Model:").Draw(b, buffer.NewRect(r.X, r.Y+7, r.Width, 1))
					radioGroup.Draw(b, buffer.NewRect(r.X, r.Y+9, r.Width, 4))
				}))),
				ui.Percent(50, ui.NewCard("Thresholds & Action Buttons", ui.ViewFunc(func(b *buffer.Buffer, r buffer.Rect) {
					// Slider
					slider.Draw(b, buffer.NewRect(r.X, r.Y, r.Width, 1))

					// Select
					selector.Draw(b, buffer.NewRect(r.X, r.Y+3, r.Width, 1))

					// Action Buttons
					btn1.Draw(b, buffer.NewRect(r.X, r.Y+6, 20, 1))
					btn2.Draw(b, buffer.NewRect(r.X+24, r.Y+6, 26, 1))
				}))),
			)),
		).Draw(f.Buffer, area)
	})

	// Route keyboard and mouse to controls tab widgets
	app.SetTabOnKey("controls", func(key input.Key) bool {
		if selector.IsOpen() {
			return selector.HandleKey(key)
		}
		if checkboxStyleGroup.HandleKey(key) {
			return true
		}
		if radioGroup.HandleKey(key) {
			return true
		}
		if slider.HandleKey(key) {
			return true
		}
		if selector.HandleKey(key) {
			return true
		}
		if btn1.HandleKey(key) {
			return true
		}
		if btn2.HandleKey(key) {
			return true
		}
		return false
	})

	app.SetTabOnMouse("controls", func(msg tea.MouseMsg) bool {
		if selector.IsOpen() {
			if selector.HandleMouse(msg) {
				return true
			}
		}
		if checkboxStyleGroup.HandleMouse(msg) {
			return true
		}
		if radioGroup.HandleMouse(msg) {
			return true
		}
		if slider.HandleMouse(msg) {
			return true
		}
		if selector.HandleMouse(msg) {
			return true
		}
		if btn1.HandleMouse(msg) {
			return true
		}
		if btn2.HandleMouse(msg) {
			return true
		}
		return false
	})

	// Hotkeys
	app.OnKeyRune('f', func() {
		app.Confirm("EMERGENCY FLATTEN", "Liquidate all positions at market price?", func() {
			app.ToastError("FLATTENED", "All positions liquidated")
		}, nil)
	})

	app.OnKeyRune('a', func() {
		app.Alert("SYSTEM NOTICE", "Routine cluster health audit completed successfully.", nil)
	})

	app.OnKeyRune('p', func() {
		app.Prompt("PARAMETER OVERRIDE", "Enter new risk ceiling (e.g. 2.5):", func(val string) {
			app.ToastInfo("PARAMETER UPDATED", fmt.Sprintf("New ceiling set to: %s", val))
		}, nil)
	})

	app.OnKeyRune('t', func() {
		app.ToastSuccess("TELEMETRY", "Snapshot exported to disk")
	})

	app.OnKeyRune('q', func() {
		app.Quit()
	})

	app.OnKey(input.KeyEsc, func() {
		app.CloseModal()
	})

	// Initial toast
	app.ToastInfo("READY", "Dashboard loaded. Press [?] or key hints below.")

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "App error: %v\n", err)
	}
}
