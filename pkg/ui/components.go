package ui

import (
	"fmt"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/driver/input"
	"github.com/baibeicha/goatui/pkg/style"
	"github.com/baibeicha/goatui/pkg/tea"
	"github.com/baibeicha/goatui/pkg/validation"
	"github.com/baibeicha/goatui/pkg/widgets"
)

// Card represents a structured container with border, title, optional subtitle, content, and footer.
type Card struct {
	title       string
	subtitle    string
	border      style.Border
	content     View
	footer      View
	borderStyle style.Style
	headerStyle style.Style
}

// NewCard creates a new Card component.
func NewCard(title string, content View) *Card {
	return &Card{
		title:    title,
		border:   style.BorderRounded,
		content:  content,
		borderStyle: style.NewStyle().
			Border(style.BorderRounded).
			BorderForeground(cell.ColorHex("#444466")).
			Background(cell.Color256(234)),
		headerStyle: style.NewStyle().
			Bold(true).
			Foreground(cell.ColorHex("#00D2FF")),
	}
}

// SetSubtitle sets an optional subtitle for the card.
func (c *Card) SetSubtitle(sub string) *Card {
	c.subtitle = sub
	return c
}

// SetBorder sets the border style (e.g. style.BorderRounded, style.BorderThick).
func (c *Card) SetBorder(b style.Border) *Card {
	c.border = b
	c.borderStyle = c.borderStyle.Border(b)
	return c
}

// SetBorderColor sets the color of the card border.
func (c *Card) SetBorderColor(fg cell.Color) *Card {
	c.borderStyle = c.borderStyle.BorderForeground(fg)
	return c
}

// SetHeaderColor sets the title foreground color.
func (c *Card) SetHeaderColor(fg cell.Color) *Card {
	c.headerStyle = c.headerStyle.Foreground(fg)
	return c
}

// SetFooter sets an optional footer view.
func (c *Card) SetFooter(footer View) *Card {
	c.footer = footer
	return c
}

// Draw renders the card box and its children.
func (c *Card) Draw(buf *buffer.Buffer, area buffer.Rect) {
	if area.IsEmpty() {
		return
	}

	titleText := ""
	if c.title != "" {
		titleText = " " + c.title
		if c.subtitle != "" {
			titleText += " ─ " + c.subtitle
		}
		titleText += " "
	}

	st := c.borderStyle.Title("")
	st.Draw(buf, area, "")

	if titleText != "" && area.Width >= 6 {
		titleArea := buffer.NewRect(area.X+2, area.Y, max(0, area.Width-4), 1)
		c.headerStyle.Draw(buf, titleArea, titleText)
	}

	inner := area.Inset(1, 1)
	if inner.IsEmpty() {
		return
	}

	contentArea := inner
	if c.footer != nil && inner.Height >= 3 {
		contentArea = buffer.NewRect(inner.X, inner.Y, inner.Width, inner.Height-1)
		footerArea := buffer.NewRect(inner.X, inner.Bottom()-1, inner.Width, 1)
		c.footer.Draw(buf, footerArea)
	}

	if c.content != nil && !contentArea.IsEmpty() {
		c.content.Draw(buf, contentArea)
	}
}

// StatCard displays a single prominent KPI/statistic with a label, value, and trend.
type StatCard struct {
	title       string
	value       string
	trend       string
	trendUp     bool
	accentColor cell.Color
	borderStyle style.Style
}

// NewStatCard creates a KPI telemetry card.
func NewStatCard(title, value, trend string) *StatCard {
	return &StatCard{
		title:       title,
		value:       value,
		trend:       trend,
		trendUp:     true,
		accentColor: cell.ColorHex("#00FFAA"),
		borderStyle: style.NewStyle().
			Border(style.BorderRounded).
			BorderForeground(cell.ColorHex("#444455")).
			Background(cell.Color256(234)),
	}
}

// SetUp toggles the trend direction (true = positive/green, false = negative/red).
func (s *StatCard) SetUp(up bool) *StatCard {
	s.trendUp = up
	return s
}

// SetAccent sets the value's highlight color.
func (s *StatCard) SetAccent(color cell.Color) *StatCard {
	s.accentColor = color
	return s
}

// SetBorderColor sets the border color.
func (s *StatCard) SetBorderColor(color cell.Color) *StatCard {
	s.borderStyle = s.borderStyle.BorderForeground(color)
	return s
}

