package window

import (
	"strings"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/driver/input"
	"github.com/baibeicha/goatui/pkg/style"
	"github.com/baibeicha/goatui/pkg/tea"
)

// Modal extends Screen with geometry bounds and backdrop dimming control.
type Modal interface {
	Screen
	Bounds(area buffer.Rect) buffer.Rect
	BackdropDim() bool
}

// CloseModalMsg is a command message that requests closing the active modal.
type CloseModalMsg struct{}

// CloseModal returns a command to dismiss the active modal.
func CloseModal() tea.Msg {
	return CloseModalMsg{}
}

// ShowModalMsg is a command message that requests opening a modal.
type ShowModalMsg struct {
	Modal Modal
}

// ShowModal returns a command to display a modal dialog.
func ShowModal(m Modal) tea.Cmd {
	return func() tea.Msg {
		return ShowModalMsg{Modal: m}
	}
}

// AlertModal creates a simple alert dialog with an OK button.
func AlertModal(title, message string, onDismiss func() tea.Cmd) Modal {
	return &alertModalImpl{
		title:     title,
		message:   message,
		onDismiss: onDismiss,
	}
}

type alertModalImpl struct {
	BaseScreen
	title     string
	message   string
	onDismiss func() tea.Cmd
	btnBounds buffer.Rect
}

func (a *alertModalImpl) Bounds(area buffer.Rect) buffer.Rect {
	lines := strings.Split(a.message, "\n")
	maxW := buffer.StringWidth(a.title)
	for _, l := range lines {
		if lw := buffer.StringWidth(l); lw > maxW {
			maxW = lw
		}
	}
	w := max(44, maxW+6)
	w = min(w, max(10, area.Width-4))
	h := len(lines) + 6
	h = min(h, max(5, area.Height-2))
	x := area.X + max(0, (area.Width-w)/2)
	y := area.Y + max(0, (area.Height-h)/2)
	b := buffer.NewRect(x, y, w, h)
	inner := b.Inset(1, 1)
	if !inner.IsEmpty() {
		btnY := min(inner.Bottom()-1, inner.Y+2+len(lines)+1)
		a.btnBounds = buffer.NewRect(inner.X+max(0, (inner.Width-10)/2), btnY, 10, 1)
	}
	return b
}

func (a *alertModalImpl) BackdropDim() bool {
	return true
}

func (a *alertModalImpl) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Key.Type {
		case input.KeyEnter, input.KeyEsc, input.KeySpace:
			if a.onDismiss != nil {
				return a, tea.Batch(func() tea.Msg { return CloseModalMsg{} }, a.onDismiss())
			}
			return a, func() tea.Msg { return CloseModalMsg{} }
		}
	case tea.HitMsg:
		if msg.IsLeftClick() && msg.ID() == "btn-alert-ok" {
			if a.onDismiss != nil {
				return a, tea.Batch(func() tea.Msg { return CloseModalMsg{} }, a.onDismiss())
			}
			return a, func() tea.Msg { return CloseModalMsg{} }
		}
	case tea.MouseMsg:
		if msg.Action == input.MousePress && msg.Button == input.MouseLeft {
			if !a.btnBounds.IsEmpty() && a.btnBounds.Contains(msg.X, msg.Y) {
				if a.onDismiss != nil {
					return a, tea.Batch(func() tea.Msg { return CloseModalMsg{} }, a.onDismiss())
				}
				return a, func() tea.Msg { return CloseModalMsg{} }
			}
		}
	}
	return a, nil
}

