package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/style"
)

// ToastLevel represents the severity of a notification card.
type ToastLevel int

const (
	ToastInfo ToastLevel = iota
	ToastSuccess
	ToastWarn
	ToastError
)

// ToastItem represents a floating notification card.
type ToastItem struct {
	ID        string
	Level     ToastLevel
	Title     string
	Message   string
	CreatedAt time.Time
	Duration  time.Duration
	ExpiresAt time.Time
}

// ToastManager coordinates floating toast notifications with timer expirations.
type ToastManager struct {
	toasts  []*ToastItem
	maxShow int
	counter int64
}

// NewToastManager creates a new toast notification manager.
func NewToastManager(maxShow int) *ToastManager {
	if maxShow <= 0 {
		maxShow = 4
	}
	return &ToastManager{
		toasts:  make([]*ToastItem, 0, maxShow),
		maxShow: maxShow,
	}
}

// Add pushes a new notification card into the queue.
func (tm *ToastManager) Add(level ToastLevel, title, message string, duration ...time.Duration) {
	tm.counter++
	dur := 2000 * time.Millisecond
	if len(duration) > 0 && duration[0] > 0 {
		dur = duration[0]
	}

	now := time.Now()
	item := &ToastItem{
		ID:        fmt.Sprintf("toast-%d", tm.counter),
		Level:     level,
		Title:     title,
		Message:   message,
		CreatedAt: now,
		Duration:  dur,
		ExpiresAt: now.Add(dur),
	}

	limit := tm.maxShow
	if limit <= 0 {
		limit = 4
	}
	tm.toasts = append(tm.toasts, item)
	if len(tm.toasts) > limit {
		tm.toasts = tm.toasts[len(tm.toasts)-limit:]
	}
}

// Info adds an informational toast.
func (tm *ToastManager) Info(title, message string) {
	tm.Add(ToastInfo, title, message)
}

// Success adds a success toast.
func (tm *ToastManager) Success(title, message string) {
	tm.Add(ToastSuccess, title, message)
}

// Warn adds a warning toast.
func (tm *ToastManager) Warn(title, message string) {
	tm.Add(ToastWarn, title, message)
}

// Error adds an error toast.
func (tm *ToastManager) Error(title, message string) {
	tm.Add(ToastError, title, message)
}

// Tick evaluates toast expirations and removes outdated notifications.
func (tm *ToastManager) Tick(now time.Time) {
	valid := tm.toasts[:0]
	for _, t := range tm.toasts {
		if now.Before(t.ExpiresAt) {
			valid = append(valid, t)
		}
	}
	tm.toasts = valid
}

// Count returns the number of active toasts.
func (tm *ToastManager) Count() int {
	return len(tm.toasts)
}

// Toasts returns active toasts.
func (tm *ToastManager) Toasts() []*ToastItem {
	return tm.toasts
}

// Dismiss removes a toast by index.
func (tm *ToastManager) Dismiss(index int) bool {
	if index >= 0 && index < len(tm.toasts) {
		tm.toasts = append(tm.toasts[:index], tm.toasts[index+1:]...)
		return true
	}
	return false
}

// DismissByID removes a toast by its unique ID.
func (tm *ToastManager) DismissByID(id string) bool {
	for i, t := range tm.toasts {
		if t.ID == id {
			return tm.Dismiss(i)
		}
	}
	return false
}

// HandleClick checks if a mouse click hit any active toast card or its close button [✕].
func (tm *ToastManager) HandleClick(x, y int, screen buffer.Rect) bool {
	if len(tm.toasts) == 0 || screen.IsEmpty() || screen.Width < 12 || screen.Height < 6 {
		return false
	}
	toastWidth := min(30, screen.Width-4)
	if toastWidth < 8 {
		return false
	}
	toastHeight := 5
	currY := screen.Y + 1

	for i := 0; i < len(tm.toasts); i++ {
		if currY+toastHeight > screen.Bottom()-2 {
			break
		}
		cardX := screen.Right() - toastWidth - 1
		cardRect := buffer.NewRect(cardX, currY, toastWidth, toastHeight)
		if x >= cardRect.X && x < cardRect.Right() && y >= cardRect.Y && y < cardRect.Bottom() {
			tm.Dismiss(i)
			return true
		}
		currY += toastHeight + 1
	}
	return false
}