// Draw renders the KPI stat card.
func (s *StatCard) Draw(buf *buffer.Buffer, area buffer.Rect) {
	if area.IsEmpty() {
		return
	}

	s.borderStyle.Draw(buf, area, "")
	inner := area.Inset(1, 1)
	if inner.IsEmpty() {
		return
	}

	// 1. Title at top
	buf.SetStringAligned(buffer.NewRect(inner.X, inner.Y, inner.Width, 1), s.title, buffer.AlignLeft, cell.ColorHex("#8888AA"), cell.DefaultColor(), cell.AttrNone)

	// 2. Large Value in middle
	valY := inner.Y + 1
	if inner.Height >= 2 {
		buf.SetStringAligned(buffer.NewRect(inner.X, valY, inner.Width, 1), s.value, buffer.AlignLeft, s.accentColor, cell.DefaultColor(), cell.AttrBold)
	}

	// 3. Trend at bottom
	if s.trend != "" && inner.Height >= 3 {
		trendColor := cell.ColorHex("#00FFAA")
		arrow := "▲ "
		if !s.trendUp {
			trendColor = cell.ColorHex("#FF5555")
			arrow = "▼ "
		}
		buf.SetStringAligned(buffer.NewRect(inner.X, inner.Y+2, inner.Width, 1), arrow+s.trend, buffer.AlignLeft, trendColor, cell.DefaultColor(), cell.AttrNone)
	}
}

// Badge renders a styled status tag or pill.
type Badge struct {
	text string
	fg   cell.Color
	bg   cell.Color
}

// NewBadge creates a new Badge component.
func NewBadge(text string, fg, bg cell.Color) *Badge {
	return &Badge{text: text, fg: fg, bg: bg}
}

// Draw renders the badge into area.
func (b *Badge) Draw(buf *buffer.Buffer, area buffer.Rect) {
	if area.IsEmpty() {
		return
	}
	text := " " + b.text + " "
	st := style.NewStyle().
		Bold(true).
		Foreground(b.fg).
		Background(b.bg).
		AlignCenter()
	st.Draw(buf, area, text)
}

// KeyHint represents a single keyboard shortcut mapping.
type KeyHint struct {
	Key  string
	Desc string
}

// KeyHints renders a horizontal row of keyboard shortcuts.
type KeyHints struct {
	hints    []KeyHint
	keyFg    cell.Color
	keyBg    cell.Color
	descFg   cell.Color
	spacing  int
}

// NewKeyHints creates a key hints toolbar.
func NewKeyHints(hints ...KeyHint) *KeyHints {
	return &KeyHints{
		hints:   hints,
		keyFg:   cell.ColorHex("#000000"),
		keyBg:   cell.ColorHex("#00D2FF"),
		descFg:  cell.ColorHex("#8888AA"),
		spacing: 2,
	}
}

// Add adds a shortcut to the toolbar.
func (kh *KeyHints) Add(key, desc string) *KeyHints {
	kh.hints = append(kh.hints, KeyHint{Key: key, Desc: desc})
	return kh
}

// Hints returns the slice of registered key hints.
func (kh *KeyHints) Hints() []KeyHint {
	if kh == nil {
		return nil
	}
	return kh.hints
}

// Draw renders the key hints.
func (kh *KeyHints) Draw(buf *buffer.Buffer, area buffer.Rect) {
	if area.IsEmpty() || len(kh.hints) == 0 {
		return
	}

	currX := area.X
	for _, h := range kh.hints {
		if currX >= area.Right() {
			break
		}

		keyText := fmt.Sprintf("[%s]", h.Key)
		keyW := buffer.StringWidth(keyText)
		if currX+keyW > area.Right() {
			break
		}

		buf.SetStringAligned(buffer.NewRect(currX, area.Y, keyW, 1), keyText, buffer.AlignLeft, kh.keyFg, kh.keyBg, cell.AttrBold)
		currX += keyW + 1

		descText := h.Desc
		descW := buffer.StringWidth(descText)
		if currX >= area.Right() {
			break
		}
		availDesc := min(descW, area.Right()-currX)
		if availDesc > 0 {
			buf.SetStringAligned(buffer.NewRect(currX, area.Y, availDesc, 1), descText, buffer.AlignLeft, kh.descFg, cell.DefaultColor(), cell.AttrNone)
			currX += availDesc + kh.spacing
		}
	}
}

