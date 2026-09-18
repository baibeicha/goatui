package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/baibeicha/goatui"
	"github.com/baibeicha/goatui/pkg/tea"
	"github.com/baibeicha/goatui/pkg/theme"
	"github.com/baibeicha/goatui/pkg/widgets"
)

type tickMsg time.Time

var (
	globalPulsePhase float64
	globalIconMode   = goatui.IconModeASCII
	globalTabs       = goatui.NewTabs(
		goatui.TabItem{ID: "/dashboard", Title: "Dashboard", Hotkey: '1'},
		goatui.TabItem{ID: "/processes", Title: "Processes", Hotkey: '2'},
		goatui.TabItem{ID: "/admin", Title: "Admin", Hotkey: '3'},
		goatui.TabItem{ID: "/settings", Title: "Settings", Hotkey: '4'},
		goatui.TabItem{ID: "/files", Title: "Files", Hotkey: '5'},
		goatui.TabItem{ID: "/media", Title: "Media", Hotkey: '6'},
	)
)

func toggleAdminUser() {
	sm := goatui.DefaultSecurity()
	u := sm.CurrentUser()
	if u != nil && u.HasRole("admin") {
		sm.Login(&goatui.User{
			ID:          "usr_guest",
			Username:    "guest_user",
			Roles:       []string{"guest"},
			Permissions: []string{"read:public"},
		}, "guest-token")
	} else {
		sm.Login(&goatui.User{
			ID:          "usr_admin",
			Username:    "alice",
			Roles:       []string{"admin", "user"},
			Permissions: []string{"*"},
		}, "admin-token")
	}
}

// -----------------------------------------------------------------------------
// Base Application Screen Wrapper (Common Header & Navigation Bar)
// -----------------------------------------------------------------------------

type appScreen struct {
	goatui.BaseScreen
	activeRoute string
}

func (a *appScreen) OnMount(ctx *goatui.RouteContext) {
	if ctx != nil {
		a.activeRoute = ctx.Path
	}
}

func (a *appScreen) renderChrome(f *tea.Frame, title string) goatui.Rect {
	area := f.Area()
	if area.IsEmpty() {
		return area
	}

	curTheme := goatui.DefaultTheme().Current()
	p := curTheme.Colors

	// Guard against tiny terminal window
	if area.Width < 50 || area.Height < 8 {
		f.Buffer.Fill(area, goatui.Cell{
			Rune:   ' ',
			Width:  1,
			FgType: goatui.DefaultColor().Type,
			BgType: p.Background.Type,
			Bg:     p.Background.Value,
		})
		msg := "[Terminal Too Small: Please resize window]"
		f.Buffer.SetString(area.X+max(0, (area.Width-len(msg))/2), area.Y+area.Height/2, msg, p.Warning, p.Background, goatui.AttrBold)
		return goatui.NewRect(0, 0, 0, 0)
	}

	// 1. Header Bar
	headerArea := goatui.NewRect(area.X, area.Y, area.Width, 1)
	headerStyle := goatui.NewStyle().
		Bold(true).
		Foreground(p.Background).
		Background(p.Primary)
	headerTitle := " [GoatUI Enterprise Desktop] "
	if area.Width < 60 {
		headerTitle = " [GoatUI] "
	}
	headerStyle.Draw(f.Buffer, headerArea, headerTitle)

	// Status info on right of header with animated pulsing glow
	user := goatui.DefaultSecurity().CurrentUser()
	authStatus := " [Guest (Press T to Elevate)] "
	if area.Width < 80 {
		authStatus = " [Guest] "
	}
	authBg := goatui.PulseColor(p.Warning, p.Danger, globalPulsePhase)
	if user != nil && user.HasRole("admin") {
		authStatus = " [Admin: " + user.Username + " (Press T to Drop)] "
		if area.Width < 80 {
			authStatus = " [Admin] "
		}
		authBg = goatui.PulseColor(p.Success, p.Accent, globalPulsePhase)
	}
	authW := goatui.StringWidth(authStatus)
	authArea := goatui.NewRect(headerArea.Right()-authW-1, area.Y, authW, 1)
	f.Buffer.SetString(authArea.X, authArea.Y, authStatus, p.Background, authBg, goatui.AttrBold)

	if area.Width >= 95 {
		themeInfo := " Theme: " + curTheme.Name + " "
		f.Buffer.SetString(authArea.X-goatui.StringWidth(themeInfo)-1, area.Y, themeInfo, p.Foreground, p.Secondary, 0)
	}

	// 2. Navigation Bar (Top subheader) using interactive Tabs widget
	navArea := goatui.NewRect(area.X, area.Y+1, area.Width, 1)
	switch a.activeRoute {
	case "/dashboard", "/":
		globalTabs.SetActive(0)
	case "/processes":
		globalTabs.SetActive(1)
	case "/admin":
		globalTabs.SetActive(2)
	case "/settings":
		globalTabs.SetActive(3)
	case "/files":
		globalTabs.SetActive(4)
	case "/media":
		globalTabs.SetActive(5)
	}
	globalTabs.SetActiveStyle(goatui.NewStyle().
		Bold(true).
		Foreground(p.Background).
		Background(p.Accent))
	globalTabs.SetInactiveStyle(goatui.NewStyle().
		Foreground(p.Foreground).
		Background(p.Secondary))
	globalTabs.Draw(f.Buffer, navArea)

	if area.Width >= 88 {
		routeInfo := fmt.Sprintf(" [Ctrl+K] Omnibar | Route: %s ", a.activeRoute)
		f.Buffer.SetString(navArea.Right()-goatui.StringWidth(routeInfo)-1, navArea.Y, routeInfo, p.Muted, p.Background, 0)
	}

	// 3. Footer Bar
	footerArea := goatui.NewRect(area.X, area.Bottom()-1, area.Width, 1)
	footerText := " [1..6] Switch Tab | [T] Toggle Admin/Guest | [Ctrl+K] Omnibar | [Esc] Back | [Q] Exit"
	if area.Width < 85 {
		footerText = " [1..6] Tabs | [T] Role | [Ctrl+K] Omni | [Esc] Back | [Q] Exit"
	}
	f.Buffer.SetString(footerArea.X, footerArea.Y, footerText, p.Muted, p.Background, 0)

	// Return content area between header and footer
	contentArea := goatui.NewRect(area.X, area.Y+2, area.Width, max(0, area.Height-3))
	return contentArea
}

func handleCommonKeys(msg tea.Msg) (bool, tea.Cmd) {
	switch msg := msg.(type) {
	case goatui.ToggleRoleMsg:
		toggleAdminUser()
		return true, nil
	case tea.MouseMsg:
		if msg.Action == goatui.MousePress && msg.Button == goatui.MouseLeft {
			if globalTabs.HandleMouse(msg) {
				if item := globalTabs.ActiveItem(); item != nil {
					return true, goatui.NavigateReplace(item.ID)
				}
			}
		}
	case tea.KeyMsg:
		switch msg.Key.Rune {
		case '1':
			globalTabs.SetActive(0)
			return true, goatui.NavigateReplace("/dashboard")
		case '2':
			globalTabs.SetActive(1)
			return true, goatui.NavigateReplace("/processes")
		case '3':
			globalTabs.SetActive(2)
			return true, goatui.NavigateReplace("/admin")
		case '4':
			globalTabs.SetActive(3)
			return true, goatui.NavigateReplace("/settings")
		case '5':
			globalTabs.SetActive(4)
			return true, goatui.NavigateReplace("/files")
		case '6':
			globalTabs.SetActive(5)
			return true, goatui.NavigateReplace("/media")
		case 't', 'T':
			toggleAdminUser()
			return true, func() tea.Msg { return goatui.ToggleRoleMsg{} }
		case 'q', 'Q':
			return true, goatui.Quit
		}
		if msg.Key.Type == goatui.KeyEsc {
			return true, func() tea.Msg { return goatui.PopMsg{} }
		}
	}
	return false, nil
}