// wrapToastText breaks a message into up to maxLines lines fitted within maxWidth columns.
func wrapToastText(msg string, maxWidth, maxLines int) []string {
	if maxWidth <= 0 || maxLines <= 0 {
		return nil
	}
	words := strings.Fields(msg)
	if len(words) == 0 {
		return nil
	}

	var lines []string
	var curLine strings.Builder

	for _, w := range words {
		wordLen := buffer.StringWidth(w)
		if curLine.Len() == 0 {
			if wordLen <= maxWidth {
				curLine.WriteString(w)
			} else {
				runes := []rune(w)
				for len(runes) > 0 {
					take := min(len(runes), maxWidth)
					lines = append(lines, string(runes[:take]))
					runes = runes[take:]
					if len(lines) >= maxLines {
						break
					}
				}
				if len(lines) >= maxLines {
					break
				}
			}
		} else {
			if buffer.StringWidth(curLine.String())+1+wordLen <= maxWidth {
				curLine.WriteByte(' ')
				curLine.WriteString(w)
			} else {
				lines = append(lines, curLine.String())
				curLine.Reset()
				if len(lines) >= maxLines {
					break
				}
				curLine.WriteString(w)
			}
		}
	}
	if curLine.Len() > 0 && len(lines) < maxLines {
		lines = append(lines, curLine.String())
	}

	if len(lines) == maxLines && buffer.StringWidth(msg) > buffer.StringWidth(strings.Join(lines, " ")) {
		last := lines[maxLines-1]
		runes := []rune(last)
		ellW := buffer.StringWidth("…")
		for len(runes) > 0 && buffer.StringWidth(string(runes))+ellW > maxWidth {
			runes = runes[:len(runes)-1]
		}
		lines[maxLines-1] = string(runes) + "…"
	}

	return lines
}

// Draw renders active toast notifications into the top-right corner of the screen.
func (tm *ToastManager) Draw(buf *buffer.Buffer, screen buffer.Rect) {
	if len(tm.toasts) == 0 || screen.IsEmpty() || screen.Width < 12 || screen.Height < 6 {
		return
	}

	toastWidth := min(30, screen.Width-4)
	if toastWidth < 8 {
		return
	}
	toastHeight := 5

	now := time.Now()
	currY := screen.Y + 1

	for _, t := range tm.toasts {
		if currY+toastHeight > screen.Bottom()-2 {
			break
		}

		cardX := screen.Right() - toastWidth - 1
		cardRect := buffer.NewRect(cardX, currY, toastWidth, toastHeight)

		borderFg := cell.ColorHex("#00D2FF")
		icon := "[INFO]"
		switch t.Level {
		case ToastSuccess:
			borderFg = cell.ColorHex("#00FFAA")
			icon = "[OK]"
		case ToastWarn:
			borderFg = cell.ColorHex("#FFB86C")
			icon = "[WARN]"
		case ToastError:
			borderFg = cell.ColorHex("#FF5555")
			icon = "[ERR]"
		}

		cardBg := cell.Color256(234)

		// Draw card frame
		titleText := fmt.Sprintf(" %s %s ", icon, t.Title)
		st := style.NewStyle().
			Border(style.BorderRounded).
			BorderForeground(borderFg).
			BorderBackground(cardBg).
			Background(cardBg).
			Title(titleText)
		st.Draw(buf, cardRect, "")

		// Draw close button [✕] in top-right of card
		closeBtnX := cardRect.Right() - 3
		if closeBtnX > cardRect.X+len(titleText) {
			buf.SetRune(closeBtnX, cardRect.Y, '✕', cell.ColorHex("#FF5555"), cardBg, cell.AttrBold)
		}

		// Message text (wrapped to up to 2 lines)
		innerMsgW := cardRect.Width - 4
		if innerMsgW > 3 {
			lines := wrapToastText(t.Message, innerMsgW, 2)
			for li, line := range lines {
				if li < 2 {
					buf.SetString(cardRect.X+2, cardRect.Y+1+li, line, cell.ColorHex("#FFFFFF"), cardBg, cell.AttrNone)
				}
			}
		}

		// Expiration bar (on row cardRect.Y+3)
		remaining := t.ExpiresAt.Sub(now)
		progressRatio := float64(remaining) / float64(t.Duration)
		if progressRatio < 0 {
			progressRatio = 0
		} else if progressRatio > 1.0 {
			progressRatio = 1.0
		}

		barW := cardRect.Width - 4
		filledW := int(float64(barW) * progressRatio)
		for i := 0; i < barW; i++ {
			bx := cardRect.X + 2 + i
			by := cardRect.Y + 3
			if i < filledW {
				buf.SetRune(bx, by, '━', borderFg, cardBg, cell.AttrNone)
			} else {
				buf.SetRune(bx, by, '─', cell.ColorHex("#444455"), cardBg, cell.AttrDim)
			}
		}

		currY += toastHeight + 1
	}
}
