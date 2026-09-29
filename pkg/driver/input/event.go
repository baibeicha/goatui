package input

import (
	"fmt"
	"strings"

	"github.com/baibeicha/goatui/pkg/core/cell"
)

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
	EventKittyMode // Notification of terminal Kitty keyboard flags
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

	// Extended functional keys (F13-F35)
	KeyF13
	KeyF14
	KeyF15
	KeyF16
	KeyF17
	KeyF18
	KeyF19
	KeyF20
	KeyF21
	KeyF22
	KeyF23
	KeyF24
	KeyF25
	KeyF26
	KeyF27
	KeyF28
	KeyF29
	KeyF30
	KeyF31
	KeyF32
	KeyF33
	KeyF34
	KeyF35

	// System & Lock keys
	KeyCapsLock
	KeyScrollLock
	KeyNumLock
	KeyPrintScreen
	KeyPause
	KeyMenu

	// Keypad keys
	KeyKp0
	KeyKp1
	KeyKp2
	KeyKp3
	KeyKp4
	KeyKp5
	KeyKp6
	KeyKp7
	KeyKp8
	KeyKp9
	KeyKpDecimal
	KeyKpDivide
	KeyKpMultiply
	KeyKpSubtract
	KeyKpAdd
	KeyKpEnter
	KeyKpEqual
	KeyKpSeparator
	KeyKpLeft
	KeyKpRight
	KeyKpUp
	KeyKpDown
	KeyKpPageUp
	KeyKpPageDown
	KeyKpHome
	KeyKpEnd
	KeyKpInsert
	KeyKpDelete
	KeyKpBegin

	// Media control keys
	KeyMediaPlay
	KeyMediaPause
	KeyMediaPlayPause
	KeyMediaReverse
	KeyMediaStop
	KeyMediaFastForward
	KeyMediaRewind
	KeyMediaTrackNext
	KeyMediaTrackPrevious
	KeyMediaRecord
	KeyLowerVolume
	KeyRaiseVolume
	KeyMuteVolume

	// Modifier key events (when reported individually)
	KeyLeftShift
	KeyRightShift
	KeyLeftCtrl
	KeyRightCtrl
	KeyLeftAlt
	KeyRightAlt
	KeyLeftSuper
	KeyRightSuper
	KeyLeftHyper
	KeyRightHyper
	KeyLeftMeta
	KeyRightMeta
	KeyIsoLevel3Shift
	KeyIsoLevel5Shift
)

// KeyAction denotes the physical keyboard interaction state.
type KeyAction uint8

const (
	KeyPress KeyAction = iota
	KeyRepeat
	KeyRelease
)

// Modifier flags for keyboard and mouse events mapped to cell.Modifier bits.
const (
	ModCtrl     cell.Modifier = cell.AttrBold          // 1 << 0
	ModAlt      cell.Modifier = cell.AttrDim           // 1 << 1
	ModSuper    cell.Modifier = cell.AttrItalic        // 1 << 2
	ModShift    cell.Modifier = cell.AttrUnderline     // 1 << 3
	ModMeta     cell.Modifier = cell.AttrBlink         // 1 << 4
	ModHyper    cell.Modifier = cell.AttrReverse       // 1 << 5
	ModCapsLock cell.Modifier = cell.AttrHidden        // 1 << 6
	ModNumLock  cell.Modifier = cell.AttrStrikethrough // 1 << 7
)

// Kitty keyboard protocol progressive enhancement flags
const (
	KittyModeDisambiguateEscapeCodes = 1 << 0 // 1: Disambiguate escape codes
	KittyModeReportEventTypes        = 1 << 1 // 2: Report key repeat and release events
	KittyModeReportAlternateKeys     = 1 << 2 // 4: Report alternate keys (shifted and base layout)
	KittyModeReportAllKeysAsEscape   = 1 << 3 // 8: Report all keys as escape codes
	KittyModeReportAssociatedText    = 1 << 4 // 16: Report associated text
)

