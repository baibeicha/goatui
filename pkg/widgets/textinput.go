package widgets

import (
	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/driver/input"
	"github.com/baibeicha/goatui/pkg/style"
	"github.com/baibeicha/goatui/pkg/tea"
)

// TextInput is an interactive single-line text entry field.
type TextInput struct {
	prompt      string
	placeholder string
	value       []rune
	cursor      int
	scrollOffset int
	focused     bool
	promptStyle style.Style
	textStyle   style.Style
}

// NewTextInput creates a new text input field.
func NewTextInput() *TextInput {
	return &TextInput{
		prompt:      "> ",
		value:       make([]rune, 0, 64),
		focused:     true,
		promptStyle: style.NewStyle().Bold(true).Foreground(cell.ColorHex("#7D56F4")),
		textStyle:   style.NewStyle(),
	}
}

// SetPrompt configures the leading prompt text.
func (ti *TextInput) SetPrompt(p string) {
	ti.prompt = p
}

// SetPlaceholder sets the ghost text when value is empty.
func (ti *TextInput) SetPlaceholder(ph string) {
	ti.placeholder = ph
}

// Value returns the current text string.
func (ti *TextInput) Value() string {
	return string(ti.value)
}

// SetValue updates the text content and resets cursor to end.
func (ti *TextInput) SetValue(s string) {
	ti.value = []rune(s)
	ti.cursor = len(ti.value)
}

// Focus enables input capture and cursor rendering.
func (ti *TextInput) Focus() {
	ti.focused = true
}

// Blur disables input capture.
func (ti *TextInput) Blur() {
	ti.focused = false
}

// HandleKey processes a keyboard event. Returns true if the key modified state.
func (ti *TextInput) HandleKey(k input.Key) bool {
	if !ti.focused {
		return false
	}

	switch k.Type {
	case input.KeyLeft:
		if ti.cursor > 0 {
			ti.cursor--
			return true
		}
	case input.KeyRight:
		if ti.cursor < len(ti.value) {
			ti.cursor++
			return true
		}
	case input.KeyHome:
		ti.cursor = 0
		return true
	case input.KeyEnd:
		ti.cursor = len(ti.value)
		return true
	case input.KeyBackspace:
		if ti.cursor > 0 {
			ti.value = append(ti.value[:ti.cursor-1], ti.value[ti.cursor:]...)
			ti.cursor--
			return true
		}
	case input.KeyDelete:
		if ti.cursor < len(ti.value) {
			ti.value = append(ti.value[:ti.cursor], ti.value[ti.cursor+1:]...)
			return true
		}
	case input.KeySpace:
		ti.insertRune(' ')
		return true
	case input.KeyRune:
		if k.Rune != 0 {
			ti.insertRune(k.Rune)
			return true
		}
	}

	return false
}

func (ti *TextInput) insertRune(r rune) {
	ti.value = append(ti.value[:ti.cursor], append([]rune{r}, ti.value[ti.cursor:]...)...)
	ti.cursor++
}

// Draw renders the text input field and cursor into area.
func (ti *TextInput) Draw(buf *buffer.Buffer, area buffer.Rect) {
	if area.IsEmpty() {
		return
	}

	// Draw Prompt
	curX := area.X
	promptLen := 0
	if ti.prompt != "" {
		promptLen = buffer.StringWidth(ti.prompt)
		curX = buf.SetString(curX, area.Y, ti.prompt, cell.ColorHex("#7D56F4"), cell.DefaultColor(), cell.AttrBold)
	}

	availWidth := area.Width - promptLen
	if availWidth <= 0 {
		return
	}

	// Clamp cursor
	if ti.cursor < 0 {
		ti.cursor = 0
	} else if ti.cursor > len(ti.value) {
		ti.cursor = len(ti.value)
	}

	// Adjust scrollOffset
	if ti.cursor < ti.scrollOffset {
		ti.scrollOffset = ti.cursor
	} else if ti.cursor >= ti.scrollOffset+availWidth {
		ti.scrollOffset = ti.cursor - availWidth + 1
	}
	if len(ti.value)-ti.scrollOffset < availWidth {
		ti.scrollOffset = max(0, len(ti.value)-availWidth+1)
	}

	visibleRunes := ti.value[ti.scrollOffset:]
	strToDraw := string(visibleRunes)

	if len(visibleRunes) == 0 && ti.placeholder != "" && !ti.focused {
		strToDraw = ti.placeholder
		buf.SetString(curX, area.Y, strToDraw, cell.ColorHex("#626262"), cell.DefaultColor(), cell.AttrItalic)
	} else {
		buf.SetString(curX, area.Y, strToDraw, cell.DefaultColor(), cell.DefaultColor(), cell.AttrNone)
	}

	// Draw cursor if focused
	if ti.focused {
		cursorVisualX := curX
		for i := ti.scrollOffset; i < ti.cursor && i < len(ti.value); i++ {
			cursorVisualX += buffer.RuneWidth(ti.value[i])
		}

		cursorRune := ' '
		if ti.cursor < len(ti.value) {
			cursorRune = ti.value[ti.cursor]
		}
		buf.Set(cursorVisualX, area.Y, cell.Cell{
			Rune:     cursorRune,
			Width:    1,
			Modifier: cell.AttrReverse,
		})
	}
}

// HandleMouse processes mouse events for the text input.
func (ti *TextInput) HandleMouse(msg tea.MouseMsg, area buffer.Rect) bool {
	if !area.Contains(msg.X, msg.Y) {
		return false
	}
	if msg.Action == input.MousePress && msg.Button == input.MouseLeft {
		if !ti.focused {
			ti.Focus()
		}
		promptLen := buffer.StringWidth(ti.prompt)
		relX := msg.X - area.X - promptLen
		if relX < 0 {
			relX = 0
		}
		newCursor := ti.scrollOffset
		currX := 0
		for i := ti.scrollOffset; i < len(ti.value); i++ {
			w := buffer.RuneWidth(ti.value[i])
			if currX+w > relX {
				break
			}
			currX += w
			newCursor++
		}
		ti.cursor = min(newCursor, len(ti.value))
		return true
	}
	return false
}