// DividerView adapts widgets.Divider into the ui.View interface.
type DividerView struct {
	*widgets.Divider
}

// NewDivider creates a horizontal divider view.
func NewDivider(title string) *DividerView {
	return &DividerView{Divider: widgets.NewHorizontalDivider(title)}
}

// ButtonView adapts widgets.Button into the ui.View interface.
type ButtonView struct {
	*widgets.Button
}

// NewButton creates a declarative button view.
func NewButton(id, label string, onClick func()) *ButtonView {
	return &ButtonView{Button: widgets.NewButton(id, label, onClick)}
}

// FormField combines a label, an interactive TextInput, required indicator, and validation error/helper display.
type FormField struct {
	label       string
	input       *widgets.TextInput
	required    bool
	helperText  string
	labelStyle  style.Style
	helperStyle style.Style
	errorStyle  style.Style
}

// NewFormField creates a declarative form field wrapping a TextInput.
func NewFormField(label string, input *widgets.TextInput) *FormField {
	if input == nil {
		input = widgets.NewTextInput()
	}
	return &FormField{
		label:       label,
		input:       input,
		labelStyle:  style.NewStyle().Bold(true).Foreground(cell.ColorHex("#E0E0FF")),
		helperStyle: style.NewStyle().Foreground(cell.ColorHex("#888888")),
		errorStyle:  style.NewStyle().Foreground(cell.ColorHex("#FF5555")).Bold(true),
	}
}

// SetRequired marks the field as required and visually appends an asterisk.
// If req is true, also registers a validation.Required rule on the underlying input.
func (ff *FormField) SetRequired(req bool) *FormField {
	ff.required = req
	if req && ff.input != nil {
		ff.input.AddValidation(validation.Required())
	}
	return ff
}

// Required returns whether the field is marked as required.
func (ff *FormField) Required() bool {
	return ff.required
}

// SetLabel sets the label text.
func (ff *FormField) SetLabel(label string) *FormField {
	ff.label = label
	return ff
}

// Label returns the label text.
func (ff *FormField) Label() string {
	return ff.label
}

// SetHelperText sets a secondary helper/hint message shown when valid.
func (ff *FormField) SetHelperText(text string) *FormField {
	ff.helperText = text
	return ff
}

// HelperText returns the helper text.
func (ff *FormField) HelperText() string {
	return ff.helperText
}

// SetLabelStyle configures the styling for the field's label.
func (ff *FormField) SetLabelStyle(s style.Style) *FormField {
	ff.labelStyle = s
	return ff
}

// SetHelperStyle configures the styling for the helper text.
func (ff *FormField) SetHelperStyle(s style.Style) *FormField {
	ff.helperStyle = s
	return ff
}

// SetErrorStyle configures the styling for error text and indicators.
func (ff *FormField) SetErrorStyle(s style.Style) *FormField {
	ff.errorStyle = s
	if ff.input != nil {
		ff.input.SetErrorStyle(s)
	}
	return ff
}

// Input returns the underlying TextInput widget.
func (ff *FormField) Input() *widgets.TextInput {
	return ff.input
}

// Value returns the current text value of the input.
func (ff *FormField) Value() string {
	if ff.input == nil {
		return ""
	}
	return ff.input.Value()
}

// SetValue sets the text content of the input.
func (ff *FormField) SetValue(val string) *FormField {
	if ff.input != nil {
		ff.input.SetValue(val)
	}
	return ff
}

// AddValidation adds validation rules to the underlying input.
func (ff *FormField) AddValidation(rules ...validation.Rule) *FormField {
	if ff.input != nil {
		ff.input.AddValidation(rules...)
	}
	return ff
}

// Validate triggers validation on the underlying input and returns the result.
func (ff *FormField) Validate() validation.Result {
	if ff.input == nil {
		return validation.OK()
	}
	return ff.input.Validate()
}

// IsValid returns whether the field is valid.
func (ff *FormField) IsValid() bool {
	if ff.input == nil {
		return true
	}
	return ff.input.IsValid()
}

// ErrorMessage returns the current validation error message.
func (ff *FormField) ErrorMessage() string {
	if ff.input == nil {
		return ""
	}
	return ff.input.ErrorMessage()
}

