package widgets

import (
	"unicode"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/driver/input"
	"github.com/baibeicha/goatui/pkg/style"
	"github.com/baibeicha/goatui/pkg/tea"
	"github.com/baibeicha/goatui/pkg/validation"
)

// EchoMode defines how characters are displayed in the text input.
type EchoMode uint8

const (
	// EchoNormal displays typed characters in plaintext.
	EchoNormal EchoMode = iota
	// EchoPassword displays masked characters (e.g. • or *).
	EchoPassword
	// EchoNone hides characters completely (e.g. Unix sudo password prompt).
	EchoNone
)

// TextInput is an interactive single-line text entry field.
type TextInput struct {
	prompt           string
	placeholder      string
	value            []rune
	cursor           int
	scrollOffset     int
	focused          bool
	echoMode         EchoMode
	maskRune         rune
	promptStyle      style.Style
	textStyle        style.Style
	validator        *validation.Validator
	validationRes    *validation.Result
	validateOnChange bool
	validateOnBlur   bool
	showError        bool
	errorStyle       style.Style
	customError      string
	onValidate       func(res validation.Result)
}

// NewTextInput creates a new text input field.
func NewTextInput() *TextInput {
	return &TextInput{
		prompt:      "> ",
		value:       make([]rune, 0, 64),
		focused:     true,
		echoMode:    EchoNormal,
		maskRune:    '•',
		promptStyle: style.NewStyle().Bold(true).Foreground(cell.ColorHex("#7D56F4")),
		textStyle:   style.NewStyle(),
		showError:   true,
		errorStyle: style.NewStyle().
			Bold(true).
			Foreground(cell.ColorHex("#FF5555")),
	}
}

// SetPrompt configures the leading prompt text.
func (ti *TextInput) SetPrompt(p string) *TextInput {
	ti.prompt = p
	return ti
}

// SetPlaceholder sets the ghost text when value is empty.
func (ti *TextInput) SetPlaceholder(ph string) *TextInput {
	ti.placeholder = ph
	return ti
}

// SetPromptStyle sets the style for the leading prompt.
func (ti *TextInput) SetPromptStyle(s style.Style) *TextInput {
	ti.promptStyle = s
	return ti
}

// SetTextStyle sets the style for the entered text.
func (ti *TextInput) SetTextStyle(s style.Style) *TextInput {
	ti.textStyle = s
	return ti
}

// SetEchoMode sets the echo display mode (EchoNormal, EchoPassword, EchoNone).
func (ti *TextInput) SetEchoMode(mode EchoMode) *TextInput {
	ti.echoMode = mode
	return ti
}

// SetMaskRune configures the rune used when EchoMode is EchoPassword (defaults to '•').
func (ti *TextInput) SetMaskRune(r rune) *TextInput {
	ti.maskRune = r
	return ti
}

// SetPasswordMode is a convenience method to toggle password masking on or off.
func (ti *TextInput) SetPasswordMode(enable bool) *TextInput {
	if enable {
		ti.echoMode = EchoPassword
		if ti.maskRune == 0 {
			ti.maskRune = '•'
		}
	} else {
		ti.echoMode = EchoNormal
	}
	return ti
}

// EchoMode returns the currently active echo mode.
func (ti *TextInput) EchoMode() EchoMode {
	return ti.echoMode
}

// Value returns the current text string.
func (ti *TextInput) Value() string {
	return string(ti.value)
}

func (ti *TextInput) onContentChanged() {
	ti.validationRes = nil
	if ti.validateOnChange {
		ti.Validate()
	}
}

// SetValue updates the text content and resets cursor to end.
func (ti *TextInput) SetValue(s string) *TextInput {
	ti.value = []rune(s)
	ti.cursor = len(ti.value)
	ti.onContentChanged()
	return ti
}

// Cursor returns the current cursor position in runes.
func (ti *TextInput) Cursor() int {
	return ti.cursor
}

// SetCursor sets the cursor position with bounds checking.
func (ti *TextInput) SetCursor(pos int) *TextInput {
	if pos < 0 {
		pos = 0
	}
	if pos > len(ti.value) {
		pos = len(ti.value)
	}
	ti.cursor = pos
	return ti
}

// Focus enables input capture and cursor rendering.
func (ti *TextInput) Focus() *TextInput {
	ti.focused = true
	return ti
}