// -----------------------------------------------------------------------------
// 1. Dashboard Screen
// -----------------------------------------------------------------------------

type DashboardScreen struct {
	appScreen
	ticks      int
	sparkline  *widgets.Sparkline
	gauge      *widgets.Gauge
	data       []float64
	smoothLoad *goatui.SmoothFloat
}

func newDashboardScreen() *DashboardScreen {
	data := make([]float64, 80)
	for i := range data {
		data[i] = 50 + 35*math.Sin(float64(i)*0.15)
	}

	sp := widgets.NewSparkline(data)
	g := widgets.NewGauge()
	g.SetPercent(0.68)
	g.SetLabel("CPU Cluster Load: 68%")

	return &DashboardScreen{
		sparkline:  sp,
		gauge:      g,
		data:       data,
		smoothLoad: goatui.NewSmoothFloat(0.68, 10.0),
	}
}

func (s *DashboardScreen) Init() tea.Cmd {
	return goatui.Tick(40*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (s *DashboardScreen) Update(msg tea.Msg) (goatui.Screen, tea.Cmd) {
	if handled, cmd := handleCommonKeys(msg); handled {
		return s, cmd
	}

	switch msg.(type) {
	case tickMsg:
		s.ticks++
		globalPulsePhase += 0.08
		v := 50 + 38*math.Sin(float64(s.ticks)*0.12) + 8*math.Cos(float64(s.ticks)*0.3)
		if v < 0 {
			v = 0
		} else if v > 100 {
			v = 100
		}
		s.data = append(s.data[1:], v)
		s.sparkline.SetData(s.data)

		s.smoothLoad.SetTarget(v / 100.0)
		smoothed := s.smoothLoad.Update(0.04)
		s.gauge.SetPercent(smoothed)
		s.gauge.SetLabel(fmt.Sprintf("%d%%", int(smoothed*100)))

		return s, goatui.Tick(40*time.Millisecond, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	}
	return s, nil
}

func (s *DashboardScreen) View(f *tea.Frame) {
	contentArea := s.renderChrome(f, "Dashboard")
	if contentArea.IsEmpty() {
		return
	}

	// Buffer clip guarantees no element leaks outside contentArea
	f.Buffer.SetClip(contentArea)
	defer f.Buffer.ResetClip()

	curTheme := goatui.DefaultTheme().Current()
	p := curTheme.Colors

	// Adaptive column splitting: stacked if width < 90, side-by-side if width >= 90
	var leftArea, rightArea goatui.Rect
	if contentArea.Width < 90 {
		rows := goatui.SplitVertical(contentArea,
			goatui.Percent(55),
			goatui.Percent(45),
		)
		leftArea = rows[0]
		rightArea = rows[1]
	} else {
		cols := goatui.SplitHorizontal(contentArea,
			goatui.Percent(50),
			goatui.Percent(50),
		)
		leftArea = cols[0]
		rightArea = cols[1]
	}

	// Left: Telemetry Panel
	leftBox := goatui.NewStyle().
		Border(curTheme.Borders.Border).
		BorderForeground(p.Primary).
		Title("Live Cluster Telemetry")
	leftBox.Draw(f.Buffer, leftArea, "")

	innerLeft := leftArea.Inset(1, 1)
	if !innerLeft.IsEmpty() {
		f.Buffer.SetClip(innerLeft)
		sparkH := min(8, max(2, innerLeft.Height/3))
		rows := goatui.SplitVertical(innerLeft,
			goatui.Fixed(1),      // Sparkline title
			goatui.Fixed(sparkH), // Sparkline
			goatui.Fixed(1),      // Spacer
			goatui.Fixed(1),      // Gauge title
			goatui.Fixed(1),      // Gauge
			goatui.Flex(1),       // Metrics summary
		)

		f.Buffer.SetString(rows[0].X, rows[0].Y, "Throughput Sparkline (Pure UTF-8 Cells):", p.Accent, goatui.DefaultColor(), goatui.AttrBold)
		s.sparkline.SetFg(p.Secondary)
		s.sparkline.Draw(f.Buffer, rows[1])

		f.Buffer.SetString(rows[3].X, rows[3].Y, "Aggregate Resource Gauge:", p.Accent, goatui.DefaultColor(), goatui.AttrBold)
		s.gauge.SetFg(p.Success).SetBg(p.Secondary)
		s.gauge.Draw(f.Buffer, rows[4])

		// Stats
		if len(rows) > 5 && rows[5].Height > 1 {
			statY := rows[5].Y + 1
			f.Buffer.SetString(rows[5].X, statY, "• Active Goroutines: 48", p.Foreground, goatui.DefaultColor(), 0)
			if rows[5].Height > 2 {
				f.Buffer.SetString(rows[5].X, statY+1, "• Terminal Buffer: Zero-Allocation Radix Diff", p.Foreground, goatui.DefaultColor(), 0)
			}
			if rows[5].Height > 3 {
				f.Buffer.SetString(rows[5].X, statY+2, "• Router Status: Trie-based dynamic URL router", p.Foreground, goatui.DefaultColor(), 0)
			}
			if rows[5].Height > 4 {
				f.Buffer.SetString(rows[5].X, statY+3, "• Omnibar: Press [Ctrl+K] to launch anywhere", p.Primary, goatui.DefaultColor(), goatui.AttrBold)
			}
		}
		f.Buffer.SetClip(contentArea)
	}

	// Right: Quick Navigation & System Health
	rightBox := goatui.NewStyle().
		Border(curTheme.Borders.Border).
		BorderForeground(p.Secondary).
		Title("System Operations & Quick Launch")
	rightBox.Draw(f.Buffer, rightArea, "")

	innerRight := rightArea.Inset(1, 1)
	if !innerRight.IsEmpty() {
		f.Buffer.SetClip(innerRight)
		f.Buffer.SetString(innerRight.X, innerRight.Y, "Navigation Shortcuts:", p.Primary, goatui.DefaultColor(), goatui.AttrBold)
		if innerRight.Height >= 3 {
			f.Buffer.SetString(innerRight.X, innerRight.Y+2, " [2] Process Registry  -> 10,000 tasks table", p.Foreground, goatui.DefaultColor(), 0)
		}
		if innerRight.Height >= 5 {
			f.Buffer.SetString(innerRight.X, innerRight.Y+4, " [3] Security & RBAC   -> Route-guarded admin panel", p.Foreground, goatui.DefaultColor(), 0)
		}
		if innerRight.Height >= 7 {
			f.Buffer.SetString(innerRight.X, innerRight.Y+6, " [4] Theme Customizer  -> Instant YAML themes", p.Foreground, goatui.DefaultColor(), 0)
		}
		if innerRight.Height >= 9 {
			f.Buffer.SetString(innerRight.X, innerRight.Y+8, " [Ctrl+K] Omnibar      -> Command palette & paths", p.Accent, goatui.DefaultColor(), goatui.AttrBold)
		}

		// Only render diagBox if we have enough vertical space (at least 15 rows)
		if innerRight.Height >= 15 {
			diagBoxH := min(5, innerRight.Height-10)
			diagBox := goatui.NewRect(innerRight.X, innerRight.Y+10, innerRight.Width, diagBoxH)
			goatui.NewStyle().Border(goatui.BorderRounded).BorderForeground(p.Muted).Title("Engine Status").Draw(f.Buffer, diagBox, "")
			if diagBoxH >= 3 {
				f.Buffer.SetString(diagBox.X+2, diagBox.Y+1, "Cell Buffer: Fast Bitmask Runes", p.Success, goatui.DefaultColor(), 0)
			}
			if diagBoxH >= 4 {
				f.Buffer.SetString(diagBox.X+2, diagBox.Y+2, "Theme Engine: Dynamic Palette Swapping", p.Info, goatui.DefaultColor(), 0)
			}
			if diagBoxH >= 5 {
				f.Buffer.SetString(diagBox.X+2, diagBox.Y+3, "Security: Enterprise Route Guards Active", p.Primary, goatui.DefaultColor(), 0)
			}
		}
		f.Buffer.SetClip(contentArea)
	}
}

// -----------------------------------------------------------------------------
// 2. Processes Screen (VirtualTable)
// -----------------------------------------------------------------------------

type ProcessesScreen struct {
	appScreen
	table *widgets.VirtualTable
}

func newProcessesScreen() *ProcessesScreen {
	cols := []widgets.TableColumn{
		{Title: "PID", Width: 8, Align: widgets.AlignRight, HeaderAlign: widgets.AlignCenter},
		{Title: "Command Name", Flex: 1, MinWidth: 16, Align: widgets.AlignLeft},
		{Title: "CPU %", Width: 10, Align: widgets.AlignRight},
		{Title: "Memory", Width: 12, Align: widgets.AlignRight},
		{Title: "User", Width: 10, Align: widgets.AlignCenter},
		{Title: "Status", Width: 12, Align: widgets.AlignCenter},
	}

	tbl := widgets.NewTable(cols).
		SetTotalRows(10_000).
		SetBorder(widgets.TableBorderClean).
		SetSelectionPrefix("► ").
		SetZebra(true).
		SetPadding(1)

	// Generate 10,000 realistic processes
	tbl.SetRowProvider(func(row int) []string {
		pid := strconv.Itoa(1000 + row)
		cmds := []string{"systemd", "goatui-daemon", "sshd", "postgres", "redis-server", "nginx", "docker", "kubelet", "top", "bash"}
		cmd := cmds[row%len(cmds)]
		cpu := fmt.Sprintf("%.1f%%", float64((row*17)%1000)/10.0)
		mem := fmt.Sprintf("%d MB", 16+((row*43)%4096))
		user := "root"
		if row%3 == 0 {
			user = "goat"
		} else if row%5 == 0 {
			user = "postgres"
		}
		status := "RUNNING"
		if row%7 == 0 {
			status = "SLEEP"
		} else if row%19 == 0 {
			status = "IDLE"
		}
		return []string{pid, cmd, cpu, mem, user, status}
	})

	return &ProcessesScreen{table: tbl}
}

func (p *ProcessesScreen) Update(msg tea.Msg) (goatui.Screen, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.Key.Type {
		case goatui.KeyUp:
			p.table.SelectPrev()
			return p, nil
		case goatui.KeyDown:
			p.table.SelectNext()
			return p, nil
		case goatui.KeyEnter:
			selectedRow := p.table.SelectedRow()
			targetPID := strconv.Itoa(1000 + selectedRow)
			return p, goatui.Navigate("/processes/" + targetPID)
		}
	}

	if handled, cmd := handleCommonKeys(msg); handled {
		return p, cmd
	}
	return p, nil
}

func (p *ProcessesScreen) View(f *tea.Frame) {
	contentArea := p.renderChrome(f, "Processes")
	if contentArea.IsEmpty() {
		return
	}

	curTheme := goatui.DefaultTheme().Current()
	colors := curTheme.Colors

	// Apply current theme to table
	p.table.SetBorderStyle(goatui.NewStyle().Foreground(colors.Muted)).
		SetHeaderStyle(goatui.NewStyle().Bold(true).Foreground(colors.Primary)).
		SetHeaderBorderStyle(goatui.NewStyle().Foreground(colors.Secondary)).
		SetRowStyle(goatui.NewStyle().Foreground(colors.Foreground)).
		SetAlternateRowStyle(goatui.NewStyle().Foreground(colors.Foreground).Background(goatui.Color256(234))).
		SetSelectedStyle(goatui.NewStyle().Bold(true).Foreground(colors.Background).Background(colors.Accent))

	box := goatui.NewStyle().
		Border(curTheme.Borders.Border).
		BorderForeground(colors.Primary).
		Title("Process Registry (10,000 Tasks - Enter=Details)")
	box.Draw(f.Buffer, contentArea, "")

	tableInner := contentArea.Inset(1, 1)
	if !tableInner.IsEmpty() {
		p.table.Draw(f.Buffer, tableInner)
	}
}

// -----------------------------------------------------------------------------
// 3. Process Detail Screen (Dynamic Route /processes/:id)
// -----------------------------------------------------------------------------

type ProcessDetailScreen struct {
	appScreen
	pid string
}

func newProcessDetailScreen(pid string) *ProcessDetailScreen {
	return &ProcessDetailScreen{pid: pid}
}

func (d *ProcessDetailScreen) OnMount(ctx *goatui.RouteContext) {
	d.appScreen.OnMount(ctx)
	if ctx != nil && ctx.HasParam("id") {
		d.pid = ctx.Param("id")
	}
}

func (d *ProcessDetailScreen) Update(msg tea.Msg) (goatui.Screen, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.Key.Rune {
		case 'k', 'K', 'x', 'X':
			// Open Confirm Termination Modal!
			confirmModal := goatui.NewConfirmModal(
				"Confirm Termination",
				fmt.Sprintf("Are you sure you want to terminate process PID #%s?", d.pid),
				func() tea.Cmd {
					// On Confirm: show Alert Modal and Pop back to processes
					return func() tea.Msg {
						return goatui.ShowModalMsg{
							Modal: goatui.NewAlertModal("Process Terminated", fmt.Sprintf("Process #%s successfully killed.", d.pid), func() tea.Cmd {
								return func() tea.Msg { return goatui.PopMsg{} }
							}),
						}
					}
				},
				func() tea.Cmd {
					return func() tea.Msg { return goatui.CloseModalMsg{} }
				},
			)
			return d, func() tea.Msg {
				return goatui.ShowModalMsg{Modal: confirmModal}
			}
		}
	}

	if handled, cmd := handleCommonKeys(msg); handled {
		return d, cmd
	}
	return d, nil
}

func (d *ProcessDetailScreen) View(f *tea.Frame) {
	contentArea := d.renderChrome(f, "Process Details")
	if contentArea.IsEmpty() {
		return
	}

	curTheme := goatui.DefaultTheme().Current()
	p := curTheme.Colors

	box := goatui.NewStyle().
		Border(curTheme.Borders.Border).
		BorderForeground(p.Accent).
		Title(fmt.Sprintf("Process Inspector: PID #%s", d.pid))
	box.Draw(f.Buffer, contentArea, "")

	inner := contentArea.Inset(2, 1)
	if inner.IsEmpty() {
		return
	}

	f.Buffer.SetString(inner.X, inner.Y, fmt.Sprintf("Process Identification: PID %s", d.pid), p.Primary, goatui.DefaultColor(), goatui.AttrBold)
	f.Buffer.SetString(inner.X, inner.Y+1, "Command Line:           /usr/bin/goatui-worker --daemon --threads=8", p.Foreground, goatui.DefaultColor(), 0)
	f.Buffer.SetString(inner.X, inner.Y+2, "Execution Path:         /proc/"+d.pid+"/exe", p.Foreground, goatui.DefaultColor(), 0)
	f.Buffer.SetString(inner.X, inner.Y+3, "Working Directory:      /var/lib/goatui", p.Foreground, goatui.DefaultColor(), 0)
	f.Buffer.SetString(inner.X, inner.Y+4, "Parent Process:         PID 1 (systemd)", p.Foreground, goatui.DefaultColor(), 0)
	f.Buffer.SetString(inner.X, inner.Y+5, "Assigned User:          root (UID: 0, GID: 0)", p.Foreground, goatui.DefaultColor(), 0)

	f.Buffer.SetString(inner.X, inner.Y+7, "Performance Metrics:", p.Primary, goatui.DefaultColor(), goatui.AttrBold)
	f.Buffer.SetString(inner.X, inner.Y+8, "CPU Utilization:        14.2% (Kernel: 2.1%, User: 12.1%)", p.Success, goatui.DefaultColor(), 0)
	f.Buffer.SetString(inner.X, inner.Y+9, "Memory Resident (RSS):  128.4 MB", p.Info, goatui.DefaultColor(), 0)
	f.Buffer.SetString(inner.X, inner.Y+10, "Open File Descriptors:  34 / 1024", p.Foreground, goatui.DefaultColor(), 0)

	btnArea := goatui.NewRect(inner.X, inner.Y+13, 24, 1)
	goatui.NewStyle().
		Bold(true).
		Foreground(goatui.ColorHex("#FFFFFF")).
		Background(p.Danger).
		Draw(f.Buffer, btnArea, " [K] Terminate Process ")

	f.Buffer.SetString(inner.X+26, inner.Y+13, "Press [Esc] to go back to process list", p.Muted, goatui.DefaultColor(), 0)
}

// -----------------------------------------------------------------------------
// 4. Admin Screen (RBAC Protected)
// -----------------------------------------------------------------------------

type AdminScreen struct {
	appScreen
}

func newAdminScreen() *AdminScreen {
	return &AdminScreen{}
}

func (a *AdminScreen) Update(msg tea.Msg) (goatui.Screen, tea.Cmd) {
	if handled, cmd := handleCommonKeys(msg); handled {
		return a, cmd
	}
	return a, nil
}

func (a *AdminScreen) View(f *tea.Frame) {
	contentArea := a.renderChrome(f, "Admin")
	if contentArea.IsEmpty() {
		return
	}

	curTheme := goatui.DefaultTheme().Current()
	p := curTheme.Colors
	user := goatui.DefaultSecurity().CurrentUser()

	box := goatui.NewStyle().
		Border(curTheme.Borders.Border).
		BorderForeground(p.Success).
		Title("Enterprise Security & RBAC Control Panel (Role: Admin)")
	box.Draw(f.Buffer, contentArea, "")

	inner := contentArea.Inset(2, 1)
	if inner.IsEmpty() {
		return
	}

	f.Buffer.SetString(inner.X, inner.Y, "Authenticated Security Context:", p.Primary, goatui.DefaultColor(), goatui.AttrBold)
	f.Buffer.SetString(inner.X, inner.Y+1, fmt.Sprintf("Active Identity: %s (ID: %s)", user.Username, user.ID), p.Foreground, goatui.DefaultColor(), 0)
	f.Buffer.SetString(inner.X, inner.Y+2, fmt.Sprintf("Assigned Roles:  %v", user.Roles), p.Success, goatui.DefaultColor(), goatui.AttrBold)
	f.Buffer.SetString(inner.X, inner.Y+3, fmt.Sprintf("RBAC Permission: %v", user.Permissions), p.Info, goatui.DefaultColor(), 0)
	f.Buffer.SetString(inner.X, inner.Y+4, fmt.Sprintf("Active Token:    %s", goatui.DefaultSecurity().Token()), p.Muted, goatui.DefaultColor(), 0)

	f.Buffer.SetString(inner.X, inner.Y+6, "Security Audit Trail:", p.Primary, goatui.DefaultColor(), goatui.AttrBold)
	logs := []string{
		"[12:00:01] RouteGuard: /admin allowed for subject alice [Role: admin]",
		"[12:00:05] SessionStore: Session cookie renewed with 3600s TTL",
		"[12:00:10] RBAC: Permission check 'cluster:write' verified OK",
		"[12:00:15] SecurityManager: Concurrency locks verified with -race check",
	}
	for i, l := range logs {
		f.Buffer.SetString(inner.X, inner.Y+7+i, "• "+l, p.Foreground, goatui.DefaultColor(), 0)
	}

	actionArea := goatui.NewRect(inner.X, inner.Y+12, 36, 1)
	goatui.NewStyle().
		Bold(true).
		Foreground(goatui.ColorHex("#FFFFFF")).
		Background(p.Warning).
		Draw(f.Buffer, actionArea, " [T] Drop Admin Privileges (Guest) ")

	f.Buffer.SetString(inner.X+38, inner.Y+12, "(Switching to guest will trigger 403 Access Denied)", p.Muted, goatui.DefaultColor(), 0)
}

// -----------------------------------------------------------------------------
// 5. Settings & Theme Switcher Screen
// -----------------------------------------------------------------------------

type SettingsScreen struct {
	appScreen
	selectedTheme int
	themes        []string
}

func newSettingsScreen() *SettingsScreen {
	return &SettingsScreen{
		themes: []string{
			"Dracula",
			"Catppuccin Mocha",
			"Nord",
			"Monokai",
			"Goat Dark",
			"Matrix",
			"Cyberpunk",
		},
	}
}

func (s *SettingsScreen) Update(msg tea.Msg) (goatui.Screen, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		switch key.Key.Type {
		case goatui.KeyUp:
			if s.selectedTheme > 0 {
				s.selectedTheme--
			}
			return s, nil
		case goatui.KeyDown:
			if s.selectedTheme < len(s.themes)-1 {
				s.selectedTheme++
			}
			return s, nil
		case goatui.KeyEnter:
			themeName := s.themes[s.selectedTheme]
			return s, goatui.SwitchTheme(themeName)
		}

		switch key.Key.Rune {
		case 'i', 'I':
			globalIconMode = (globalIconMode + 1) % 4
			return s, nil
		case 'm', 'M':
			// Trigger test modal
			alert := goatui.NewAlertModal("Settings Alert", "Configuration updated successfully!", func() tea.Cmd {
				return func() tea.Msg { return goatui.CloseModalMsg{} }
			})
			return s, func() tea.Msg { return goatui.ShowModalMsg{Modal: alert} }
		}
	}

	if handled, cmd := handleCommonKeys(msg); handled {
		return s, cmd
	}
	return s, nil
}

