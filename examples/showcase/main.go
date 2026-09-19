package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/baibeicha/goatui"
	"github.com/baibeicha/goatui/pkg/widgets"
)

type tickMsg time.Time

type appModel struct {
	ticks       int
	counter     int
	selectedRow int
	lastClicked string
	table       *widgets.VirtualTable
	sparkline   *widgets.Sparkline
	gauge       *widgets.Gauge
	braille     *widgets.BrailleCanvas
	data        []float64
}

func initialModel() appModel {
	data := make([]float64, 120)
	for i := range data {
		data[i] = 50 + 40*math.Sin(float64(i)*0.1)
	}

	cols := []widgets.TableColumn{
		{Title: "ID", Width: 9, Align: widgets.AlignRight, HeaderAlign: widgets.AlignCenter},
		{Title: "Process Name", Flex: 1, MinWidth: 20, Align: widgets.AlignLeft},
		{Title: "Memory", Width: 12, Align: widgets.AlignRight},
		{Title: "Status", Width: 14, Align: widgets.AlignCenter},
	}

	table := widgets.NewTable(cols).
		SetTotalRows(100_000).
		SetBorder(widgets.TableBorderClean).
		SetBorderStyle(goatui.NewStyle().Foreground(goatui.ColorHex("#4A4B68"))).
		SetHeaderStyle(goatui.NewStyle().Bold(true).Foreground(goatui.ColorHex("#00D2FF"))).
		SetHeaderBorderStyle(goatui.NewStyle().Foreground(goatui.ColorHex("#7D56F4"))).
		SetRowStyle(goatui.NewStyle().Foreground(goatui.ColorHex("#D0D0E0"))).
		SetAlternateRowStyle(goatui.NewStyle().Foreground(goatui.ColorHex("#D0D0E0")).Background(goatui.Color256(234))).
		SetSelectedStyle(goatui.NewStyle().Bold(true).Foreground(goatui.ColorHex("#00FFAA")).Background(goatui.Color256(236))).
		SetZebra(true).
		SetSelectionPrefix("> ").
		SetPadding(1)

	// High-performance on-demand row data provider
	table.SetRowProvider(func(row int) []string {
		status := "RUNNING"
		if row%5 == 0 {
			status = "IDLE"
		} else if row%13 == 0 {
			status = "SLEEP"
		}
		return []string{
			fmt.Sprintf("#%06d", row),
			"worker-daemon-" + strconv.Itoa(row%16),
			strconv.Itoa(128+(row*17)%8192) + " MB",
			status,
		}
	})

	// Custom cell renderer to draw vibrant colored status badges!
	table.SetCellRenderer(func(buf *goatui.Buffer, col widgets.TableColumn, cellRect goatui.Rect, rowIndex int, colIndex int, selected bool) bool {
		if colIndex == 3 { // Status column
			status := "● RUNNING"
			badgeFg := goatui.ColorHex("#00FF88")
			if rowIndex%5 == 0 {
				status = "○ IDLE"
				badgeFg = goatui.ColorHex("#00D2FF")
			} else if rowIndex%13 == 0 {
				status = "◌ SLEEP"
				badgeFg = goatui.ColorHex("#FFB86C")
			}
			bg := goatui.DefaultColor()
			if selected {
				bg = goatui.Color256(236)
			} else if rowIndex%2 == 1 {
				bg = goatui.Color256(234)
			}
			widgets.RenderCellText(buf, cellRect, status, widgets.AlignCenter, 1, badgeFg, bg, 1<<0)
			return true
		}
		return false
	})

	return appModel{
		table:     table,
		sparkline: widgets.NewSparkline(data),
		gauge:     widgets.NewGauge().SetAlign(goatui.AlignCenter),
		braille:   widgets.NewBrailleCanvas(40, 10),
		data:      data,
	}
}

func (m appModel) Init() goatui.Cmd {
	return goatui.Tick(16*time.Millisecond, func(t time.Time) goatui.Msg {
		return tickMsg(t)
	})
}

