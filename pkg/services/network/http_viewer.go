package network

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/driver/input"
	"github.com/baibeicha/goatui/pkg/router"
	"github.com/baibeicha/goatui/pkg/style"
	"github.com/baibeicha/goatui/pkg/tea"
	"github.com/baibeicha/goatui/pkg/theme"
	"github.com/baibeicha/goatui/pkg/window"
)

// HTTPResponseMsg conveys the outcome of an asynchronous HTTP request.
type HTTPResponseMsg struct {
	URL        string
	StatusCode int
	StatusText string
	Duration   time.Duration
	Headers    http.Header
	Body       string
	Err        error
}

// FetchURLCmd returns a tea.Cmd that performs an asynchronous HTTP GET request.
func FetchURLCmd(urlStr string) tea.Cmd {
	return func() tea.Msg {
		start := time.Now()
		client := &http.Client{
			Timeout: 10 * time.Second,
		}

		resp, err := client.Get(urlStr)
		duration := time.Since(start)
		if err != nil {
			return HTTPResponseMsg{
				URL:      urlStr,
				Duration: duration,
				Err:      err,
			}
		}
		defer resp.Body.Close()

		bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024)) // 512 KB cap
		if err != nil {
			return HTTPResponseMsg{
				URL:        urlStr,
				StatusCode: resp.StatusCode,
				StatusText: resp.Status,
				Duration:   duration,
				Headers:    resp.Header,
				Err:        err,
			}
		}

		// Prettify JSON if applicable
		bodyStr := string(bodyBytes)
		if strings.Contains(resp.Header.Get("Content-Type"), "json") {
			var pretty bytes.Buffer
			if err := json.Indent(&pretty, bodyBytes, "", "  "); err == nil {
				bodyStr = pretty.String()
			}
		}

		return HTTPResponseMsg{
			URL:        urlStr,
			StatusCode: resp.StatusCode,
			StatusText: resp.Status,
			Duration:   duration,
			Headers:    resp.Header,
			Body:       bodyStr,
		}
	}
}

// HTTPViewerScreen displays HTTP/HTTPS responses with status codes, headers, and formatted payloads.
type HTTPViewerScreen struct {
	window.BaseScreen

	url        string
	loading    bool
	statusCode int
	statusText string
	duration   time.Duration
	headers    http.Header
	lines      []string
	scroll     int
	err        error
}

// NewHTTPViewerScreen creates an HTTP inspector for the given URL.
func NewHTTPViewerScreen(urlStr string) *HTTPViewerScreen {
	return &HTTPViewerScreen{
		url:     urlStr,
		loading: true,
	}
}

// OnMount captures the target URL from the router context if specified.
func (h *HTTPViewerScreen) OnMount(ctx *router.RouteContext) {
	if ctx != nil {
		if u, ok := ctx.Params["url"]; ok && u != "" {
			h.url = u
		} else if strings.HasPrefix(ctx.Path, "http://") || strings.HasPrefix(ctx.Path, "https://") {
			h.url = ctx.Path
		}
	}
	h.loading = true
}

// Init begins the asynchronous HTTP request.
func (h *HTTPViewerScreen) Init() tea.Cmd {
	if h.url != "" {
		return FetchURLCmd(h.url)
	}
	return nil
}

