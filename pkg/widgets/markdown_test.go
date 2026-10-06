package widgets

import (
	"testing"

	"github.com/baibeicha/goatui/pkg/core/buffer"
)

func TestMarkdownRenderer_ParseAndRender(t *testing.T) {
	raw := "# Header 1\n## Header 2\nParagraph\n```go\nfmt.Println()\n```\n- [ ] Task 1\n- [x] Task 2\n- Bullet"
	mr := NewMarkdownRenderer(raw)

	if len(mr.Lines) == 0 {
		t.Fatal("expected parsed lines, got 0")
	}

	buf := buffer.NewBuffer(80, 24)
	mr.Render(buf, 0, 0, 80, 24, 0, DefaultMarkdownStyle())

	// Validate cells were written
	c := buf.Cell(3, 0)
	if c == nil || c.Rune != 'H' {
		t.Fatalf("expected 'H' for header title, got %v", c)
	}
}