func (s *SettingsScreen) View(f *tea.Frame) {
	contentArea := s.renderChrome(f, "Settings")
	if contentArea.IsEmpty() {
		return
	}

	curTheme := goatui.DefaultTheme().Current()
	p := curTheme.Colors

	cols := goatui.SplitHorizontal(contentArea,
		goatui.Percent(50),
		goatui.Percent(50),
	)

	// Left: Theme Selector
	box := goatui.NewStyle().
		Border(curTheme.Borders.Border).
		BorderForeground(p.Primary).
		Title("Theme Palette (Use Up/Down + Enter)")
	box.Draw(f.Buffer, cols[0], "")

	innerLeft := cols[0].Inset(1, 1)
	if !innerLeft.IsEmpty() {
		f.Buffer.SetString(innerLeft.X, innerLeft.Y, "Select Active Color Theme:", p.Primary, goatui.DefaultColor(), goatui.AttrBold)
		for i, name := range s.themes {
			y := innerLeft.Y + 2 + i
			prefix := "  "
			style := goatui.NewStyle().Foreground(p.Foreground)
			if i == s.selectedTheme {
				prefix = "► "
				style = style.Bold(true).Foreground(p.Background).Background(p.Accent)
			}
			activeMark := ""
			if isCurrentTheme(curTheme.Name, name) {
				activeMark = " (Active)"
			}
			style.Draw(f.Buffer, goatui.NewRect(innerLeft.X, y, innerLeft.Width-2, 1), prefix+name+activeMark)
		}

		f.Buffer.SetString(innerLeft.X, innerLeft.Y+12, "Hot-Reload Directory: ./themes/*.yaml", p.Muted, goatui.DefaultColor(), 0)
		f.Buffer.SetString(innerLeft.X, innerLeft.Y+13, "Editing YAML files will reload styles instantly.", p.Muted, goatui.DefaultColor(), 0)
	}

	// Right: Security & Modals test panel
	secBox := goatui.NewStyle().
		Border(curTheme.Borders.Border).
		BorderForeground(p.Secondary).
		Title("Security & Preferences")
	secBox.Draw(f.Buffer, cols[1], "")

	innerRight := cols[1].Inset(1, 1)
	if !innerRight.IsEmpty() {
		user := goatui.DefaultSecurity().CurrentUser()
		roleStr := "Guest"
		if user != nil && user.HasRole("admin") {
			roleStr = "Administrator"
		}

		f.Buffer.SetString(innerRight.X, innerRight.Y, "Role Management:", p.Primary, goatui.DefaultColor(), goatui.AttrBold)
		f.Buffer.SetString(innerRight.X, innerRight.Y+1, fmt.Sprintf("Current Privileges: %s", roleStr), p.Foreground, goatui.DefaultColor(), 0)
		f.Buffer.SetString(innerRight.X, innerRight.Y+2, "Press [T] anywhere to toggle Admin role on/off", p.Accent, goatui.DefaultColor(), goatui.AttrBold)
		f.Buffer.SetString(innerRight.X, innerRight.Y+3, "When Guest, navigating to [3] Admin will show 403.", p.Muted, goatui.DefaultColor(), 0)

		// Icon set customization
		modeNames := [4]string{"ASCII", "Unicode", "NerdFont", "Emoji"}
		iconTitle := fmt.Sprintf("File Explorer Icons: [%s] (Press [I] to cycle)", modeNames[globalIconMode])
		f.Buffer.SetString(innerRight.X, innerRight.Y+5, iconTitle, p.Primary, goatui.DefaultColor(), goatui.AttrBold)
		switch globalIconMode {
		case goatui.IconModeASCII:
			f.Buffer.SetString(innerRight.X, innerRight.Y+6, "Style: [DIR] [SRC] [IMG] [VID] [CFG] [BIN] (Zero drift)", p.Success, goatui.DefaultColor(), 0)
		case goatui.IconModeUnicode:
			f.Buffer.SetString(innerRight.X, innerRight.Y+6, "Style: ■ · § ◆ ▶ ▲ (Clean mono glyphs)", p.Info, goatui.DefaultColor(), 0)
		case goatui.IconModeNerdFont:
			f.Buffer.SetString(innerRight.X, innerRight.Y+6, "Style: \uF07B \uF15C \uF03E \uF008 \uF013 (Requires patched font)", p.Accent, goatui.DefaultColor(), 0)
		case goatui.IconModeEmoji:
			f.Buffer.SetString(innerRight.X, innerRight.Y+6, "Style: 📁 📄 🖼 🎬 ⚙ (Legacy emoji mode)", p.Warning, goatui.DefaultColor(), 0)
		}

		f.Buffer.SetString(innerRight.X, innerRight.Y+8, "Modal Overlays:", p.Primary, goatui.DefaultColor(), goatui.AttrBold)
		f.Buffer.SetString(innerRight.X, innerRight.Y+10, "Press [Ctrl+K] to launch Omnibar palette", p.Success, goatui.DefaultColor(), 0)
	}
}