// Blur disables input capture.
func (ti *TextInput) Blur() *TextInput {
	ti.focused = false
	if ti.validateOnBlur {
		ti.Validate()
	}
	return ti
}

// Focused returns whether the input is currently focused.
func (ti *TextInput) Focused() bool {
	return ti.focused
}

// SetValidator assigns a field validator to this text input.
func (ti *TextInput) SetValidator(v *validation.Validator) *TextInput {
	ti.validator = v
	return ti
}

// AddValidation appends validation rules to the input's validator.
func (ti *TextInput) AddValidation(rules ...validation.Rule) *TextInput {
	if ti.validator == nil {
		ti.validator = validation.New()
	}
	ti.validator.Add(rules...)
	return ti
}

// SetValidateOnChange enables or disables validation execution whenever the text value changes.
func (ti *TextInput) SetValidateOnChange(enable bool) *TextInput {
	ti.validateOnChange = enable
	return ti
}

// SetValidateOnBlur enables or disables validation execution when focus is lost.
func (ti *TextInput) SetValidateOnBlur(enable bool) *TextInput {
	ti.validateOnBlur = enable
	return ti
}

// SetShowError controls whether validation error messages are displayed visually.
func (ti *TextInput) SetShowError(show bool) *TextInput {
	ti.showError = show
	return ti
}

// ShowError returns whether validation error messages are displayed visually.
func (ti *TextInput) ShowError() bool {
	return ti.showError
}

// SetErrorStyle sets the styling applied to invalid input state and error text.
func (ti *TextInput) SetErrorStyle(s style.Style) *TextInput {
	ti.errorStyle = s
	return ti
}

// SetOnValidate sets a callback triggered upon validation execution.
func (ti *TextInput) SetOnValidate(fn func(res validation.Result)) *TextInput {
	ti.onValidate = fn
	return ti
}

// SetCustomError sets an explicit manual error message, marking the input invalid.
func (ti *TextInput) SetCustomError(errMsg string) *TextInput {
	ti.customError = errMsg
	ti.Validate()
	return ti
}

// ClearError clears any manual custom error and resets validation result.
func (ti *TextInput) ClearError() *TextInput {
	ti.customError = ""
	ti.validationRes = nil
	return ti
}

// Validate executes validation rules against the current value and caches the result.
func (ti *TextInput) Validate() validation.Result {
	var res validation.Result
	if ti.validator != nil {
		res = ti.validator.Validate(ti.Value())
	} else {
		res = validation.OK()
	}
	if ti.customError != "" {
		res.Valid = false
		res.Errors = append([]string{ti.customError}, res.Errors...)
	}
	ti.validationRes = &res
	if ti.onValidate != nil {
		ti.onValidate(res)
	}
	return res
}

// ValidationResult returns the most recent validation result (running Validate if not yet cached).
func (ti *TextInput) ValidationResult() validation.Result {
	if ti.validationRes == nil {
		return ti.Validate()
	}
	return *ti.validationRes
}

// IsValid returns true if the input currently passes all validation rules.
func (ti *TextInput) IsValid() bool {
	return ti.ValidationResult().IsValid()
}

// ErrorMessage returns the first validation error message, or empty string.
func (ti *TextInput) ErrorMessage() string {
	return ti.ValidationResult().Error()
}

func isWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

func (ti *TextInput) moveWordLeft() {
	for ti.cursor > 0 && !isWordRune(ti.value[ti.cursor-1]) {
		ti.cursor--
	}
	for ti.cursor > 0 && isWordRune(ti.value[ti.cursor-1]) {
		ti.cursor--
	}
}

func (ti *TextInput) moveWordRight() {
	for ti.cursor < len(ti.value) && isWordRune(ti.value[ti.cursor]) {
		ti.cursor++
	}
	for ti.cursor < len(ti.value) && !isWordRune(ti.value[ti.cursor]) {
		ti.cursor++
	}
}

func (ti *TextInput) deleteWordLeft() {
	if ti.cursor <= 0 {
		return
	}
	orig := ti.cursor
	ti.moveWordLeft()
	ti.value = append(ti.value[:ti.cursor], ti.value[orig:]...)
}