// Update processes incoming response messages and scrolling keyboard inputs.
func (h *HTTPViewerScreen) Update(msg tea.Msg) (window.Screen, tea.Cmd) {
	switch m := msg.(type) {
	case HTTPResponseMsg:
		h.loading = false
		h.statusCode = m.StatusCode
		h.statusText = m.StatusText
		h.duration = m.Duration
		h.headers = m.Headers
		h.err = m.Err
		if m.Body != "" {
			h.lines = strings.Split(m.Body, "\n")
		} else if m.Err != nil {
			h.lines = []string{fmt.Sprintf("HTTP Error: %v", m.Err)}
		}

	case tea.KeyMsg:
		switch m.Key.Type {
		case input.KeyUp:
			if h.scroll > 0 {
				h.scroll--
			}
		case input.KeyDown:
			if h.scroll < len(h.lines)-1 {
				h.scroll++
			}
		case input.KeyPgUp:
			h.scroll = max(0, h.scroll-15)
		case input.KeyPgDown:
			h.scroll = min(len(h.lines)-1, h.scroll+15)
		case input.KeyHome:
			h.scroll = 0
		case input.KeyEnd:
			h.scroll = max(0, len(h.lines)-1)
		case input.KeyRune:
			if m.Key.Rune == 'r' {
				h.loading = true
				return h, FetchURLCmd(h.url)
			}
		}
	}

	return h, nil
}

// View draws the HTTP Inspector interface into the terminal frame buffer.
func (h *HTTPViewerScreen) View(f *tea.Frame) {
	area := f.Area()
	if area.IsEmpty() {
		return
	}

	curTheme := theme.Current()
	p := curTheme.Colors

	// Header summary box
	headerArea := buffer.NewRect(area.X, area.Y, area.Width, 3)
	f.Buffer.Fill(headerArea, cell.Cell{
		Rune:   ' ',
		Width:  1,
		FgType: cell.ColorDefault,
		BgType: p.Background.Type,
		Bg:     p.Background.Value,
	})

	statusColor := p.Success
	if h.statusCode >= 400 || h.err != nil {
		statusColor = p.Danger
	} else if h.statusCode >= 300 {
		statusColor = p.Warning
	}

	statusStr := fmt.Sprintf("[%d %s]", h.statusCode, h.statusText)
	if h.loading {
		statusStr = "[CONNECTING...]"
		statusColor = p.Accent
	} else if h.err != nil {
		statusStr = "[REQUEST FAILED]"
	}

	f.Buffer.SetString(area.X+2, area.Y, "🌐 URL: "+h.url, p.Accent, p.Background, cell.AttrBold)
	f.Buffer.SetString(area.X+2, area.Y+1, statusStr, statusColor, p.Background, cell.AttrBold)
	if !h.loading && h.duration > 0 {
		meta := fmt.Sprintf("Latency: %v | Lines: %d | [↑/↓/PgUp/PgDn] Scroll | [R] Reload", h.duration.Round(time.Millisecond), len(h.lines))
		f.Buffer.SetString(area.X+len(statusStr)+4, area.Y+1, meta, p.Foreground, p.Background, cell.AttrDim)
	}

	// Body content box
	bodyArea := buffer.NewRect(area.X, area.Y+3, area.Width, area.Height-3)
	if bodyArea.Height <= 0 {
		return
	}

	box := style.NewStyle().
		Border(curTheme.Borders.Border).
		BorderForeground(p.Secondary)
	box.Draw(f.Buffer, bodyArea, " Response Payload ")

	inner := bodyArea.Inset(1, 1)
	if inner.IsEmpty() {
		return
	}

	if h.loading {
		f.Buffer.SetString(inner.X+2, inner.Y+2, "Sending HTTP request and awaiting response...", p.Accent, cell.DefaultColor(), cell.AttrItalic)
		return
	}

	for i := 0; i < inner.Height; i++ {
		idx := h.scroll + i
		if idx >= len(h.lines) {
			break
		}
		lineNum := fmt.Sprintf("%4d │ ", idx+1)
		f.Buffer.SetString(inner.X+1, inner.Y+i, lineNum, p.Secondary, cell.DefaultColor(), cell.AttrDim)

		lineContent := h.lines[idx]
		maxW := inner.Width - len(lineNum) - 2
		if maxW > 0 && len(lineContent) > maxW {
			lineContent = lineContent[:maxW]
		}
		f.Buffer.SetString(inner.X+1+len(lineNum), inner.Y+i, lineContent, p.Foreground, cell.DefaultColor(), 0)
	}
}