func isCurrentTheme(curName, targetName string) bool {
	return strings.EqualFold(curName, targetName) ||
		normalizeTheme(curName) == normalizeTheme(targetName)
}

func normalizeTheme(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "_", "")
	return s
}

// -----------------------------------------------------------------------------
// 5. Files & File Explorer Screen
// -----------------------------------------------------------------------------

type FilesScreen struct {
	appScreen
	explorer *goatui.FileExplorerScreen
}

func newFilesScreen(initialPath string) *FilesScreen {
	exp := goatui.NewFileExplorerScreen(initialPath)
	exp.SetIconMode(globalIconMode)
	return &FilesScreen{
		explorer: exp,
	}
}

func (s *FilesScreen) OnMount(ctx *goatui.RouteContext) {
	s.appScreen.OnMount(ctx)
	if s.explorer != nil {
		s.explorer.OnMount(ctx)
	}
}

func (s *FilesScreen) CurrentDir() string {
	if s.explorer != nil {
		return s.explorer.CurrentDir()
	}
	return ""
}

func (s *FilesScreen) CurrentAddress() string {
	if s.explorer != nil {
		return s.explorer.CurrentDir()
	}
	return ""
}

func (s *FilesScreen) Init() tea.Cmd {
	if s.explorer != nil {
		return s.explorer.Init()
	}
	return nil
}