// HandleKey processes a keyboard event. Returns true if the key modified state.
func (ti *TextInput) HandleKey(k input.Key) bool {
	if !ti.focused {
		return false
	}

	isCtrl := k.HasCtrl()
	isAlt := k.HasAlt()

	// Ctrl/Alt word navigation and deletion shortcuts
	if isCtrl || isAlt {
		switch k.Type {
		case input.KeyLeft:
			if ti.cursor > 0 {
				ti.moveWordLeft()
				return true
			}
		case input.KeyRight:
			if ti.cursor < len(ti.value) {
				ti.moveWordRight()
				return true
			}
		case input.KeyBackspace:
			if ti.cursor > 0 {
				ti.deleteWordLeft()
				ti.onContentChanged()
				return true
			}
		case input.KeyRune:
			switch k.Rune {
			case 'a': // Ctrl+A: Jump to beginning
				ti.cursor = 0
				return true
			case 'e': // Ctrl+E: Jump to end
				ti.cursor = len(ti.value)
				return true
			case 'u': // Ctrl+U: Clear text before cursor
				if ti.cursor > 0 {
					ti.value = ti.value[ti.cursor:]
					ti.cursor = 0
					ti.onContentChanged()
					return true
				}
			case 'k': // Ctrl+K: Clear text after cursor
				if ti.cursor < len(ti.value) {
					ti.value = ti.value[:ti.cursor]
					ti.onContentChanged()
					return true
				}
			case 'w': // Ctrl+W: Delete previous word
				if ti.cursor > 0 {
					ti.deleteWordLeft()
					ti.onContentChanged()
					return true
				}
			}
		}
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
			ti.onContentChanged()
			return true
		}
	case input.KeyDelete:
		if ti.cursor < len(ti.value) {
			ti.value = append(ti.value[:ti.cursor], ti.value[ti.cursor+1:]...)
			ti.onContentChanged()
			return true
		}
	case input.KeySpace:
		ti.insertRune(' ')
		ti.onContentChanged()
		return true
	case input.KeyRune:
		if k.Rune != 0 && !isCtrl && !isAlt {
			ti.insertRune(k.Rune)
			ti.onContentChanged()
			return true
		}
	}

	return false
}

func (ti *TextInput) insertRune(r rune) {
	ti.value = append(ti.value[:ti.cursor], append([]rune{r}, ti.value[ti.cursor:]...)...)
	ti.cursor++
}

