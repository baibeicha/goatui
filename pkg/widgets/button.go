package widgets

import (
	"unicode"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/driver/input"
	"github.com/baibeicha/goatui/pkg/style"
	"github.com/baibeicha/goatui/pkg/tea"
)

// ButtonState represents the visual/interaction state of a button.
type ButtonState int

const (
	ButtonNormal ButtonState = iota
	ButtonHover
	ButtonPressed
	ButtonDisabled
)

// ButtonVariant defines the semantic color scheme for a button.
type ButtonVariant int

const (
	ButtonVariantDefault ButtonVariant = iota
	ButtonVariantPrimary
	ButtonVariantSuccess
	ButtonVariantWarning
	ButtonVariantDanger
	ButtonVariantGhost
)

// Button is an interactive button widget supporting mouse and keyboard activation.
type Button struct {
	id       string
	label    string
	icon     string
	hotkey   rune
	variant  ButtonVariant
	state    ButtonState
	disabled bool
	focused  bool
	bounds   buffer.Rect
	onClick  func()

	styleNormal   style.Style
	styleHover    style.Style
	stylePressed  style.Style
	styleDisabled style.Style
}

// NewButton creates a new interactive button widget.
func NewButton(id, label string, onClick func()) *Button {
	b := &Button{
		id:      id,
		label:   label,
		variant: ButtonVariantDefault,
		state:   ButtonNormal,
		onClick: onClick,
	}
	b.applyVariantStyles()
	return b
}

// ID returns the button identifier.
func (b *Button) ID() string {
	return b.id
}

// Label returns the button label.
func (b *Button) Label() string {
	return b.label
}

// SetLabel updates the button label text.
func (b *Button) SetLabel(label string) *Button {
	b.label = label
	return b
}

// SetIcon sets an optional icon prefix (e.g. "✔ ", "✖ ", "⚙ ").
func (b *Button) SetIcon(icon string) *Button {
	b.icon = icon
	return b
}

// SetHotkey assigns a keyboard shortcut rune to trigger this button.
func (b *Button) SetHotkey(r rune) *Button {
	b.hotkey = r
	return b
}

// SetDisabled sets whether the button is disabled.
func (b *Button) SetDisabled(disabled bool) *Button {
	b.disabled = disabled
	if disabled {
		b.state = ButtonDisabled
	} else if b.state == ButtonDisabled {
		b.state = ButtonNormal
	}
	return b
}

// IsDisabled returns true if the button is disabled.
func (b *Button) IsDisabled() bool {
	return b.disabled
}

// SetFocused sets focus state.
func (b *Button) SetFocused(focused bool) *Button {
	b.focused = focused
	return b
}

// IsFocused returns true if the button is focused.
func (b *Button) IsFocused() bool {
	return b.focused
}

// SetOnClick configures the click callback.
func (b *Button) SetOnClick(fn func()) *Button {
	b.onClick = fn
	return b
}

// SetVariant updates the button's semantic color variant.
func (b *Button) SetVariant(v ButtonVariant) *Button {
	b.variant = v
	b.applyVariantStyles()
	return b
}

func (b *Button) applyVariantStyles() {
	var fg, bg cell.Color
	switch b.variant {
	case ButtonVariantPrimary:
		fg = cell.ColorHex("#000000")
		bg = cell.ColorHex("#00D2FF")
	case ButtonVariantSuccess:
		fg = cell.ColorHex("#000000")
		bg = cell.ColorHex("#00FFAA")
	case ButtonVariantWarning:
		fg = cell.ColorHex("#000000")
		bg = cell.ColorHex("#FFB86C")
	case ButtonVariantDanger:
		fg = cell.ColorHex("#FFFFFF")
		bg = cell.ColorHex("#FF5555")
	case ButtonVariantGhost:
		fg = cell.ColorHex("#8888AA")
		bg = cell.DefaultColor()
	case ButtonVariantDefault:
		fallthrough
	default:
		fg = cell.ColorHex("#FFFFFF")
		bg = cell.Color256(238)
	}

	b.styleNormal = style.NewStyle().
		Bold(true).
		Foreground(fg).
		Background(bg).
		AlignCenter()

	b.styleHover = style.NewStyle().
		Bold(true).
		Foreground(cell.ColorHex("#FFFFFF")).
		Background(cell.ColorHex("#555577")).
		AlignCenter()

	b.stylePressed = style.NewStyle().
		Bold(true).
		Reverse(true).
		Foreground(fg).
		Background(bg).
		AlignCenter()

	b.styleDisabled = style.NewStyle().
		Foreground(cell.ColorHex("#666677")).
		Background(cell.Color256(235)).
		AlignCenter()
}

// SetStyles allows custom style overrides.
func (b *Button) SetStyles(normal, hover, pressed, disabled style.Style) *Button {
	b.styleNormal = normal
	b.styleHover = hover
	b.stylePressed = pressed
	b.styleDisabled = disabled
	return b
}

// Click programmatically triggers the button's click handler if not disabled.
func (b *Button) Click() *Button {
	if !b.disabled && b.onClick != nil {
		b.onClick()
	}
	return b
}

// HandleKey processes keyboard activation (Enter, Space, or hotkey).
func (b *Button) HandleKey(key input.Key) bool {
	if b.disabled {
		return false
	}
	if b.focused && (key.Type == input.KeyEnter || key.Type == input.KeySpace) {
		b.Click()
		return true
	}
	if b.hotkey != 0 && key.Type == input.KeyRune && !key.HasCtrl() && !key.HasAlt() {
		if unicode.ToLower(key.Rune) == unicode.ToLower(b.hotkey) {
			b.Click()
			return true
		}
	}
	return false
}

// HandleMouse processes mouse events (press, release, motion) for hover and click states.
func (b *Button) HandleMouse(msg tea.MouseMsg) bool {
	if b.disabled || b.bounds.IsEmpty() {
		return false
	}

	inside := b.bounds.Contains(msg.X, msg.Y)

	switch msg.Action {
	case input.MouseMotion:
		if inside {
			if b.state != ButtonPressed {
				b.state = ButtonHover
			}
			return true
		} else if b.state == ButtonHover {
			b.state = ButtonNormal
		}
	case input.MousePress:
		if inside && msg.Button == input.MouseLeft {
			b.state = ButtonPressed
			b.focused = true
			return true
		}
	case input.MouseRelease:
		if b.state == ButtonPressed {
			if inside && (msg.Button == input.MouseLeft || msg.Button == input.MouseNone) {
				b.state = ButtonNormal
				b.Click()
				return true
			}
			b.state = ButtonNormal
		}
	}
	return false
}

// Draw renders the button into the provided area.
func (b *Button) Draw(buf *buffer.Buffer, area buffer.Rect) {
	if area.IsEmpty() {
		return
	}
	b.bounds = area

	st := b.styleNormal
	if b.disabled {
		st = b.styleDisabled
	} else if b.state == ButtonPressed {
		st = b.stylePressed
	} else if b.state == ButtonHover || b.focused {
		st = b.styleHover
	}

	text := b.icon + b.label
	if b.hotkey != 0 {
		text += " (" + string(b.hotkey) + ")"
	}
	text = "[ " + text + " ]"

	st.Draw(buf, area, text)
}