func (a *alertModalImpl) View(f *tea.Frame) {
	bounds := a.Bounds(f.Area())
	if bounds.IsEmpty() {
		return
	}

	// Box border and solid opaque background
	st := style.NewStyle().
		Border(style.BorderRounded).
		BorderForeground(cell.ColorHex("#00D2FF")).
		Background(cell.Color256(234))
	st.Draw(f.Buffer, bounds, "")

	inner := bounds.Inset(1, 1)
	if inner.IsEmpty() {
		return
	}

	// Title
	titleX := inner.X + max(0, (inner.Width-buffer.StringWidth(a.title))/2)
	f.Buffer.SetString(titleX, inner.Y, a.title, cell.ColorHex("#00FFAA"), cell.DefaultColor(), cell.AttrBold)

	// Message lines
	lines := strings.Split(a.message, "\n")
	maxLines := max(1, inner.Height-3)
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	for i, line := range lines {
		lineY := inner.Y + 2 + i
		if lineY >= inner.Bottom()-1 {
			break
		}
		f.Buffer.SetString(inner.X+1, lineY, line, cell.DefaultColor(), cell.DefaultColor(), cell.AttrNone)
	}

	// OK Button
	btnY := min(inner.Bottom()-1, inner.Y+2+len(lines)+1)
	btnArea := buffer.NewRect(inner.X+max(0, (inner.Width-10)/2), btnY, 10, 1)
	btnStyle := style.NewStyle().
		Bold(true).
		Foreground(cell.ColorHex("#000000")).
		Background(cell.ColorHex("#00D2FF")).
		AlignCenter()
	btnStyle.Draw(f.Buffer, btnArea, "[ OK ]")
	f.RegisterHit("btn-alert-ok", btnArea, 10, nil)
}

// ConfirmModal creates a dialog with Confirm and Cancel buttons.
func ConfirmModal(title, message string, onConfirm, onCancel func() tea.Cmd) Modal {
	return &confirmModalImpl{
		title:     title,
		message:   message,
		onConfirm: onConfirm,
		onCancel:  onCancel,
		focused:   0, // 0: Confirm, 1: Cancel
	}
}

type confirmModalImpl struct {
	BaseScreen
	title          string
	message        string
	onConfirm      func() tea.Cmd
	onCancel       func() tea.Cmd
	focused        int
	confirmBtnArea buffer.Rect
	cancelBtnArea  buffer.Rect
}

func (c *confirmModalImpl) Bounds(area buffer.Rect) buffer.Rect {
	lines := strings.Split(c.message, "\n")
	maxW := buffer.StringWidth(c.title)
	for _, l := range lines {
		if lw := buffer.StringWidth(l); lw > maxW {
			maxW = lw
		}
	}
	w := max(48, maxW+6)
	w = min(w, max(10, area.Width-4))
	h := len(lines) + 6
	h = min(h, max(5, area.Height-2))
	x := area.X + max(0, (area.Width-w)/2)
	y := area.Y + max(0, (area.Height-h)/2)
	b := buffer.NewRect(x, y, w, h)
	inner := b.Inset(1, 1)
	if !inner.IsEmpty() {
		btnW := 12
		totalBtnW := btnW*2 + 4
		startX := inner.X + max(0, (inner.Width-totalBtnW)/2)
		btnY := min(inner.Bottom()-1, inner.Y+2+len(lines)+1)
		c.confirmBtnArea = buffer.NewRect(startX, btnY, btnW, 1)
		c.cancelBtnArea = buffer.NewRect(startX+btnW+4, btnY, btnW, 1)
	}
	return b
}

func (c *confirmModalImpl) BackdropDim() bool {
	return true
}