func (s *FilesScreen) Update(msg tea.Msg) (goatui.Screen, tea.Cmd) {
	// When editing the address bar or filtering search, do not intercept common hotkeys (1..6, t, q)
	if s.explorer != nil && (s.explorer.IsEditingPath() || s.explorer.IsFiltering()) {
		updated, cmd := s.explorer.Update(msg)
		if exp, ok := updated.(*goatui.FileExplorerScreen); ok {
			s.explorer = exp
			globalIconMode = exp.IconMode()
		}
		return s, cmd
	}

	if handled, cmd := handleCommonKeys(msg); handled {
		return s, cmd
	}

	if kmsg, ok := msg.(tea.KeyMsg); ok {
		if (kmsg.Key.Rune == 'n' || kmsg.Key.Rune == 'N') && kmsg.Key.Mod == 0 {
			return s, func() tea.Msg {
				return goatui.ShowModalMsg{
					Modal: goatui.NewInputModal(
						"Create New Directory",
						"Folder name:",
						"new_folder",
						func(name string) tea.Cmd {
							return func() tea.Msg {
								s.explorer.CreateNewFolder(name)
								return goatui.CloseModalMsg{}
							}
						},
						func() tea.Cmd {
							return func() tea.Msg {
								return goatui.CloseModalMsg{}
							}
						},
					),
				}
			}
		}
	}

	if s.explorer != nil {
		updated, cmd := s.explorer.Update(msg)
		if exp, ok := updated.(*goatui.FileExplorerScreen); ok {
			s.explorer = exp
			globalIconMode = exp.IconMode()
		}
		return s, cmd
	}
	return s, nil
}

func (s *FilesScreen) View(f *tea.Frame) {
	contentArea := s.renderChrome(f, "Files")
	if s.explorer != nil && !contentArea.IsEmpty() {
		s.explorer.DrawInArea(f, contentArea)
	}
}

// -----------------------------------------------------------------------------
// 6. Media, Animation & Physics Engine Showcase Screen
// -----------------------------------------------------------------------------

type MediaScreen struct {
	appScreen
	animCtrl     *goatui.AnimationController
	springSim    *goatui.SpringSimulation
	imageWidget  *goatui.ImageWidget
	videoPlayer  *goatui.VideoPlayerWidget
	springTarget float64
	ticks        int
	rawImage     image.Image
	useBraille   bool
	brailleRend  *goatui.BrailleRenderer
	particles    *goatui.ParticleSystem
}

func smoothstep(edge0, edge1, x float64) float64 {
	t := (x - edge0) / (edge1 - edge0)
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return t * t * (3 - 2*t)
}

