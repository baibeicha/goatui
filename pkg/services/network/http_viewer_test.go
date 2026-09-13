package network

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/driver/input"
	"github.com/baibeicha/goatui/pkg/tea"
)

func TestHTTPViewerScreen(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","count":42}`))
	}))
	defer ts.Close()

	viewer := NewHTTPViewerScreen(ts.URL)
	cmd := viewer.Init()
	if cmd == nil {
		t.Fatalf("Expected non-nil cmd from Init()")
	}

	msg := cmd()
	respMsg, ok := msg.(HTTPResponseMsg)
	if !ok {
		t.Fatalf("Expected HTTPResponseMsg, got %T", msg)
	}
	if respMsg.StatusCode != 200 {
		t.Errorf("Expected 200 OK, got %d", respMsg.StatusCode)
	}

	// Update viewer with response
	viewer.Update(respMsg)
	if viewer.loading {
		t.Errorf("Viewer should no longer be loading")
	}
	if len(viewer.lines) == 0 {
		t.Errorf("Expected formatted lines in viewer")
	}

	// Test scroll Down
	viewer.Update(tea.KeyMsg{Key: input.Key{Type: input.KeyDown}})

	// Test View rendering
	buf := buffer.NewBuffer(80, 24)
	frame := &tea.Frame{Buffer: buf}
	viewer.View(frame)

	c := buf.Cell(2, 0)
	if c == nil || c.Rune == ' ' {
		t.Errorf("Expected rendered header in HTTPViewer")
	}
}
