package widgets

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
)

// MarkdownLineKind classifies parsed markdown lines for styled presentation.
type MarkdownLineKind int

const (
	MdText MarkdownLineKind = iota
	MdH1
	MdH2
	MdH3
	MdH4
	MdBlockquote
	MdBullet
	MdNumbered
	MdTaskUnchecked
	MdTaskChecked
	MdCodeFenceStart
	MdCodeLine
	MdCodeFenceEnd
	MdTable
	MdDivider
)

// FormattedMdLine represents a rendered line with visual attributes.
type FormattedMdLine struct {
	Kind   MarkdownLineKind
	Text   string
	Aux    string // Language tag for code fence, or original row
	Indent int
}

// MarkdownStyle specifies colors for markdown preview rendering.
type MarkdownStyle struct {
	Fg       cell.Color
	Bg       cell.Color
	H1Fg     cell.Color
	H2Fg     cell.Color
	BorderFg cell.Color
	CodeFg   cell.Color
	CodeBg   cell.Color
	MutedFg  cell.Color
	StringFg cell.Color
}

// DefaultMarkdownStyle returns sensible default colors for markdown rendering.
func DefaultMarkdownStyle() MarkdownStyle {
	return MarkdownStyle{
		Fg:       cell.RGB(220, 220, 220),
		Bg:       cell.RGB(25, 25, 35),
		H1Fg:     cell.RGB(245, 169, 127),
		H2Fg:     cell.RGB(196, 167, 231),
		BorderFg: cell.RGB(90, 95, 120),
		CodeFg:   cell.RGB(235, 111, 146),
		CodeBg:   cell.RGB(35, 38, 52),
		MutedFg:  cell.RGB(110, 115, 141),
		StringFg: cell.RGB(156, 207, 216),
	}
}

// MarkdownRenderer parses raw markdown text and renders a formatted terminal preview buffer.
type MarkdownRenderer struct {
	Lines []FormattedMdLine
}

// NewMarkdownRenderer parses markdown text into structured preview lines.
func NewMarkdownRenderer(rawText string) *MarkdownRenderer {
	mr := &MarkdownRenderer{}
	mr.Parse(rawText)
	return mr
}