func newMediaScreen() *MediaScreen {
	// 1. Controller for easing curves (PingPong 0.0 .. 1.0)
	ctrl := goatui.NewAnimationController(1600*time.Millisecond, goatui.Linear)
	ctrl.SetMode(goatui.LoopPingPong)
	ctrl.Play()

	// 2. Physics Spring Simulation (Bouncy physics)
	spring := goatui.NewSpringSimulation(0.0, 100.0, goatui.BouncySpring())

	// 3. TrueColor 24-bit RGB Plasma image
	img := image.NewRGBA(image.Rect(0, 0, 48, 36))
	for y := 0; y < 36; y++ {
		for x := 0; x < 48; x++ {
			dx := float64(x - 24)
			dy := float64(y - 18)
			dist := math.Sqrt(dx*dx + dy*dy)
			r := uint8(math.Sin(float64(x)*0.18)*127 + 128)
			g := uint8(math.Cos(float64(y)*0.22)*127 + 128)
			b := uint8(math.Sin(dist*0.3)*127 + 128)
			img.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
		}
	}
	imgWidget := goatui.NewImageWidget().SetImage(img).SetScaleMode(goatui.ScaleFit)

	// 4. In-Memory Video Frame Sequence (12 rotating frames at 15 FPS)
	frames := make([]image.Image, 12)
	delays := make([]time.Duration, 12)
	for f := 0; f < 12; f++ {
		fImg := image.NewRGBA(image.Rect(0, 0, 144, 96))
		angle := float64(f) * (2 * math.Pi / 12.0)
		cx0, cy0 := 72.0, 48.0
		maxRad := 36.0
		for y := 0; y < 96; y++ {
			for x := 0; x < 144; x++ {
				dx := float64(x) - cx0
				dy := float64(y) - cy0
				rad := math.Sqrt(dx*dx + dy*dy)
				theta := math.Atan2(dy, dx) + angle

				// Smoothstep antialiased edge (1.5 pixel transition)
				edge := smoothstep(maxRad+0.75, maxRad-0.75, rad)
				if edge < 0.01 {
					fImg.Set(x, y, color.RGBA{R: 25, G: 25, B: 35, A: 255})
					continue
				}

				hueR := uint8(float64(int((math.Sin(theta)+1.0)*127)) * edge)
				hueG := uint8(float64(int((math.Sin(theta+2.094)+1.0)*127)) * edge)
				hueB := uint8(float64(int((math.Sin(theta+4.188)+1.0)*127)) * edge)
				// Blend with background
				bgR, bgG, bgB := uint8(25), uint8(25), uint8(35)
				finalR := uint8(float64(hueR)*edge + float64(bgR)*(1-edge))
				finalG := uint8(float64(hueG)*edge + float64(bgG)*(1-edge))
				finalB := uint8(float64(hueB)*edge + float64(bgB)*(1-edge))
				fImg.Set(x, y, color.RGBA{R: finalR, G: finalG, B: finalB, A: 255})
			}
		}
		frames[f] = fImg
		delays[f] = 66 * time.Millisecond
	}
	videoWidget := goatui.NewVideoPlayer()
	videoWidget.SetFrames(frames, delays)
	videoWidget.Play()

	return &MediaScreen{
		animCtrl:     ctrl,
		springSim:    spring,
		imageWidget:  imgWidget,
		videoPlayer:  videoWidget,
		springTarget: 100.0,
		rawImage:     img,
		brailleRend:  goatui.NewBrailleRenderer(),
		particles:    goatui.NewParticleSystem(goatui.NewRect(0, 0, 80, 24), 25.0),
	}
}