// KittyPushFlags generates the CSI sequence to push flags onto the Kitty keyboard mode stack.
func KittyPushFlags(flags int) string {
	return fmt.Sprintf("\x1b[>%du", flags)
}

// KittyPopFlags generates the CSI sequence to pop count modes from the Kitty keyboard mode stack.
func KittyPopFlags(count int) string {
	if count <= 1 {
		return "\x1b[<u"
	}
	return fmt.Sprintf("\x1b[<%du", count)
}

// KittySetFlags generates the CSI sequence to set, clear, or overwrite Kitty keyboard mode flags.
// mode: 1 = overwrite all flags, 2 = set (OR) given flags, 3 = clear (AND NOT) given flags.
func KittySetFlags(flags int, mode int) string {
	return fmt.Sprintf("\x1b[=%d;%du", flags, mode)
}

// KittyQuery generates the CSI sequence to query current Kitty keyboard protocol flags.
func KittyQuery() string {
	return "\x1b[?u"
}

// KittyDisable generates the sequence to pop modes and reset Kitty keyboard protocol flags to 0.
func KittyDisable() string {
	return "\x1b[<u\x1b[=0u"
}

// Key represents a keyboard event with modifiers and action.
type Key struct {
	Type       KeyType
	Rune       rune
	Mod        cell.Modifier
	Action     KeyAction
	ShiftedKey rune
	BaseKey    rune
	Text       string
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

// HasSuper returns whether the Super (Win / Cmd) modifier is active.
func (k Key) HasSuper() bool {
	return k.Mod.Has(ModSuper)
}

// HasMeta returns whether the Meta modifier is active.
func (k Key) HasMeta() bool {
	return k.Mod.Has(ModMeta)
}

// HasHyper returns whether the Hyper modifier is active.
func (k Key) HasHyper() bool {
	return k.Mod.Has(ModHyper)
}

// HasCapsLock returns whether the CapsLock modifier is active.
func (k Key) HasCapsLock() bool {
	return k.Mod.Has(ModCapsLock)
}

// HasNumLock returns whether the NumLock modifier is active.
func (k Key) HasNumLock() bool {
	return k.Mod.Has(ModNumLock)
}

// IsPress returns whether this is a key press event.
func (k Key) IsPress() bool {
	return k.Action == KeyPress
}

// IsRepeat returns whether this is a key repeat event.
func (k Key) IsRepeat() bool {
	return k.Action == KeyRepeat
}

// IsRelease returns whether this is a key release event.
func (k Key) IsRelease() bool {
	return k.Action == KeyRelease
}

// String returns a human-readable identifier for the key event (e.g. "ctrl+c", "alt+up", "enter").
func (k Key) String() string {
	var parts []string
	if k.HasCtrl() {
		parts = append(parts, "ctrl")
	}
	if k.HasAlt() {
		parts = append(parts, "alt")
	}
	if k.HasSuper() {
		parts = append(parts, "super")
	}
	if k.HasShift() && k.Type != KeyBacktab {
		parts = append(parts, "shift")
	}

	var name string
	switch k.Type {
	case KeyEnter:
		name = "enter"
	case KeyEsc:
		name = "esc"
	case KeyBackspace:
		name = "backspace"
	case KeyTab:
		name = "tab"
	case KeyBacktab:
		name = "backtab"
	case KeySpace:
		name = "space"
	case KeyUp:
		name = "up"
	case KeyDown:
		name = "down"
	case KeyLeft:
		name = "left"
	case KeyRight:
		name = "right"
	case KeyHome:
		name = "home"
	case KeyEnd:
		name = "end"
	case KeyPgUp:
		name = "pgup"
	case KeyPgDown:
		name = "pgdown"
	case KeyInsert:
		name = "insert"
	case KeyDelete:
		name = "delete"
	case KeyCapsLock:
		name = "capslock"
	case KeyScrollLock:
		name = "scrolllock"
	case KeyNumLock:
		name = "numlock"
	case KeyPrintScreen:
		name = "printscreen"
	case KeyPause:
		name = "pause"
	case KeyMenu:
		name = "menu"
	case KeyKpDecimal:
		name = "kp_decimal"
	case KeyKpDivide:
		name = "kp_divide"
	case KeyKpMultiply:
		name = "kp_multiply"
	case KeyKpSubtract:
		name = "kp_subtract"
	case KeyKpAdd:
		name = "kp_add"
	case KeyKpEnter:
		name = "kp_enter"
	case KeyKpEqual:
		name = "kp_equal"
	case KeyKpSeparator:
		name = "kp_separator"
	case KeyKpLeft:
		name = "kp_left"
	case KeyKpRight:
		name = "kp_right"
	case KeyKpUp:
		name = "kp_up"
	case KeyKpDown:
		name = "kp_down"
	case KeyKpPageUp:
		name = "kp_pageup"
	case KeyKpPageDown:
		name = "kp_pagedown"
	case KeyKpHome:
		name = "kp_home"
	case KeyKpEnd:
		name = "kp_end"
	case KeyKpInsert:
		name = "kp_insert"
	case KeyKpDelete:
		name = "kp_delete"
	case KeyKpBegin:
		name = "kp_begin"
	case KeyMediaPlay:
		name = "media_play"
	case KeyMediaPause:
		name = "media_pause"
	case KeyMediaPlayPause:
		name = "media_playpause"
	case KeyMediaReverse:
		name = "media_reverse"
	case KeyMediaStop:
		name = "media_stop"
	case KeyMediaFastForward:
		name = "media_fastforward"
	case KeyMediaRewind:
		name = "media_rewind"
	case KeyMediaTrackNext:
		name = "media_tracknext"
	case KeyMediaTrackPrevious:
		name = "media_trackprevious"
	case KeyMediaRecord:
		name = "media_record"
	case KeyLowerVolume:
		name = "lower_volume"
	case KeyRaiseVolume:
		name = "raise_volume"
	case KeyMuteVolume:
		name = "mute_volume"
	case KeyLeftShift:
		name = "left_shift"
	case KeyRightShift:
		name = "right_shift"
	case KeyLeftCtrl:
		name = "left_ctrl"
	case KeyRightCtrl:
		name = "right_ctrl"
	case KeyLeftAlt:
		name = "left_alt"
	case KeyRightAlt:
		name = "right_alt"
	case KeyLeftSuper:
		name = "left_super"
	case KeyRightSuper:
		name = "right_super"
	case KeyLeftHyper:
		name = "left_hyper"
	case KeyRightHyper:
		name = "right_hyper"
	case KeyLeftMeta:
		name = "left_meta"
	case KeyRightMeta:
		name = "right_meta"
	case KeyIsoLevel3Shift:
		name = "iso_level3_shift"
	case KeyIsoLevel5Shift:
		name = "iso_level5_shift"
	default:
		if k.Type >= KeyF1 && k.Type <= KeyF35 {
			name = fmt.Sprintf("f%d", int(k.Type-KeyF1)+1)
		} else if k.Type >= KeyKp0 && k.Type <= KeyKp9 {
			name = fmt.Sprintf("kp_%d", int(k.Type-KeyKp0))
		} else if k.Rune != 0 {
			name = string(k.Rune)
		} else {
			name = "key"
		}
	}

	parts = append(parts, name)
	res := strings.Join(parts, "+")
	if k.Action == KeyRelease {
		res += ":release"
	} else if k.Action == KeyRepeat {
		res += ":repeat"
	}
	return res
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

// HasSuper returns whether the Super modifier is active.
func (m Mouse) HasSuper() bool {
	return m.Mod.Has(ModSuper)
}

// Event is a unified input event received from the terminal driver.
type Event struct {
	Type       EventType
	Key        Key
	Mouse      Mouse
	Width      int    // Populated for EventResize
	Height     int    // Populated for EventResize
	PasteText  string // Populated for EventPaste
	KittyFlags int    // Populated for EventKittyMode
}
