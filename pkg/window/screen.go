package window

import (
	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/router"
	"github.com/baibeicha/goatui/pkg/style"
	"github.com/baibeicha/goatui/pkg/tea"
)

// ToggleRoleMsg requests toggling between Admin and Guest role globally.
type ToggleRoleMsg struct{}

// ToggleRole returns a command to toggle user role.
func ToggleRole() tea.Msg {
	return ToggleRoleMsg{}
}

// Screen defines the complete contract and lifecycle for a fullscreen scene or page.
type Screen interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (Screen, tea.Cmd)
	View(f *tea.Frame)

	// Lifecycle hooks
	OnMount(ctx *router.RouteContext)
	OnPause()
	OnResume()
	OnDestroy()
}

// Destroyer is an optional interface for screens to clean up resources.
type Destroyer interface {
	OnDestroy()
}

// BaseScreen provides default no-op implementations for all Screen lifecycle methods.
type BaseScreen struct{}

func (b *BaseScreen) Init() tea.Cmd {
	return nil
}

func (b *BaseScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	return b, nil
}

func (b *BaseScreen) View(f *tea.Frame) {}

func (b *BaseScreen) OnMount(ctx *router.RouteContext) {}

func (b *BaseScreen) OnPause() {}

func (b *BaseScreen) OnResume() {}

func (b *BaseScreen) OnDestroy() {}

// AccessDeniedScreen is displayed when route guards reject navigation.
type AccessDeniedScreen struct {
	BaseScreen
	Reason string
	Path   string
}

// NewAccessDeniedScreen creates a 403 Access Denied screen.
func NewAccessDeniedScreen(reason, path string) *AccessDeniedScreen {
	return &AccessDeniedScreen{Reason: reason, Path: path}
}

func (s *AccessDeniedScreen) OnMount(ctx *router.RouteContext) {
	if ctx != nil && s.Path == "" {
		s.Path = ctx.Path
	}
}

func (s *AccessDeniedScreen) View(f *tea.Frame) {
	area := f.Area()
	w := min(56, area.Width-4)
	h := 9
	x := area.X + max(0, (area.Width-w)/2)
	y := area.Y + max(0, (area.Height-h)/2)
	boxArea := buffer.NewRect(x, y, w, h)

	st := style.NewStyle().
		Border(style.BorderRounded).
		BorderForeground(cell.ColorHex("#FF0055")).
		Background(cell.Color256(234))
	st.Draw(f.Buffer, boxArea, " 403 - Access Denied ")

	inner := boxArea.Inset(2, 1)
	if inner.IsEmpty() {
		return
	}

	if s.Reason != "" {
		f.Buffer.SetString(inner.X, inner.Y, s.Reason, cell.ColorHex("#FFB86C"), cell.Color256(234), cell.AttrNone)
	}
	if s.Path != "" {
		f.Buffer.SetString(inner.X, inner.Y+1, "Requested URL: "+s.Path, cell.ColorHex("#AAAAAA"), cell.Color256(234), cell.AttrNone)
	}

	f.Buffer.SetString(inner.X, inner.Y+3, "Quick Actions:", cell.ColorHex("#BD93F9"), cell.Color256(234), cell.AttrBold)
	f.Buffer.SetString(inner.X, inner.Y+4, " [T]   Elevate / Toggle Role (Become Admin)", cell.ColorHex("#00FFAA"), cell.Color256(234), cell.AttrBold)
	f.Buffer.SetString(inner.X, inner.Y+5, " [Esc] Go back | [1] Dashboard | [4] Settings", cell.ColorHex("#6272A4"), cell.Color256(234), cell.AttrDim)
}

func (s *AccessDeniedScreen) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Key.Type {
		case 2: // Esc
			return s, func() tea.Msg { return PopMsg{} }
		}
		switch msg.Key.Rune {
		case 't', 'T':
			return s, func() tea.Msg { return ToggleRoleMsg{} }
		case '1':
			return s, func() tea.Msg { return NavigateMsg{URL: "/dashboard"} }
		case '2':
			return s, func() tea.Msg { return NavigateMsg{URL: "/processes"} }
		case '4':
			return s, func() tea.Msg { return NavigateMsg{URL: "/settings"} }
		}
	}
	return s, nil
}
