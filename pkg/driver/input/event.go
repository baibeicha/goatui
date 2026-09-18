package input

import "github.com/baibeicha/goatui/pkg/core/cell"

// EventType represents the category of terminal input event.
type EventType uint8

const (
	EventNone EventType = iota
	EventKey
	EventMouse
	EventResize
	EventPaste
	EventFocus
	EventBlur
)

// KeyType identifies non-character or functional keys.
type KeyType int16

const (
	KeyRune KeyType = iota
	KeyEnter
	KeyEsc
	KeyBackspace
	KeyTab
	KeyBacktab // Shift+Tab
	KeySpace
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyHome
	KeyEnd
	KeyPgUp
	KeyPgDown
	KeyInsert
	KeyDelete
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12
)

// Modifier flags for keyboard and mouse events
const (
	ModCtrl  cell.Modifier = cell.AttrBold
	ModAlt   cell.Modifier = cell.AttrDim
	ModShift cell.Modifier = cell.AttrUnderline
)

// Key represents a keyboard event with modifiers.
type Key struct {
	Type KeyType
	Rune rune
	Mod  cell.Modifier
}

// HasCtrl returns whether the Ctrl modifier is active.
func (k Key) HasCtrl() bool {
	return k.Mod.Has(ModCtrl)
}

// HasAlt returns whether the Alt modifier is active.
func (k Key) HasAlt() bool {
	return k.Mod.Has(ModAlt)
}

// HasShift returns whether the Shift modifier is active.
func (k Key) HasShift() bool {
	return k.Mod.Has(ModShift)
}

// MouseButton denotes which mouse button triggered an event.
type MouseButton uint8

const (
	MouseNone MouseButton = iota
	MouseLeft
	MouseMiddle
	MouseRight
	MouseWheelUp
	MouseWheelDown
	MouseWheelLeft
	MouseWheelRight
)

// MouseAction denotes the physical mouse interaction.
type MouseAction uint8

const (
	MousePress MouseAction = iota
	MouseRelease
	MouseMotion
	MouseDrag
)

// Mouse represents a mouse event with coordinates and modifiers.
type Mouse struct {
	X      int
	Y      int
	Button MouseButton
	Action MouseAction
	Mod    cell.Modifier
}

// HasCtrl returns whether the Ctrl modifier is active.
func (m Mouse) HasCtrl() bool {
	return m.Mod.Has(ModCtrl)
}

// HasAlt returns whether the Alt modifier is active.
func (m Mouse) HasAlt() bool {
	return m.Mod.Has(ModAlt)
}

// HasShift returns whether the Shift modifier is active.
func (m Mouse) HasShift() bool {
	return m.Mod.Has(ModShift)
}

// Event is a unified input event received from the terminal driver.
type Event struct {
	Type      EventType
	Key       Key
	Mouse     Mouse
	Width     int    // Populated for EventResize
	Height    int    // Populated for EventResize
	PasteText string // Populated for EventPaste
}
