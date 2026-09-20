package ui

import (
	"fmt"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/style"
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