func (m appModel) Update(msg goatui.Msg) (goatui.Model, goatui.Cmd) {
	switch msg := msg.(type) {
	case goatui.KeyMsg:
		switch msg.Key.Type {
		case goatui.KeyRune:
			switch msg.Key.Rune {
			case 'q':
				return m, goatui.Quit
			case 'j':
				m.table.ScrollDown(1)
			case 'k':
				m.table.ScrollUp(1)
			case '+':
				m.counter++
			case '-':
				m.counter--
			}
		case goatui.KeyUp:
			m.table.ScrollUp(1)
		case goatui.KeyDown:
			m.table.ScrollDown(1)
		}

	case goatui.HitMsg:
		if !msg.IsLeftClick() {
			break
		}
		m.lastClicked = msg.ID()
		switch msg.ID() {
		case "btn-inc":
			m.counter += 10
		case "btn-dec":
			m.counter -= 10
		case "btn-quit":
			return m, goatui.Quit
		case "table-body":
			if area, ok := msg.Target.UserData.(goatui.Rect); ok {
				if row := m.table.RowAt(area, msg.Mouse.Y); row >= 0 {
					m.table.Select(row)
					m.lastClicked = fmt.Sprintf("row #%d", row)
				}
			}
		}

	case tickMsg:
		m.ticks++
		// Update telemetry
		m.data = append(m.data[1:], 50+40*math.Sin(float64(m.ticks)*0.1))
		m.sparkline.SetData(m.data)

		pct := (math.Sin(float64(m.ticks)*0.05) + 1.0) / 2.0
		m.gauge.SetPercent(pct)

		return m, goatui.Tick(16*time.Millisecond, func(t time.Time) goatui.Msg {
			return tickMsg(t)
		})
	}

	return m, nil
}