func (c *confirmModalImpl) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Key.Type {
		case input.KeyTab, input.KeyBacktab, input.KeyLeft, input.KeyRight:
			c.focused = 1 - c.focused
		case input.KeyEsc:
			if c.onCancel != nil {
				return c, tea.Batch(func() tea.Msg { return CloseModalMsg{} }, c.onCancel())
			}
			return c, func() tea.Msg { return CloseModalMsg{} }
		case input.KeyEnter, input.KeySpace:
			if c.focused == 0 {
				if c.onConfirm != nil {
					return c, tea.Batch(func() tea.Msg { return CloseModalMsg{} }, c.onConfirm())
				}
			} else {
				if c.onCancel != nil {
					return c, tea.Batch(func() tea.Msg { return CloseModalMsg{} }, c.onCancel())
				}
			}
			return c, func() tea.Msg { return CloseModalMsg{} }
		}
	case tea.HitMsg:
		if msg.IsLeftClick() {
			switch msg.ID() {
			case "btn-confirm":
				if c.onConfirm != nil {
					return c, tea.Batch(func() tea.Msg { return CloseModalMsg{} }, c.onConfirm())
				}
				return c, func() tea.Msg { return CloseModalMsg{} }
			case "btn-cancel":
				if c.onCancel != nil {
					return c, tea.Batch(func() tea.Msg { return CloseModalMsg{} }, c.onCancel())
				}
				return c, func() tea.Msg { return CloseModalMsg{} }
			}
		}
	case tea.MouseMsg:
		if msg.Action == input.MousePress && msg.Button == input.MouseLeft {
			if !c.confirmBtnArea.IsEmpty() && c.confirmBtnArea.Contains(msg.X, msg.Y) {
				if c.onConfirm != nil {
					return c, tea.Batch(func() tea.Msg { return CloseModalMsg{} }, c.onConfirm())
				}
				return c, func() tea.Msg { return CloseModalMsg{} }
			}
			if !c.cancelBtnArea.IsEmpty() && c.cancelBtnArea.Contains(msg.X, msg.Y) {
				if c.onCancel != nil {
					return c, tea.Batch(func() tea.Msg { return CloseModalMsg{} }, c.onCancel())
				}
				return c, func() tea.Msg { return CloseModalMsg{} }
			}
		}
	}
	return c, nil
}

func (c *confirmModalImpl) View(f *tea.Frame) {
	bounds := c.Bounds(f.Area())
	if bounds.IsEmpty() {
		return
	}

	st := style.NewStyle().
		Border(style.BorderRounded).
		BorderForeground(cell.ColorHex("#FF007F")).
		Background(cell.Color256(234))
	st.Draw(f.Buffer, bounds, "")

	inner := bounds.Inset(1, 1)
	if inner.IsEmpty() {
		return
	}

	// Title
	titleX := inner.X + max(0, (inner.Width-buffer.StringWidth(c.title))/2)
	f.Buffer.SetString(titleX, inner.Y, c.title, cell.ColorHex("#FF5555"), cell.DefaultColor(), cell.AttrBold)

	// Message lines
	lines := strings.Split(c.message, "\n")
	maxLines := max(1, inner.Height-3)
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	for i, line := range lines {
		lineY := inner.Y + 2 + i
		if lineY >= inner.Bottom()-1 {
			break
		}
		f.Buffer.SetString(inner.X+1, lineY, line, cell.DefaultColor(), cell.DefaultColor(), cell.AttrNone)
	}

	// Buttons
	btnW := 12
	totalBtnW := btnW*2 + 4
	startX := inner.X + max(0, (inner.Width-totalBtnW)/2)
	btnY := min(inner.Bottom()-1, inner.Y+2+len(lines)+1)

	// Confirm button
	confirmArea := buffer.NewRect(startX, btnY, btnW, 1)
	cStyle := style.NewStyle().AlignCenter()
	if c.focused == 0 {
		cStyle = cStyle.Bold(true).Foreground(cell.ColorHex("#000000")).Background(cell.ColorHex("#FF5555"))
	} else {
		cStyle = cStyle.Foreground(cell.ColorHex("#FF5555")).Background(cell.Color256(235))
	}
	cStyle.Draw(f.Buffer, confirmArea, "[ Confirm ]")
	f.RegisterHit("btn-confirm", confirmArea, 10, nil)

	// Cancel button
	cancelArea := buffer.NewRect(startX+btnW+4, btnY, btnW, 1)
	cancelStyle := style.NewStyle().AlignCenter()
	if c.focused == 1 {
		cancelStyle = cancelStyle.Bold(true).Foreground(cell.ColorHex("#000000")).Background(cell.ColorHex("#AAAAAA"))
	} else {
		cancelStyle = cancelStyle.Foreground(cell.ColorHex("#AAAAAA")).Background(cell.Color256(235))
	}
	cancelStyle.Draw(f.Buffer, cancelArea, "[ Cancel ]")
	f.RegisterHit("btn-cancel", cancelArea, 10, nil)
}