func (s *MediaScreen) Init() tea.Cmd {
	return goatui.Tick(20*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (s *MediaScreen) Update(msg tea.Msg) (goatui.Screen, tea.Cmd) {
	if handled, cmd := handleCommonKeys(msg); handled {
		return s, cmd
	}

	switch m := msg.(type) {
	case tea.KeyMsg:
		switch m.Key.Rune {
		case 'm', 'M':
			s.useBraille = !s.useBraille
		case 'p', 'P':
			s.videoPlayer.Toggle()
		case 'n', 'N':
			s.videoPlayer.NextFrame()
		case 'c', 'C':
			s.particles.EmitConfetti(45, 40, 4)
			s.particles.EmitSparks(25, 40, 4)
		case 'f', 'F':
			mode := (s.imageWidget.ScaleMode() + 1) % 3
			s.imageWidget.SetScaleMode(mode)
			s.videoPlayer.SetScaleMode(mode)
		case 'b', 'B':
			// Flip spring simulation target
			if s.springTarget > 50.0 {
				s.springTarget = 0.0
			} else {
				s.springTarget = 100.0
			}
			s.springSim = goatui.NewSpringSimulation(s.springSim.Current, s.springTarget, goatui.BouncySpring())
			s.particles.EmitSparks(30, 40, 8)
		}

	case tickMsg:
		s.ticks++
		globalPulsePhase += 0.08

		// Advance animation controller
		s.animCtrl.Advance(20 * time.Millisecond)

		// Advance spring physics (20ms step)
		s.springSim.Update(0.02)

		// Advance video player frame
		s.videoPlayer.Advance(20 * time.Millisecond)

		// Advance particle simulation
		s.particles.Update(0.02)

		return s, goatui.Tick(20*time.Millisecond, func(t time.Time) tea.Msg {
			return tickMsg(t)
		})
	}

	return s, nil
}

func (s *MediaScreen) View(f *tea.Frame) {
	contentArea := s.renderChrome(f, "Media & Physics Engine")
	if contentArea.IsEmpty() {
		return
	}

	f.Buffer.SetClip(contentArea)
	defer f.Buffer.ResetClip()

	curTheme := goatui.DefaultTheme().Current()
	p := curTheme.Colors

	cols := goatui.SplitHorizontal(contentArea,
		goatui.Percent(50),
		goatui.Percent(50),
	)

	// Left: Capabilities & Physics / Easing
	leftRows := goatui.SplitVertical(cols[0],
		goatui.Fixed(8), // Capabilities Box
		goatui.Flex(1),  // Spring & Easing Playground
	)

	// Box 1: Capabilities
	capBox := goatui.NewStyle().
		Border(curTheme.Borders.Border).
		BorderForeground(p.Secondary).
		Title("Terminal Capability & Protocol Detector")
	capBox.Draw(f.Buffer, leftRows[0], "")
	innerCap := leftRows[0].Inset(1, 1)
	if !innerCap.IsEmpty() {
		f.Buffer.SetString(innerCap.X, innerCap.Y, "• Active Graphics Protocol: "+goatui.CurrentProtocol().String(), p.Accent, goatui.DefaultColor(), goatui.AttrBold)
		f.Buffer.SetString(innerCap.X, innerCap.Y+1, "• Detection Strategy: Single-Pass (sync.Once evaluated strictly once)", p.Foreground, goatui.DefaultColor(), 0)
		f.Buffer.SetString(innerCap.X, innerCap.Y+2, "• Priority Chain: Kitty -> iTerm2 -> Sixel -> Half-Block TrueColor -> Braille", p.Muted, goatui.DefaultColor(), 0)
		f.Buffer.SetString(innerCap.X, innerCap.Y+3, "• Half-Block Resolution: 2 vertical pixels/cell via ▀ (TrueColor 24-bit RGB)", p.Success, goatui.DefaultColor(), 0)
		f.Buffer.SetString(innerCap.X, innerCap.Y+4, "• Windows Terminal / PowerShell / cmd.exe: 100% Native TrueColor Compatibility", p.Primary, goatui.DefaultColor(), goatui.AttrBold)
	}

	// Box 2: Physics & Easing curves
	physBox := goatui.NewStyle().
		Border(curTheme.Borders.Border).
		BorderForeground(p.Primary).
		Title("Harmonic Spring Physics & Easing Curves")
	physBox.Draw(f.Buffer, leftRows[1], "")
	innerPhys := leftRows[1].Inset(1, 1)
	if !innerPhys.IsEmpty() {
		f.Buffer.SetString(innerPhys.X, innerPhys.Y, "Spring Harmonic Oscillation (Press [B] for impulse, [C] Particles):", p.Primary, goatui.DefaultColor(), goatui.AttrBold)

		// Draw physics track
		trackW := max(10, innerPhys.Width-16)
		posPct := s.springSim.Current / 100.0
		if posPct < 0 {
			posPct = 0
		} else if posPct > 1 {
			posPct = 1
		}
		ballPos := int(math.Round(float64(trackW-1) * posPct))

		var track strings.Builder
		track.WriteString("[ ")
		for i := 0; i < trackW; i++ {
			if i == ballPos {
				track.WriteString("●")
			} else {
				track.WriteString("─")
			}
		}
		track.WriteString(" ]")
		f.Buffer.SetString(innerPhys.X, innerPhys.Y+1, track.String(), p.Accent, goatui.DefaultColor(), goatui.AttrBold)
		f.Buffer.SetString(innerPhys.X, innerPhys.Y+2, fmt.Sprintf("  Pos: %.1f | Vel: %.1f | Settled: %v", s.springSim.Current, s.springSim.Velocity, s.springSim.Settled), p.Foreground, goatui.DefaultColor(), goatui.AttrDim)

		// Easing Curves Race
		f.Buffer.SetString(innerPhys.X, innerPhys.Y+4, "Real-time Easing Curves Comparison (Ping-Pong 60 FPS):", p.Primary, goatui.DefaultColor(), goatui.AttrBold)
		t := s.animCtrl.Progress()

		drawCurve := func(y int, name string, val float64, col goatui.Color) {
			curveW := max(8, innerPhys.Width-22)
			if val < 0 {
				val = 0
			} else if val > 1 {
				val = 1
			}
			dotX := int(math.Round(float64(curveW-1) * val))
			var bar strings.Builder
			for i := 0; i < curveW; i++ {
				if i == dotX {
					bar.WriteString("◆")
				} else {
					bar.WriteString("·")
				}
			}
			f.Buffer.SetString(innerPhys.X, y, fmt.Sprintf("%-16s [%s]", name, bar.String()), col, goatui.DefaultColor(), 0)
		}

		drawCurve(innerPhys.Y+6, "Linear", goatui.Linear(t), p.Foreground)
		drawCurve(innerPhys.Y+7, "EaseInOutCubic", goatui.EaseInOutCubic(t), p.Info)
		drawCurve(innerPhys.Y+8, "EaseOutBounce", goatui.EaseOutBounce(t), p.Warning)
		drawCurve(innerPhys.Y+9, "EaseOutElastic", goatui.EaseOutElastic(t), p.Accent)
	}

	// Right: TrueColor Image & Video Player
	rightRows := goatui.SplitVertical(cols[1],
		goatui.Flex(1), // TrueColor Image Widget
		goatui.Flex(1), // Video & GIF Player Widget
	)

	// Box 3: Image Widget (HalfBlock or Braille)
	modeNames := []string{"ScaleFit", "ScaleFill", "ScaleStretch"}
	imgTitle := fmt.Sprintf(" 24-bit TrueColor HalfBlock (Mode: %s [F], [M]=Braille) ", modeNames[s.imageWidget.ScaleMode()])
	if s.useBraille {
		imgTitle = " Braille 8-Dot High-Res (Press [M] to Toggle HalfBlock) "
	}
	imgBox := goatui.NewStyle().
		Border(curTheme.Borders.Border).
		BorderForeground(p.Secondary).
		Title(imgTitle)
	imgBox.Draw(f.Buffer, rightRows[0], "")
	innerImg := rightRows[0].Inset(1, 1)
	if !innerImg.IsEmpty() {
		if s.useBraille {
			s.brailleRend.DrawImage(f.Buffer, innerImg, s.rawImage)
		} else {
			s.imageWidget.Draw(f.Buffer, innerImg)
		}
	}

	// Box 4: Video Player Widget
	status := "PLAYING"
	if !s.videoPlayer.IsPlaying() {
		status = "PAUSED"
	}
	vidTitle := fmt.Sprintf(" Video/GIF Player (%s [P], Next [N]) ", status)
	vidBox := goatui.NewStyle().
		Border(curTheme.Borders.Border).
		BorderForeground(p.Accent).
		Title(vidTitle)
	vidBox.Draw(f.Buffer, rightRows[1], "")
	innerVid := rightRows[1].Inset(1, 1)
	if !innerVid.IsEmpty() {
		s.videoPlayer.Draw(f.Buffer, innerVid)
		infoStr := fmt.Sprintf("Frame %d/%d (Press P to toggle, N to step)", s.videoPlayer.CurrentFrame()+1, s.videoPlayer.TotalFrames())
		f.Buffer.SetString(innerVid.X+1, innerVid.Bottom()-1, infoStr, p.Foreground, goatui.DefaultColor(), goatui.AttrBold)
	}

	// 5. Particle Effects Overlay
	s.particles.SetBounds(contentArea)
	s.particles.Draw(f.Buffer, contentArea)
}

// -----------------------------------------------------------------------------
// Desktop App Root TEA Model (Handles Global Role Toggling & Access Restoration)
// -----------------------------------------------------------------------------

type desktopApp struct {
	wm *goatui.WindowManager
}

func (a *desktopApp) Init() tea.Cmd {
	return a.wm.Init()
}

func (a *desktopApp) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case goatui.ToggleRoleMsg:
		toggleAdminUser()
		// If currently on AccessDeniedScreen, navigate back to target path or /admin
		if denied, ok := a.wm.ActiveScreen().(*goatui.AccessDeniedScreen); ok {
			target := "/admin"
			if denied.Path != "" {
				target = denied.Path
			}
			return a, goatui.Navigate(target)
		}
		return a, nil

	case tea.KeyMsg:
		// Global 't' / 'T' hotkey toggles role from ANY screen unless Omnibar text input is active
		if (msg.Key.Rune == 't' || msg.Key.Rune == 'T') && !a.wm.Omnibar().IsVisible() {
			toggleAdminUser()
			user := goatui.DefaultSecurity().CurrentUser()
			// If on AccessDeniedScreen and we just became admin, immediately open the page!
			if denied, ok := a.wm.ActiveScreen().(*goatui.AccessDeniedScreen); ok {
				if user != nil && user.HasRole("admin") {
					target := "/admin"
					if denied.Path != "" {
						target = denied.Path
					}
					return a, goatui.Navigate(target)
				}
			}
			// If on /admin and we just dropped admin, navigate to /admin so RouteGuard triggers 403!
			if a.wm.ActiveScreen() != nil && user != nil && !user.HasRole("admin") {
				return a, goatui.Navigate("/admin")
			}
			return a, nil
		}
	}

	model, cmd := a.wm.Update(msg)
	if m, ok := model.(*goatui.WindowManager); ok {
		a.wm = m
	}
	return a, cmd
}

func (a *desktopApp) View(f *tea.Frame) {
	a.wm.View(f)
}

// -----------------------------------------------------------------------------
// Main Application Entrypoint
// -----------------------------------------------------------------------------