func (m appModel) View(f *goatui.Frame) {
	area := f.Area()
	if area.IsEmpty() {
		return
	}

	// 1. Top-level Vertical Split: Header (3), Main Body (Flex), Footer (1)
	rows := goatui.SplitVertical(area,
		goatui.Fixed(3),
		goatui.Flex(1),
		goatui.Fixed(1),
	)

	// Header with Rounded Border and Centered Title
	headerStyle := goatui.NewStyle().
		Bold(true).
		Foreground(goatui.ColorHex("#00D2FF")).
		Border(goatui.BorderRounded).
		BorderForeground(goatui.ColorHex("#7D56F4")).
		AlignCenter().
		AlignMiddle()
	headerStyle.Draw(f.Buffer, rows[0], "[*] GoatUI: Next-Gen High-Performance TUI Framework (60 FPS / 0 Allocs)")

	// 2. Main Body Split: Left Column (Sidebar, 35%), Right Column (Table, Flex)
	bodyCols := goatui.SplitHorizontal(rows[1],
		goatui.Percent(35),
		goatui.Flex(1),
	)

	// Left Column: Responsive 3-card stack
	// Card 1: Interactive Counter/Buttons (Fixed 6 rows)
	// Card 2: Braille Wave Graphic (Flex 1, expands to fill available height)
	// Card 3: Telemetry & Progress Gauge (Fixed 7 rows)
	leftRows := goatui.SplitVertical(bodyCols[0],
		goatui.Fixed(6),
		goatui.Flex(1),
		goatui.Fixed(7),
	)

	// --- Card 1: Interactive Controls ---
	cardStyle := goatui.NewStyle().
		Border(goatui.BorderNormal).
		BorderForeground(goatui.ColorHex("#FF007F"))
	cardStyle.Draw(f.Buffer, leftRows[0], "")

	cardInner := leftRows[0].Inset(1, 1)
	if !cardInner.IsEmpty() {
		f.Buffer.SetString(cardInner.X+1, cardInner.Y, fmt.Sprintf("Counter: %d", m.counter), goatui.ColorHex("#FFFF00"), goatui.DefaultColor(), 1<<0)
		f.Buffer.SetString(cardInner.X+1, cardInner.Y+1, fmt.Sprintf("Last Hit: %s", m.lastClicked), goatui.ColorHex("#AAAAAA"), goatui.DefaultColor(), 0)

		// Interactive Buttons registered in SpatialMap for mouse clicks
		btnIncArea := goatui.NewRect(cardInner.X+1, cardInner.Y+3, 10, 1)
		f.Buffer.SetString(btnIncArea.X, btnIncArea.Y, "[ +10 ]", goatui.ColorHex("#00FF88"), goatui.Color256(235), 1<<0)
		f.RegisterHit("btn-inc", btnIncArea, 1, nil)

		btnDecArea := goatui.NewRect(cardInner.X+13, cardInner.Y+3, 10, 1)
		f.Buffer.SetString(btnDecArea.X, btnDecArea.Y, "[ -10 ]", goatui.ColorHex("#FF5555"), goatui.Color256(235), 1<<0)
		f.RegisterHit("btn-dec", btnDecArea, 1, nil)
	}

	// --- Card 2: Braille Sine Wave ---
	brailleBox := goatui.NewStyle().
		Border(goatui.BorderRounded).
		BorderForeground(goatui.ColorHex("#00FF88"))
	brailleBox.Draw(f.Buffer, leftRows[1], " Sine Wave Oscillator ")
	brailleInner := leftRows[1].Inset(1, 1)
	if !brailleInner.IsEmpty() {
		m.braille.Resize(brailleInner.Width, brailleInner.Height)
		m.braille.Clear() // Clear previous frame's points to prevent solid dot accumulation!
		w := m.braille.SubWidth()
		h := m.braille.SubHeight()
		midY := h / 2
		amp := float64(midY - 2)
		if amp < 1 {
			amp = 1
		}

		var prevX, prevY int
		for x := 0; x < w; x++ {
			y := midY + int(amp*math.Sin(float64(x+m.ticks*2)*0.08))
			if x == 0 {
				prevX, prevY = x, y
			}
			m.braille.DrawLine(prevX, prevY, x, y)
			prevX, prevY = x, y
		}
		m.braille.Draw(f.Buffer, brailleInner)
	}

	// --- Card 3: Telemetry Sparkline & Resource Gauge ---
	telemBox := goatui.NewStyle().
		Border(goatui.BorderNormal).
		BorderForeground(goatui.ColorHex("#FF9900"))
	telemBox.Draw(f.Buffer, leftRows[2], " Telemetry ")

	telemInner := leftRows[2].Inset(1, 1)
	if telemInner.Height >= 4 {
		telemSections := goatui.SplitVertical(telemInner,
			goatui.Fixed(1), // Sparkline title
			goatui.Fixed(1), // Sparkline graph
			goatui.Fixed(1), // Gauge title
			goatui.Fixed(1), // Gauge bar
		)
		f.Buffer.SetString(telemSections[0].X, telemSections[0].Y, "Telemetry Trend:", goatui.Color256(246), goatui.DefaultColor(), 0)
		m.sparkline.Draw(f.Buffer, telemSections[1])

		f.Buffer.SetString(telemSections[2].X, telemSections[2].Y, "Resource Gauge:", goatui.Color256(246), goatui.DefaultColor(), 0)
		m.gauge.Draw(f.Buffer, telemSections[3])
	}

	// --- Right Column: Virtualized Table (100,000 Rows) ---
	tableBox := goatui.NewStyle().
		Border(goatui.BorderRounded).
		BorderForeground(goatui.ColorHex("#7D56F4"))
	tableBox.Draw(f.Buffer, bodyCols[1], " Process Registry (100,000 Tasks) ")
	tableInner := bodyCols[1].Inset(1, 1)
	if !tableInner.IsEmpty() {
		m.table.Draw(f.Buffer, tableInner)
		f.RegisterHit("table-body", tableInner, 1, tableInner)
	}

	// 3. Footer Bar with Quit Button
	footerArea := goatui.NewRect(rows[2].X, rows[2].Y, rows[2].Width-14, 1)
	footerText := " Navigate: [J/K] or [Arrows] | Mouse: Click buttons/rows | Quit: [Q]"
	f.Buffer.SetString(footerArea.X, footerArea.Y, footerText, goatui.Color256(245), goatui.Color256(234), 0)

	quitBtnArea := goatui.NewRect(rows[2].Right()-12, rows[2].Y, 10, 1)
	f.Buffer.SetString(quitBtnArea.X, quitBtnArea.Y, "[ Exit ]", goatui.ColorHex("#FF0055"), goatui.Color256(235), 1<<0)
	f.RegisterHit("btn-quit", quitBtnArea, 5, nil)
}

func main() {
	p := goatui.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