// Parse converts raw markdown lines into FormattedMdLine slices.
func (mr *MarkdownRenderer) Parse(rawText string) {
	mr.Lines = nil
	scanner := bufio.NewScanner(strings.NewReader(rawText))
	inCodeFence := false
	codeLang := ""

	for scanner.Scan() {
		rawLine := scanner.Text()
		trimmed := strings.TrimSpace(rawLine)

		// Code fence toggle: ```lang
		if strings.HasPrefix(trimmed, "```") {
			if !inCodeFence {
				inCodeFence = true
				codeLang = strings.TrimPrefix(trimmed, "```")
				if codeLang == "" {
					codeLang = "code"
				}
				mr.Lines = append(mr.Lines, FormattedMdLine{
					Kind: MdCodeFenceStart,
					Aux:  codeLang,
				})
			} else {
				inCodeFence = false
				mr.Lines = append(mr.Lines, FormattedMdLine{
					Kind: MdCodeFenceEnd,
				})
			}
			continue
		}

		if inCodeFence {
			mr.Lines = append(mr.Lines, FormattedMdLine{
				Kind: MdCodeLine,
				Text: rawLine,
				Aux:  codeLang,
			})
			continue
		}

		if trimmed == "" {
			mr.Lines = append(mr.Lines, FormattedMdLine{
				Kind: MdText,
				Text: "",
			})
			continue
		}

		// Horizontal rule: --- or ***
		if trimmed == "---" || trimmed == "***" || trimmed == "___" {
			mr.Lines = append(mr.Lines, FormattedMdLine{
				Kind: MdDivider,
			})
			continue
		}

		// Headers
		if strings.HasPrefix(rawLine, "# ") {
			mr.Lines = append(mr.Lines, FormattedMdLine{
				Kind: MdH1,
				Text: strings.TrimPrefix(rawLine, "# "),
			})
			continue
		}
		if strings.HasPrefix(rawLine, "## ") {
			mr.Lines = append(mr.Lines, FormattedMdLine{
				Kind: MdH2,
				Text: strings.TrimPrefix(rawLine, "## "),
			})
			continue
		}
		if strings.HasPrefix(rawLine, "### ") {
			mr.Lines = append(mr.Lines, FormattedMdLine{
				Kind: MdH3,
				Text: strings.TrimPrefix(rawLine, "### "),
			})
			continue
		}
		if strings.HasPrefix(rawLine, "#### ") {
			mr.Lines = append(mr.Lines, FormattedMdLine{
				Kind: MdH4,
				Text: strings.TrimPrefix(rawLine, "#### "),
			})
			continue
		}

		// Blockquotes: > quote
		if strings.HasPrefix(rawLine, "> ") {
			mr.Lines = append(mr.Lines, FormattedMdLine{
				Kind: MdBlockquote,
				Text: strings.TrimPrefix(rawLine, "> "),
			})
			continue
		}

		// Checkboxes / Task lists
		if strings.HasPrefix(trimmed, "- [ ] ") || strings.HasPrefix(trimmed, "* [ ] ") {
			mr.Lines = append(mr.Lines, FormattedMdLine{
				Kind: MdTaskUnchecked,
				Text: trimmed[6:],
			})
			continue
		}
		if strings.HasPrefix(trimmed, "- [x] ") || strings.HasPrefix(trimmed, "* [x] ") ||
			strings.HasPrefix(trimmed, "- [X] ") || strings.HasPrefix(trimmed, "* [X] ") {
			mr.Lines = append(mr.Lines, FormattedMdLine{
				Kind: MdTaskChecked,
				Text: trimmed[6:],
			})
			continue
		}

		// Bullet lists: - item, * item, + item
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "+ ") {
			mr.Lines = append(mr.Lines, FormattedMdLine{
				Kind: MdBullet,
				Text: trimmed[2:],
			})
			continue
		}

		// Tables
		if strings.HasPrefix(trimmed, "|") && strings.HasSuffix(trimmed, "|") {
			mr.Lines = append(mr.Lines, FormattedMdLine{
				Kind: MdTable,
				Text: trimmed,
			})
			continue
		}

		// Standard paragraphs
		mr.Lines = append(mr.Lines, FormattedMdLine{
			Kind: MdText,
			Text: rawLine,
		})
	}
}