func main() {
	// 1. Initialize Security Manager with default admin user
	secMgr := goatui.DefaultSecurity()
	secMgr.Login(&goatui.User{
		ID:          "usr_001",
		Username:    "alice",
		Roles:       []string{"admin", "user"},
		Permissions: []string{"*"},
	}, "bearer-token-desktop-demo")

	// 2. Initialize Theme Manager & load custom YAML themes from ./themes
	themeMgr := goatui.DefaultTheme()
	_, _ = themeMgr.LoadDirectory("themes")
	// Watch themes directory for live changes
	_ = themeMgr.WatchDirectory("themes", 1*time.Second, func(t theme.Theme) {
		// Live reloaded
	})

	// 3. Build URL Router with Path Parameters & RBAC RouteGuard
	r := goatui.NewRouter()

	r.Handle("/", func(ctx *goatui.RouteContext) any {
		return newDashboardScreen()
	})
	r.Handle("/dashboard", func(ctx *goatui.RouteContext) any {
		return newDashboardScreen()
	})
	r.Handle("/processes", func(ctx *goatui.RouteContext) any {
		return newProcessesScreen()
	})
	r.Handle("/processes/:id", func(ctx *goatui.RouteContext) any {
		return newProcessDetailScreen(ctx.Param("id"))
	})
	// Route Guard: /admin strictly requires role "admin"
	r.Handle("/admin", func(ctx *goatui.RouteContext) any {
		return newAdminScreen()
	}, goatui.RequireRole("admin", ""))

	r.Handle("/settings", func(ctx *goatui.RouteContext) any {
		return newSettingsScreen()
	})
	r.Handle("/files", func(ctx *goatui.RouteContext) any {
		return newFilesScreen(".")
	})
	r.Handle("/files/*", func(ctx *goatui.RouteContext) any {
		target := strings.TrimPrefix(ctx.Path, "/files")
		target = strings.TrimPrefix(target, "/")
		if target == "" {
			target = "."
		}
		return newFilesScreen(target)
	})
	r.Handle("file://*", func(ctx *goatui.RouteContext) any {
		path := ctx.Path
		if strings.HasPrefix(path, "file://") {
			path = strings.TrimPrefix(path, "file://")
		}
		if len(path) >= 3 && path[0] == '/' && path[2] == ':' {
			path = path[1:]
		}
		if path == "" || path == "/" {
			path = "."
		}
		return newFilesScreen(path)
	})
	r.Handle("/media", func(ctx *goatui.RouteContext) any {
		return newMediaScreen()
	})
	r.Handle("http://*", func(ctx *goatui.RouteContext) any {
		return goatui.NewHTTPViewerScreen(ctx.Path)
	})
	r.Handle("https://*", func(ctx *goatui.RouteContext) any {
		return goatui.NewHTTPViewerScreen(ctx.Path)
	})

	// 4. Create WindowManager
	wm := goatui.NewWindowManager(r)

	// Populate Omnibar with custom action commands
	ob := wm.Omnibar()
	ob.AddItem(goatui.OmniItem{
		Title:       "nav:files",
		Description: "Open Dual-Pane File Explorer (Live text/media preview)",
		Action: func() tea.Cmd {
			return goatui.Navigate("/files")
		},
	})
	ob.AddItem(goatui.OmniItem{
		Title:       "nav:media",
		Description: "Open Media & Physics Engine Showcase (GIF/Image/Springs)",
		Action: func() tea.Cmd {
			return goatui.Navigate("/media")
		},
	})
	ob.AddItem(goatui.OmniItem{
		Title:       "media:toggle",
		Description: "Toggle media playback (Play/Pause)",
		Action: func() tea.Cmd {
			return func() tea.Msg {
				return tea.KeyMsg{Key: goatui.Key{Type: goatui.KeyRune, Rune: 'p'}}
			}
		},
	})
	ob.AddItem(goatui.OmniItem{
		Title:       "anim:bounce",
		Description: "Trigger spring physics impulse",
		Action: func() tea.Cmd {
			return func() tea.Msg {
				return tea.KeyMsg{Key: goatui.Key{Type: goatui.KeyRune, Rune: 'b'}}
			}
		},
	})
	ob.AddItem(goatui.OmniItem{
		Title:       "theme:goatdark",
		Description: "Switch theme to Goat Dark",
		Action: func() tea.Cmd {
			return goatui.SwitchTheme("Goat Dark")
		},
	})
	ob.AddItem(goatui.OmniItem{
		Title:       "theme:dracula",
		Description: "Switch theme to Dracula",
		Action: func() tea.Cmd {
			return goatui.SwitchTheme("Dracula")
		},
	})
	ob.AddItem(goatui.OmniItem{
		Title:       "theme:nord",
		Description: "Switch theme to Nord",
		Action: func() tea.Cmd {
			return goatui.SwitchTheme("Nord")
		},
	})
	ob.AddItem(goatui.OmniItem{
		Title:       "theme:catppuccin",
		Description: "Switch theme to Catppuccin Mocha",
		Action: func() tea.Cmd {
			return goatui.SwitchTheme("Catppuccin Mocha")
		},
	})
	ob.AddItem(goatui.OmniItem{
		Title:       "theme:monokai",
		Description: "Switch theme to Monokai",
		Action: func() tea.Cmd {
			return goatui.SwitchTheme("Monokai")
		},
	})
	ob.AddItem(goatui.OmniItem{
		Title:       "theme:matrix",
		Description: "Switch theme to Matrix (from YAML file)",
		Action: func() tea.Cmd {
			return goatui.SwitchTheme("Matrix")
		},
	})
	ob.AddItem(goatui.OmniItem{
		Title:       "theme:cyberpunk",
		Description: "Switch theme to Cyberpunk (from YAML file)",
		Action: func() tea.Cmd {
			return goatui.SwitchTheme("Cyberpunk")
		},
	})
	ob.AddItem(goatui.OmniItem{
		Title:       "admin:toggle",
		Description: "Toggle admin role (Admin <=> Guest)",
		Action: func() tea.Cmd {
			return func() tea.Msg { return goatui.ToggleRoleMsg{} }
		},
	})
	ob.AddItem(goatui.OmniItem{
		Title:       "modal:confirm",
		Description: "Open sample Confirmation Modal",
		Action: func() tea.Cmd {
			modal := goatui.NewConfirmModal("Sample Confirmation", "Execute system diagnostics now?", func() tea.Cmd {
				return func() tea.Msg { return goatui.CloseModalMsg{} }
			}, func() tea.Cmd {
				return func() tea.Msg { return goatui.CloseModalMsg{} }
			})
			return func() tea.Msg { return goatui.ShowModalMsg{Modal: modal} }
		},
	})
	ob.AddItem(goatui.OmniItem{
		Title:       "icons:ascii",
		Description: "Switch File Explorer icons to ASCII [DIR] [SRC] (Zero drift)",
		Action: func() tea.Cmd {
			globalIconMode = goatui.IconModeASCII
			return nil
		},
	})
	ob.AddItem(goatui.OmniItem{
		Title:       "icons:unicode",
		Description: "Switch File Explorer icons to Unicode geometric symbols (■ · §)",
		Action: func() tea.Cmd {
			globalIconMode = goatui.IconModeUnicode
			return nil
		},
	})
	ob.AddItem(goatui.OmniItem{
		Title:       "icons:nerdfont",
		Description: "Switch File Explorer icons to NerdFont glyphs",
		Action: func() tea.Cmd {
			globalIconMode = goatui.IconModeNerdFont
			return nil
		},
	})
	ob.AddItem(goatui.OmniItem{
		Title:       "icons:emoji",
		Description: "Switch File Explorer icons to Emoji mode",
		Action: func() tea.Cmd {
			globalIconMode = goatui.IconModeEmoji
			return nil
		},
	})
	ob.AddItem(goatui.OmniItem{
		Title:       "media:braille",
		Description: "Toggle Braille 8-dot high-res rendering in Media tab",
		Action: func() tea.Cmd {
			return func() tea.Msg {
				return tea.KeyMsg{Key: goatui.Key{Type: goatui.KeyRune, Rune: 'm'}}
			}
		},
	})
	ob.AddItem(goatui.OmniItem{
		Title:       "app:quit",
		Description: "Exit the desktop application",
		Action: func() tea.Cmd {
			return goatui.Quit
		},
	})

	// Initial Screen: Mount Dashboard
	initCtx := goatui.NewRouteContext("/dashboard", r.Session())
	wm.Push(newDashboardScreen(), initCtx)

	// 5. Run with TEA Program
	app := &desktopApp{wm: wm}
	p := goatui.NewProgram(app)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
