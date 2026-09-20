package widgets

import (
	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/layout"
	"github.com/baibeicha/goatui/pkg/style"
)

// Divider draws a styled horizontal or vertical separator line with an optional title.
type Divider struct {
	orientation layout.Direction
	title       string
	titleAlign  buffer.Alignment
	lineRune    rune
	styleLine   style.Style
	styleTitle  style.Style
}

// NewHorizontalDivider creates a horizontal divider line with an optional title.
func NewHorizontalDivider(title string) *Divider {
	return &Divider{
		orientation: layout.Horizontal,
		title:       title,
		titleAlign:  buffer.AlignCenter,
		lineRune:    '─',
		styleLine: style.NewStyle().
			Foreground(cell.ColorHex("#444455")),
		styleTitle: style.NewStyle().
			Bold(true).
			Foreground(cell.ColorHex("#00FFAA")),
	}
}

// NewVerticalDivider creates a vertical divider line.
func NewVerticalDivider() *Divider {
	return &Divider{
		orientation: layout.Vertical,
		lineRune:    '│',
		styleLine: style.NewStyle().
			Foreground(cell.ColorHex("#444455")),
		styleTitle: style.NewStyle().
			Bold(true).
			Foreground(cell.ColorHex("#00FFAA")),
	}
}

// SetOrientation sets horizontal or vertical direction.
func (d *Divider) SetOrientation(dir layout.Direction) *Divider {
	d.orientation = dir
	if dir == layout.Vertical && d.lineRune == '─' {
		d.lineRune = '│'
	} else if dir == layout.Horizontal && d.lineRune == '│' {
		d.lineRune = '─'
	}
	return d
}

// SetTitle sets the title text.
func (d *Divider) SetTitle(title string) *Divider {
	d.title = title
	return d
}

// SetTitleAlign sets the alignment for the title text.
func (d *Divider) SetTitleAlign(align buffer.Alignment) *Divider {
	d.titleAlign = align
	return d
}

// SetRune configures the separator character.
func (d *Divider) SetRune(r rune) *Divider {
	d.lineRune = r
	return d
}

// SetLineStyle sets the style for the separator line.
func (d *Divider) SetLineStyle(s style.Style) *Divider {
	d.styleLine = s
	return d
}

// SetTitleStyle sets the style for the title text.
func (d *Divider) SetTitleStyle(s style.Style) *Divider {
	d.styleTitle = s
	return d
}

// Draw renders the divider into area.
func (d *Divider) Draw(buf *buffer.Buffer, area buffer.Rect) {
	if area.IsEmpty() {
		return
	}

	lineFg := d.styleLine.GetFg()
	lineBg := d.styleLine.GetBg()
	lineAttrs := d.styleLine.GetModifier()

	if d.orientation == layout.Vertical {
		x := area.X
		for y := area.Y; y < area.Bottom(); y++ {
			buf.SetRune(x, y, d.lineRune, lineFg, lineBg, lineAttrs)
		}
		return
	}

	// Horizontal divider
	y := area.Y
	if d.title == "" {
		for x := area.X; x < area.Right(); x++ {
			buf.SetRune(x, y, d.lineRune, lineFg, lineBg, lineAttrs)
		}
		return
	}

	titleText := " " + d.title + " "
	titleW := buffer.StringWidth(titleText)

	if titleW >= area.Width {
		d.styleTitle.Draw(buf, buffer.NewRect(area.X, y, area.Width, 1), d.title)
		return
	}

	var titleX int
	switch d.titleAlign {
	case buffer.AlignLeft:
		titleX = area.X + 2
	case buffer.AlignRight:
		titleX = area.Right() - titleW - 2
	case buffer.AlignCenter:
		fallthrough
	default:
		titleX = area.X + (area.Width-titleW)/2
	}

	if titleX < area.X {
		titleX = area.X
	}
	if titleX+titleW > area.Right() {
		titleX = max(area.X, area.Right()-titleW)
	}

	titleRight := titleX + titleW
	if titleRight > area.Right() {
		titleRight = area.Right()
	}

	// Draw left line segment
	for x := area.X; x < titleX; x++ {
		buf.SetRune(x, y, d.lineRune, lineFg, lineBg, lineAttrs)
	}

	// Draw title
	titleFg := d.styleTitle.GetFg()
	titleBg := d.styleTitle.GetBg()
	titleAttrs := d.styleTitle.GetModifier()
	buf.SetStringAligned(buffer.NewRect(titleX, y, max(0, titleRight-titleX), 1), titleText, buffer.AlignLeft, titleFg, titleBg, titleAttrs)

	// Draw right line segment
	for x := titleRight; x < area.Right(); x++ {
		buf.SetRune(x, y, d.lineRune, lineFg, lineBg, lineAttrs)
	}
}