// InsertString inserts text into the input at current cursor position.
// Strips newlines and carriage returns to ensure single-line field integrity.
func (ti *TextInput) InsertString(s string) *TextInput {
	if s == "" {
		return ti
	}
	runes := make([]rune, 0, len(s))
	for _, r := range s {
		if r != '\r' && r != '\n' {
			runes = append(runes, r)
		}
	}
	if len(runes) == 0 {
		return ti
	}
	ti.value = append(ti.value[:ti.cursor], append(runes, ti.value[ti.cursor:]...)...)
	ti.cursor += len(runes)
	ti.onContentChanged()
	return ti
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
		pFg := ti.promptStyle.GetFg()
		if pFg.IsDefault() {
			pFg = cell.ColorHex("#7D56F4")
		}
		if ti.showError && !ti.IsValid() {
			errFg := ti.errorStyle.GetFg()
			if !errFg.IsDefault() {
				pFg = errFg
			}
		}
		pBg := ti.promptStyle.GetBg()
		pMod := ti.promptStyle.GetModifier()
		if pMod == cell.AttrNone {
			pMod = cell.AttrBold
		}
		curX = buf.SetString(curX, area.Y, ti.prompt, pFg, pBg, pMod)
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

	// Prepare displayed runes based on EchoMode
	var displayRunes []rune
	switch ti.echoMode {
	case EchoPassword:
		mask := ti.maskRune
		if mask == 0 {
			mask = '•'
		}
		displayRunes = make([]rune, len(ti.value))
		for i := range displayRunes {
			displayRunes[i] = mask
		}
	case EchoNone:
		displayRunes = nil
	default: // EchoNormal
		displayRunes = ti.value
	}

	// Adjust scrollOffset so the cursor is always visible within availWidth
	if ti.cursor < ti.scrollOffset {
		ti.scrollOffset = ti.cursor
	}

	for {
		cursorVisualCol := 0
		for i := ti.scrollOffset; i < ti.cursor && i < len(displayRunes); i++ {
			w := buffer.RuneWidth(displayRunes[i])
			if w == 0 {
				w = 1
			}
			cursorVisualCol += w
		}
		cursorW := 1
		if ti.cursor < len(displayRunes) {
			cursorW = buffer.RuneWidth(displayRunes[ti.cursor])
			if cursorW == 0 {
				cursorW = 1
			}
		}
		if cursorVisualCol+cursorW > availWidth && ti.scrollOffset < ti.cursor {
			ti.scrollOffset++
		} else {
			break
		}
	}

	// Avoid leaving empty space when text fits completely
	totalW := 0
	for i := ti.scrollOffset; i < len(displayRunes); i++ {
		w := buffer.RuneWidth(displayRunes[i])
		if w == 0 {
			w = 1
		}
		totalW += w
	}
	for ti.scrollOffset > 0 && totalW < availWidth {
		prevW := buffer.RuneWidth(displayRunes[ti.scrollOffset-1])
		if prevW == 0 {
			prevW = 1
		}
		if totalW+prevW <= availWidth {
			ti.scrollOffset--
			totalW += prevW
		} else {
			break
		}
	}

	// Calculate visible runes strictly constrained by availWidth
	var visibleRunes []rune
	curW := 0
	for i := ti.scrollOffset; i < len(displayRunes); i++ {
		rw := buffer.RuneWidth(displayRunes[i])
		if rw == 0 {
			rw = 1
		}
		if curW+rw > availWidth {
			break
		}
		curW += rw
		visibleRunes = append(visibleRunes, displayRunes[i])
	}
	strToDraw := string(visibleRunes)

	textFg := ti.textStyle.GetFg()
	textBg := ti.textStyle.GetBg()
	textMod := ti.textStyle.GetModifier()

	if len(ti.value) == 0 && ti.placeholder != "" && !ti.focused {
		phArea := buffer.NewRect(curX, area.Y, availWidth, 1)
		buf.SetStringAligned(phArea, ti.placeholder, buffer.AlignLeft, cell.ColorHex("#626262"), textBg, cell.AttrItalic)
	} else if strToDraw != "" {
		buf.SetString(curX, area.Y, strToDraw, textFg, textBg, textMod)
	}

	// Draw cursor if focused
	if ti.focused {
		cursorVisualX := curX
		for i := ti.scrollOffset; i < ti.cursor && i < len(displayRunes); i++ {
			w := buffer.RuneWidth(displayRunes[i])
			if w == 0 {
				w = 1
			}
			cursorVisualX += w
		}

		if cursorVisualX < area.Right() {
			cursorRune := ' '
			if ti.cursor < len(displayRunes) {
				cursorRune = displayRunes[ti.cursor]
			}
			buf.SetRune(cursorVisualX, area.Y, cursorRune, textFg, textBg, cell.AttrReverse)
		}
	}

	// Draw validation error message if invalid and showError is enabled
	if ti.showError && !ti.IsValid() {
		errMsg := ti.ErrorMessage()
		if errMsg != "" {
			errFg := ti.errorStyle.GetFg()
			if errFg.IsDefault() {
				errFg = cell.ColorHex("#FF5555")
			}
			errBg := ti.errorStyle.GetBg()

			if area.Height >= 2 {
				errRect := buffer.NewRect(area.X, area.Y+1, area.Width, 1)
				buf.SetStringAligned(errRect, "✖ "+errMsg, buffer.AlignLeft, errFg, errBg, cell.AttrBold)
			} else {
				textEndCol := curX + totalW + 2
				if textEndCol+4 < area.Right() {
					remWidth := area.Right() - textEndCol
					errRect := buffer.NewRect(textEndCol, area.Y, remWidth, 1)
					buf.SetStringAligned(errRect, "✖ "+errMsg, buffer.AlignLeft, errFg, errBg, cell.AttrBold)
				}
			}
		}
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
		if ti.echoMode == EchoNone {
			return true
		}
		promptLen := buffer.StringWidth(ti.prompt)
		relX := msg.X - area.X - promptLen
		if relX < 0 {
			relX = 0
		}
		runes := ti.value
		if ti.echoMode == EchoPassword {
			mask := ti.maskRune
			if mask == 0 {
				mask = '•'
			}
			runes = make([]rune, len(ti.value))
			for i := range runes {
				runes[i] = mask
			}
		}
		newCursor := ti.scrollOffset
		currX := 0
		for i := ti.scrollOffset; i < len(runes); i++ {
			w := buffer.RuneWidth(runes[i])
			if w == 0 {
				w = 1
			}
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
