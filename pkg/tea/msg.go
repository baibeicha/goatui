package tea

import (
	"github.com/baibeicha/goatui/pkg/driver/input"
	"github.com/baibeicha/goatui/pkg/spatial"
)

// Msg represents any state update event in The Elm Architecture.
type Msg any

// KeyMsg wraps a keyboard event.
type KeyMsg struct {
	input.Key
}

// MouseMsg wraps a mouse event with coordinates and action.
type MouseMsg struct {
	input.Mouse
}

// HitMsg is emitted when a mouse event intersects an interactive region in SpatialMap.
type HitMsg struct {
	Target spatial.HitTarget
	Mouse  input.Mouse
}

// ID returns the registered hit target identifier.
func (h HitMsg) ID() string {
	return h.Target.ID
}

// IsLeftClick returns true if this hit was triggered by a left button press.
func (h HitMsg) IsLeftClick() bool {
	return h.Mouse.Button == input.MouseLeft && h.Mouse.Action == input.MousePress
}

// WindowSizeMsg indicates that the terminal window was resized.
type WindowSizeMsg struct {
	Width  int
	Height int
}

// FocusMsg indicates the terminal window gained focus.
type FocusMsg struct{}

// BlurMsg indicates the terminal window lost focus.
type BlurMsg struct{}

// PasteMsg contains text pasted from the OS clipboard.
type PasteMsg struct {
	Text string
}

// QuitMsg requests graceful termination of the program event loop.
type QuitMsg struct{}
