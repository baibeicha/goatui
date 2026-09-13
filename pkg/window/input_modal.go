package window

import (
	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/driver/input"
	"github.com/baibeicha/goatui/pkg/style"
	"github.com/baibeicha/goatui/pkg/tea"
	"github.com/baibeicha/goatui/pkg/widgets"
)

// InputModal provides a dialog with a text input field.
type InputModal struct {
	BaseScreen
	title     string
	prompt    string
	input     *widgets.TextInput
	onConfirm func(string) tea.Cmd
	onCancel  func() tea.Cmd
}

// NewInputModal creates a new input dialog.
func NewInputModal(title, prompt, defaultValue string, onConfirm func(string) tea.Cmd, onCancel func() tea.Cmd) *InputModal {
	ti := widgets.NewTextInput()
	ti.SetPrompt("> ")
	ti.SetPlaceholder("")
	ti.SetValue(defaultValue)
	ti.Focus()

	return &InputModal{
		title:     title,
		prompt:    prompt,
		input:     ti,
		onConfirm: onConfirm,
		onCancel:  onCancel,
	}
}

func (i *InputModal) Bounds(area buffer.Rect) buffer.Rect {
	w := max(50, buffer.StringWidth(i.prompt)+10)
	w = min(w, area.Width-4)
	h := 8
	x := area.X + max(0, (area.Width-w)/2)
	y := area.Y + max(0, (area.Height-h)/2)
	return buffer.NewRect(x, y, w, h)
}

func (i *InputModal) BackdropDim() bool {
	return true
}

func (i *InputModal) Update(msg tea.Msg) (Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Key.Type {
		case input.KeyEsc:
			if i.onCancel != nil {
				return i, tea.Sequence(func() tea.Msg { return CloseModalMsg{} }, i.onCancel())
			}
			return i, func() tea.Msg { return CloseModalMsg{} }
		case input.KeyEnter:
			val := i.input.Value()
			if i.onConfirm != nil {
				return i, tea.Sequence(func() tea.Msg { return CloseModalMsg{} }, i.onConfirm(val))
			}
			return i, func() tea.Msg { return CloseModalMsg{} }
		default:
			i.input.HandleKey(msg.Key)
		}
	}
	return i, nil
}

func (i *InputModal) View(f *tea.Frame) {
	bounds := i.Bounds(f.Area())
	if bounds.IsEmpty() {
		return
	}

	st := style.NewStyle().
		Border(style.BorderRounded).
		BorderForeground(cell.ColorHex("#7D56F4")).
		Background(cell.Color256(234))
	st.Draw(f.Buffer, bounds, "")

	inner := bounds.Inset(1, 1)
	if inner.IsEmpty() {
		return
	}

	// Title
	titleX := inner.X + max(0, (inner.Width-buffer.StringWidth(i.title))/2)
	f.Buffer.SetString(titleX, inner.Y, i.title, cell.ColorHex("#00FFAA"), cell.Color256(234), cell.AttrBold)

	// Prompt
	f.Buffer.SetString(inner.X, inner.Y+2, i.prompt, cell.DefaultColor(), cell.Color256(234), cell.AttrNone)

	// Input
	inputArea := buffer.NewRect(inner.X, inner.Y+3, inner.Width, 1)
	i.input.Draw(f.Buffer, inputArea)

	// Buttons
	hint := "[Enter] OK   [Esc] Cancel"
	hintX := inner.X + max(0, (inner.Width-buffer.StringWidth(hint))/2)
	f.Buffer.SetString(hintX, inner.Y+5, hint, cell.ColorHex("#777799"), cell.Color256(234), cell.AttrNone)
}