// Render writes formatted markdown content into a GoatUI buffer.
func (mr *MarkdownRenderer) Render(buf *buffer.Buffer, startX, startY, width, height, scrollY int, st MarkdownStyle) {
	if buf == nil || width < 4 || height < 2 {
		return
	}

	for y := 0; y < height; y++ {
		lineIdx := scrollY + y
		if lineIdx >= len(mr.Lines) {
			break
		}
		screenY := startY + y
		if screenY < 0 || screenY >= buf.Height() {
			continue
		}

		md := mr.Lines[lineIdx]

		switch md.Kind {
		case MdH1:
			prefix := "▌ "
			for i, r := range prefix {
				buf.SetRune(startX+1+i, screenY, r, st.H1Fg, st.Bg, cell.AttrBold)
			}
			for i, r := range md.Text {
				if i+3 < width-2 {
					buf.SetRune(startX+3+i, screenY, r, st.H1Fg, st.Bg, cell.AttrBold)
				}
			}

		case MdH2:
			prefix := "◈ "
			for i, r := range prefix {
				buf.SetRune(startX+1+i, screenY, r, st.H2Fg, st.Bg, cell.AttrBold)
			}
			for i, r := range md.Text {
				if i+3 < width-2 {
					buf.SetRune(startX+3+i, screenY, r, st.H2Fg, st.Bg, cell.AttrBold)
				}
			}

		case MdH3, MdH4:
			prefix := "▸ "
			for i, r := range prefix {
				buf.SetRune(startX+1+i, screenY, r, st.H2Fg, st.Bg, cell.AttrNone)
			}
			for i, r := range md.Text {
				if i+3 < width-2 {
					buf.SetRune(startX+3+i, screenY, r, st.Fg, st.Bg, cell.AttrBold)
				}
			}

		case MdBlockquote:
			buf.SetRune(startX+1, screenY, '│', st.MutedFg, st.Bg, cell.AttrNone)
			for i, r := range md.Text {
				if i+3 < width-2 {
					buf.SetRune(startX+3+i, screenY, r, st.MutedFg, st.Bg, cell.AttrItalic)
				}
			}

		case MdBullet:
			bullet := "• "
			for i, r := range bullet {
				buf.SetRune(startX+2+i, screenY, r, st.H2Fg, st.Bg, cell.AttrNone)
			}
			for i, r := range md.Text {
				if i+4 < width-2 {
					buf.SetRune(startX+4+i, screenY, r, st.Fg, st.Bg, cell.AttrNone)
				}
			}

		case MdTaskUnchecked:
			box := "○ "
			for i, r := range box {
				buf.SetRune(startX+2+i, screenY, r, st.MutedFg, st.Bg, cell.AttrNone)
			}
			for i, r := range md.Text {
				if i+5 < width-2 {
					buf.SetRune(startX+5+i, screenY, r, st.Fg, st.Bg, cell.AttrNone)
				}
			}

		case MdTaskChecked:
			box := "● "
			for i, r := range box {
				buf.SetRune(startX+2+i, screenY, r, st.StringFg, st.Bg, cell.AttrBold)
			}
			for i, r := range md.Text {
				if i+5 < width-2 {
					buf.SetRune(startX+5+i, screenY, r, st.Fg, st.Bg, cell.AttrNone)
				}
			}

		case MdCodeFenceStart:
			tag := fmt.Sprintf("┌──  %s  ", md.Aux)
			for col := 0; col < width-4; col++ {
				r := '─'
				if col < len(tag) {
					r = rune(tag[col])
				}
				buf.SetRune(startX+2+col, screenY, r, st.BorderFg, st.Bg, cell.AttrNone)
			}
			buf.SetRune(startX+width-2, screenY, '┐', st.BorderFg, st.Bg, cell.AttrNone)

		case MdCodeLine:
			buf.SetRune(startX+2, screenY, '│', st.BorderFg, st.Bg, cell.AttrNone)
			for col := 0; col < width-5; col++ {
				r := ' '
				if col < len(md.Text) {
					r = rune(md.Text[col])
				}
				buf.SetRune(startX+3+col, screenY, r, st.CodeFg, st.CodeBg, cell.AttrNone)
			}
			buf.SetRune(startX+width-2, screenY, '│', st.BorderFg, st.Bg, cell.AttrNone)

		case MdCodeFenceEnd:
			for col := 0; col < width-4; col++ {
				buf.SetRune(startX+2+col, screenY, '─', st.BorderFg, st.Bg, cell.AttrNone)
			}
			buf.SetRune(startX+2, screenY, '└', st.BorderFg, st.Bg, cell.AttrNone)
			buf.SetRune(startX+width-2, screenY, '┘', st.BorderFg, st.Bg, cell.AttrNone)

		case MdDivider:
			for col := 0; col < width-4; col++ {
				buf.SetRune(startX+2+col, screenY, '─', st.BorderFg, st.Bg, cell.AttrNone)
			}

		case MdTable:
			for i, r := range md.Text {
				if i+2 < width-2 {
					cColor := st.Fg
					attr := cell.AttrNone
					if r == '|' || r == '-' || r == ':' {
						cColor = st.BorderFg
					}
					buf.SetRune(startX+2+i, screenY, r, cColor, st.Bg, attr)
				}
			}

		default: // MdText
			for i, r := range md.Text {
				if i+2 < width-2 {
					buf.SetRune(startX+2+i, screenY, r, st.Fg, st.Bg, cell.AttrNone)
				}
			}
		}
	}
}