// Focus focuses the underlying text input.
func (ff *FormField) Focus() *FormField {
	if ff.input != nil {
		ff.input.Focus()
	}
	return ff
}

// Blur blurs the underlying text input.
func (ff *FormField) Blur() *FormField {
	if ff.input != nil {
		ff.input.Blur()
	}
	return ff
}

// Focused returns whether the underlying text input is focused.
func (ff *FormField) Focused() bool {
	if ff.input == nil {
		return false
	}
	return ff.input.Focused()
}

// HandleKey forwards keyboard events to the underlying TextInput.
func (ff *FormField) HandleKey(k input.Key) bool {
	if ff.input != nil {
		return ff.input.HandleKey(k)
	}
	return false
}

// HandleMouse forwards mouse events to the underlying TextInput.
func (ff *FormField) HandleMouse(m tea.MouseMsg, area buffer.Rect) bool {
	if ff.input != nil {
		return ff.input.HandleMouse(m, area)
	}
	return false
}

// Draw renders the FormField according to available dimensions.
func (ff *FormField) Draw(buf *buffer.Buffer, area buffer.Rect) {
	if area.IsEmpty() || ff.input == nil {
		return
	}

	lbl := ff.label
	if ff.required {
		lbl += " *"
	}

	// Case 1: Stacked multiline layout (Height >= 3)
	// Line 0: Label
	// Line 1: Input
	// Line 2: Error message or Helper text
	if area.Height >= 3 {
		lblWidth := buffer.StringWidth(lbl)
		lblRect := buffer.NewRect(area.X, area.Y, min(lblWidth, area.Width), 1)
		ff.labelStyle.Draw(buf, lblRect, lbl)
		if ff.required {
			starX := area.X + buffer.StringWidth(ff.label) + 1
			if starX < area.Right() {
				buf.SetRune(starX, area.Y, '*', cell.ColorHex("#FF5555"), cell.DefaultColor(), cell.AttrBold)
			}
		}

		inputRect := buffer.NewRect(area.X, area.Y+1, area.Width, 1)
		prevShowError := ff.input.ShowError()
		ff.input.SetShowError(false)
		ff.input.Draw(buf, inputRect)
		ff.input.SetShowError(prevShowError)

		if !ff.input.IsValid() {
			errMsg := "✖ " + ff.input.ErrorMessage()
			errRect := buffer.NewRect(area.X, area.Y+2, area.Width, 1)
			ff.errorStyle.Draw(buf, errRect, errMsg)
		} else if ff.helperText != "" {
			helpRect := buffer.NewRect(area.X, area.Y+2, area.Width, 1)
			ff.helperStyle.Draw(buf, helpRect, ff.helperText)
		}
		return
	}

	// Case 2: 2-line layout (Height == 2)
	// Line 0: Label
	// Line 1: Input (renders its own inline error if space allows)
	if area.Height == 2 {
		lblWidth := buffer.StringWidth(lbl)
		lblRect := buffer.NewRect(area.X, area.Y, min(lblWidth, area.Width), 1)
		ff.labelStyle.Draw(buf, lblRect, lbl)
		if ff.required {
			starX := area.X + buffer.StringWidth(ff.label) + 1
			if starX < area.Right() {
				buf.SetRune(starX, area.Y, '*', cell.ColorHex("#FF5555"), cell.DefaultColor(), cell.AttrBold)
			}
		}

		inputRect := buffer.NewRect(area.X, area.Y+1, area.Width, 1)
		ff.input.Draw(buf, inputRect)
		return
	}

	// Case 3: Single line inline layout (Height == 1)
	// [Label *] [TextInput...]
	lblWidth := buffer.StringWidth(lbl) + 1 // label + space
	if lblWidth < area.Width {
		lblRect := buffer.NewRect(area.X, area.Y, lblWidth, 1)
		ff.labelStyle.Draw(buf, lblRect, lbl)
		if ff.required {
			starX := area.X + buffer.StringWidth(ff.label) + 1
			if starX < area.Right() {
				buf.SetRune(starX, area.Y, '*', cell.ColorHex("#FF5555"), cell.DefaultColor(), cell.AttrBold)
			}
		}

		inputRect := buffer.NewRect(area.X+lblWidth, area.Y, area.Width-lblWidth, 1)
		ff.input.Draw(buf, inputRect)
	} else {
		ff.input.Draw(buf, area)
	}
}

