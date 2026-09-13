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

// Key represents a keyboard event with modifiers.
type Key struct {
	Type KeyType
	Rune rune
	Mod  cell.Modifier
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

// Event is a unified input event received from the terminal driver.
type Event struct {
	Type      EventType
	Key       Key
	Mouse     Mouse
	Width     int    // Populated for EventResize
	Height    int    // Populated for EventResize
	PasteText string // Populated for EventPaste
}
