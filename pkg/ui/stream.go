package ui

import (
	"fmt"
	"io"
	"math"
	"os"
	"strings"
	"sync"

	"github.com/baibeicha/goatui/pkg/widgets"
)

// RenderStream provides a streaming console renderer for non-fullscreen CLI utilities,
// progress bars, animated spinners, and rich text output.
type RenderStream struct {
	mu     sync.Mutex
	writer io.Writer
	width  int
}

// NewRenderStream initializes a new streaming CLI renderer writing to w (defaults to os.Stdout if nil).
func NewRenderStream(w io.Writer) *RenderStream {
	if w == nil {
		w = os.Stdout
	}
	return &RenderStream{
		writer: w,
		width:  80,
	}
}

// SetWidth overrides the assumed terminal width in columns.
func (rs *RenderStream) SetWidth(w int) *RenderStream {
	if w > 0 {
		rs.width = w
	}
	return rs
}

// Println prints plain text followed by a newline.
func (rs *RenderStream) Println(a ...any) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	fmt.Fprintln(rs.writer, a...)
}

// Printf formats and prints text.
func (rs *RenderStream) Printf(format string, a ...any) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	fmt.Fprintf(rs.writer, format, a...)
}

// PrintLine prints a rich-text Line with full ANSI coloring.
func (rs *RenderStream) PrintLine(line Line) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	fmt.Fprintln(rs.writer, line.ANSI())
}

// ClearLine erases the current terminal line.
func (rs *RenderStream) ClearLine() {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	fmt.Fprint(rs.writer, "\r\x1b[2K")
}

// Progress updates an inline progress bar on the current line.
func (rs *RenderStream) Progress(label string, ratio float64) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	ratio = math.Max(0.0, math.Min(1.0, ratio))
	pct := int(ratio * 100.0)

	barW := 24
	filled := int(math.Round(ratio * float64(barW)))
	if filled < 0 {
		filled = 0
	} else if filled > barW {
		filled = barW
	}
	empty := barW - filled
	if empty < 0 {
		empty = 0
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)

	fmt.Fprintf(rs.writer, "\r\x1b[2K%-20s [%s] %3d%%", label, bar, pct)
}

// Spinner updates an animated spinner glyph and message on the current line.
func (rs *RenderStream) Spinner(msg string, sp *widgets.Spinner) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	glyph := "⠋"
	if sp != nil {
		sp.Tick()
		glyph = sp.CurrentGlyph()
	}
	fmt.Fprintf(rs.writer, "\r\x1b[2K\x1b[1;36m%s\x1b[0m %s", glyph, msg)
}

// Done marks a task complete, printing a green checkmark and advancing to the next line.
func (rs *RenderStream) Done(msg string) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	fmt.Fprintf(rs.writer, "\r\x1b[2K\x1b[1;32m✔\x1b[0m %s\n", msg)
}

// Fail marks a task failed, printing a red cross and advancing to the next line.
func (rs *RenderStream) Fail(msg string) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	fmt.Fprintf(rs.writer, "\r\x1b[2K\x1b[1;31m✖\x1b[0m %s\n", msg)
}
