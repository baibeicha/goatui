package input

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/baibeicha/goatui/pkg/core/cell"
)

// Parser parses raw terminal input bytes into structured events.
type Parser struct {
	buf            []byte
	inPaste        bool
	pasteData      []byte
	kittySupported bool
	kittyFlags     int
}

// NewParser creates a new input parser.
func NewParser() *Parser {
	return &Parser{
		buf:       make([]byte, 0, 512),
		pasteData: make([]byte, 0, 1024),
	}
}

// KittySupported returns true if the terminal has confirmed Kitty Keyboard Protocol support.
func (p *Parser) KittySupported() bool {
	return p.kittySupported
}

// KittyFlags returns the active Kitty Keyboard Protocol flags reported by the terminal.
func (p *Parser) KittyFlags() int {
	return p.kittyFlags
}

// Parse processes incoming bytes and invokes handler for each successfully decoded event.
// Unparsed / incomplete trailing escape sequences are retained in the buffer for the next call.
func (p *Parser) Parse(data []byte, handler func(Event)) {
	p.buf = append(p.buf, data...)

	for len(p.buf) > 0 {
		// If currently inside a bracketed paste block
		if p.inPaste {
			if endIdx := findPasteEnd(p.buf); endIdx >= 0 {
				p.pasteData = append(p.pasteData, p.buf[:endIdx]...)
				handler(Event{
					Type:      EventPaste,
					PasteText: string(p.pasteData),
				})
				p.pasteData = p.pasteData[:0]
				p.inPaste = false
				p.buf = p.buf[endIdx+len("\x1b[201~"):]
				continue
			} else {
				p.pasteData = append(p.pasteData, p.buf...)
				p.buf = p.buf[:0]
				return
			}
		}

		// Escape sequence started
		if p.buf[0] == 0x1B {
			if len(p.buf) == 1 {
				// Standalone Esc key
				handler(Event{
					Type: EventKey,
					Key:  Key{Type: KeyEsc, Action: KeyPress},
				})
				p.buf = p.buf[1:]
				continue
			}

			// Check bracketed paste start: \x1b[200~
			if hasPrefix(p.buf, "\x1b[200~") {
				p.inPaste = true
				p.buf = p.buf[len("\x1b[200~"):]
				continue
			}

			// Focus reporting
			if hasPrefix(p.buf, "\x1b[I") {
				handler(Event{Type: EventFocus})
				p.buf = p.buf[3:]
				continue
			}
			if hasPrefix(p.buf, "\x1b[O") {
				handler(Event{Type: EventBlur})
				p.buf = p.buf[3:]
				continue
			}

			// Try parsing CSI sequence (\x1b[...)
			if p.buf[1] == '[' {
				consumed, ev, ok, incomplete := p.parseCSI(p.buf)
				if incomplete {
					// Wait for more bytes
					return
				}
				if ok {
					if ev.Type == EventKittyMode {
						p.kittySupported = true
						p.kittyFlags = ev.KittyFlags
					}
					handler(ev)
				}
				if consumed > 0 {
					p.buf = p.buf[consumed:]
					continue
				}
			}

			// Handle OSC sequences (\x1b]...) e.g. terminal color queries or window titles
			if p.buf[1] == ']' {
				endIdx := -1
				for i := 2; i < len(p.buf); i++ {
					if p.buf[i] == '\a' {
						endIdx = i + 1
						break
					}
					if p.buf[i] == 0x1B && i+1 < len(p.buf) && p.buf[i+1] == '\\' {
						endIdx = i + 2
						break
					}
				}
				if endIdx == -1 {
					if len(p.buf) > 256 {
						p.buf = p.buf[2:]
						continue
					}
					return
				}
				p.buf = p.buf[endIdx:]
				continue
			}

			// SS3 sequences (\x1bO...) e.g. F1-F4 and DECCKM Application Cursor keys
			if p.buf[1] == 'O' {
				if len(p.buf) < 3 {
					// Incomplete SS3 sequence, wait for more bytes
					return
				}
				ev, ok := parseSS3(p.buf[2])
				if ok {
					handler(ev)
					p.buf = p.buf[3:]
					continue
				}
			}

			// Alt + Key combination (\x1b<char>)
			r, size := utf8.DecodeRune(p.buf[1:])
			if r != utf8.RuneError {
				handler(Event{
					Type: EventKey,
					Key: Key{
						Type:   KeyRune,
						Rune:   r,
						Mod:    ModAlt,
						Action: KeyPress,
					},
				})
				p.buf = p.buf[1+size:]
				continue
			}
		}

		// Control characters (Ctrl+A..Z, Tab, Enter, Backspace)
		b := p.buf[0]
		switch b {
		case 0x00: // Ctrl+Space or Ctrl+@
			handler(Event{
				Type: EventKey,
				Key:  Key{Type: KeySpace, Mod: ModCtrl, Action: KeyPress},
			})
			p.buf = p.buf[1:]
			continue
		case 0x09: // Tab
			handler(Event{Type: EventKey, Key: Key{Type: KeyTab, Action: KeyPress, Text: "\t"}})
			p.buf = p.buf[1:]
			continue
		case 0x0A, 0x0D: // Enter (LF or CR)
			handler(Event{Type: EventKey, Key: Key{Type: KeyEnter, Action: KeyPress, Text: "\r"}})
			if b == 0x0D && len(p.buf) > 1 && p.buf[1] == 0x0A {
				p.buf = p.buf[2:]
			} else {
				p.buf = p.buf[1:]
			}
			continue
		case 0x08, 0x7F: // Backspace
			handler(Event{Type: EventKey, Key: Key{Type: KeyBackspace, Action: KeyPress}})
			p.buf = p.buf[1:]
			continue
		}

		// Ctrl+A through Ctrl+Z
		if b >= 1 && b <= 26 && b != 0x09 && b != 0x0A && b != 0x0D {
			handler(Event{
				Type: EventKey,
				Key: Key{
					Type:   KeyRune,
					Rune:   rune('a' + b - 1),
					Mod:    ModCtrl,
					Action: KeyPress,
				},
			})
			p.buf = p.buf[1:]
			continue
		}

		// Standard UTF-8 Rune
		r, size := utf8.DecodeRune(p.buf)
		if r != utf8.RuneError {
			kt := KeyRune
			if r == ' ' {
				kt = KeySpace
			}
			handler(Event{
				Type: EventKey,
				Key: Key{
					Type:   kt,
					Rune:   r,
					Action: KeyPress,
					Text:   string(r),
				},
			})
			p.buf = p.buf[size:]
			continue
		}

		// Invalid byte: skip
		p.buf = p.buf[1:]
	}
}

func (p *Parser) parseCSI(b []byte) (consumed int, ev Event, ok bool, incomplete bool) {
	if len(b) < 3 {
		return 0, Event{}, false, true
	}

	// Find terminating character (ASCII range 0x40 - 0x7E)
	endIdx := -1
	for i := 2; i < len(b); i++ {
		c := b[i]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || c == '~' || c == '@' {
			endIdx = i
			break
		}
	}
	if endIdx == -1 {
		if len(b) > 64 {
			// Malformed sequence: drop first 2 bytes
			return 2, Event{}, false, false
		}
		return 0, Event{}, false, true
	}

	term := b[endIdx]
	seq := string(b[2:endIdx])
	consumed = endIdx + 1

	// SGR 1006 Mouse: \x1b[<btn;x;yM or \x1b[<btn;x;ym
	if b[2] == '<' && (term == 'M' || term == 'm') {
		return parseSGRMouse(b)
	}

	// Kitty Keyboard Protocol: \x1b[...u
	if term == 'u' {
		return parseKittyKeyboard(seq, consumed)
	}

	// Mode push/pop or DA query responses: ignore gracefully
	if len(seq) > 0 && (seq[0] == '<' || seq[0] == '>' || seq[0] == '=') {
		return consumed, Event{}, false, false
	}
	if term == 'c' {
		// Device Attributes response (\x1b[?...c)
		return consumed, Event{}, false, false
	}

	// Legacy & Kitty-enhanced functional keys (arrows, Home/End, Delete, F-keys)
	var mod cell.Modifier
	var action KeyAction = KeyPress

	keySeq := seq
	modPart := ""
	if semiIdx := strings.IndexByte(seq, ';'); semiIdx != -1 {
		keySeq = seq[:semiIdx]
		modPart = seq[semiIdx+1:]
	}

	if modPart == "" {
		if colonIdx := strings.IndexByte(keySeq, ':'); colonIdx != -1 {
			evTypeStr := keySeq[colonIdx+1:]
			keySeq = keySeq[:colonIdx]
			if evTypeStr == "3" {
				action = KeyRelease
			} else if evTypeStr == "2" {
				action = KeyRepeat
			}
		}
	} else {
		modStr := modPart
		if colonIdx := strings.IndexByte(modPart, ':'); colonIdx != -1 {
			modStr = modPart[:colonIdx]
			evTypeStr := modPart[colonIdx+1:]
			if evTypeStr == "3" {
				action = KeyRelease
			} else if evTypeStr == "2" {
				action = KeyRepeat
			}
		}
		mVal := 0
		for i := 0; i < len(modStr); i++ {
			if modStr[i] >= '0' && modStr[i] <= '9' {
				mVal = mVal*10 + int(modStr[i]-'0')
			}
		}
		if mVal > 1 {
			m := mVal - 1
			if m&1 != 0 {
				mod.Add(ModShift)
			}
			if m&2 != 0 {
				mod.Add(ModAlt)
			}
			if m&4 != 0 {
				mod.Add(ModCtrl)
			}
			if m&8 != 0 {
				mod.Add(ModSuper)
			}
			if m&16 != 0 {
				mod.Add(ModHyper)
			}
			if m&32 != 0 {
				mod.Add(ModMeta)
			}
			if m&64 != 0 {
				mod.Add(ModCapsLock)
			}
			if m&128 != 0 {
				mod.Add(ModNumLock)
			}
		}
	}

	// Standard functional arrows
	switch term {
	case 'A':
		return consumed, Event{Type: EventKey, Key: Key{Type: KeyUp, Mod: mod, Action: action}}, true, false
	case 'B':
		return consumed, Event{Type: EventKey, Key: Key{Type: KeyDown, Mod: mod, Action: action}}, true, false
	case 'C':
		return consumed, Event{Type: EventKey, Key: Key{Type: KeyRight, Mod: mod, Action: action}}, true, false
	case 'D':
		return consumed, Event{Type: EventKey, Key: Key{Type: KeyLeft, Mod: mod, Action: action}}, true, false
	case 'H':
		return consumed, Event{Type: EventKey, Key: Key{Type: KeyHome, Mod: mod, Action: action}}, true, false
	case 'F':
		return consumed, Event{Type: EventKey, Key: Key{Type: KeyEnd, Mod: mod, Action: action}}, true, false
	case 'P':
		return consumed, Event{Type: EventKey, Key: Key{Type: KeyF1, Mod: mod, Action: action}}, true, false
	case 'Q':
		return consumed, Event{Type: EventKey, Key: Key{Type: KeyF2, Mod: mod, Action: action}}, true, false
	case 'R':
		return consumed, Event{Type: EventKey, Key: Key{Type: KeyF3, Mod: mod, Action: action}}, true, false
	case 'S':
		return consumed, Event{Type: EventKey, Key: Key{Type: KeyF4, Mod: mod, Action: action}}, true, false
	case 'E':
		return consumed, Event{Type: EventKey, Key: Key{Type: KeyKpBegin, Mod: mod, Action: action}}, true, false
	case 'Z':
		return consumed, Event{Type: EventKey, Key: Key{Type: KeyBacktab, Mod: mod, Action: action}}, true, false
	}

	// Numeric ~ sequences (e.g. \x1b[3~ = Delete, \x1b[5~ = PgUp)
	if term == '~' {
		switch keySeq {
		case "1", "7":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyHome, Mod: mod, Action: action}}, true, false
		case "2":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyInsert, Mod: mod, Action: action}}, true, false
		case "3":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyDelete, Mod: mod, Action: action}}, true, false
		case "4", "8":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyEnd, Mod: mod, Action: action}}, true, false
		case "5":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyPgUp, Mod: mod, Action: action}}, true, false
		case "6":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyPgDown, Mod: mod, Action: action}}, true, false
		case "11":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyF1, Mod: mod, Action: action}}, true, false
		case "12":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyF2, Mod: mod, Action: action}}, true, false
		case "13":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyF3, Mod: mod, Action: action}}, true, false
		case "14":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyF4, Mod: mod, Action: action}}, true, false
		case "15":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyF5, Mod: mod, Action: action}}, true, false
		case "17":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyF6, Mod: mod, Action: action}}, true, false
		case "18":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyF7, Mod: mod, Action: action}}, true, false
		case "19":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyF8, Mod: mod, Action: action}}, true, false
		case "20":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyF9, Mod: mod, Action: action}}, true, false
		case "21":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyF10, Mod: mod, Action: action}}, true, false
		case "23":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyF11, Mod: mod, Action: action}}, true, false
		case "24":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyF12, Mod: mod, Action: action}}, true, false
		case "25":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyF13, Mod: mod, Action: action}}, true, false
		case "26":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyF14, Mod: mod, Action: action}}, true, false
		case "28":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyF15, Mod: mod, Action: action}}, true, false
		case "29":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyMenu, Mod: mod, Action: action}}, true, false
		case "31":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyF17, Mod: mod, Action: action}}, true, false
		case "32":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyF18, Mod: mod, Action: action}}, true, false
		case "33":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyF19, Mod: mod, Action: action}}, true, false
		case "34":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyF20, Mod: mod, Action: action}}, true, false
		case "57427":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyKpBegin, Mod: mod, Action: action}}, true, false
		}
	}

	return consumed, Event{}, false, false
}

func parseSS3(b byte) (Event, bool) {
	switch b {
	case 'A':
		return Event{Type: EventKey, Key: Key{Type: KeyUp, Action: KeyPress}}, true
	case 'B':
		return Event{Type: EventKey, Key: Key{Type: KeyDown, Action: KeyPress}}, true
	case 'C':
		return Event{Type: EventKey, Key: Key{Type: KeyRight, Action: KeyPress}}, true
	case 'D':
		return Event{Type: EventKey, Key: Key{Type: KeyLeft, Action: KeyPress}}, true
	case 'H':
		return Event{Type: EventKey, Key: Key{Type: KeyHome, Action: KeyPress}}, true
	case 'F':
		return Event{Type: EventKey, Key: Key{Type: KeyEnd, Action: KeyPress}}, true
	case 'P':
		return Event{Type: EventKey, Key: Key{Type: KeyF1, Action: KeyPress}}, true
	case 'Q':
		return Event{Type: EventKey, Key: Key{Type: KeyF2, Action: KeyPress}}, true
	case 'R':
		return Event{Type: EventKey, Key: Key{Type: KeyF3, Action: KeyPress}}, true
	case 'S':
		return Event{Type: EventKey, Key: Key{Type: KeyF4, Action: KeyPress}}, true
	}
	return Event{}, false
}

func parseSGRMouse(b []byte) (consumed int, ev Event, ok bool, incomplete bool) {
	endIdx := -1
	var isRelease bool
	for i := 3; i < len(b); i++ {
		if b[i] == 'M' {
			endIdx = i
			isRelease = false
			break
		}
		if b[i] == 'm' {
			endIdx = i
			isRelease = true
			break
		}
	}
	if endIdx == -1 {
		if len(b) > 24 {
			return 3, Event{}, false, false
		}
		return 0, Event{}, false, true
	}

	consumed = endIdx + 1
	body := string(b[3:endIdx])

	btn, x, y, parsed := parseThreeInts(body)
	if !parsed {
		return consumed, Event{}, false, false
	}

	var mouseMod cell.Modifier
	if (btn & 4) != 0 {
		mouseMod.Add(ModShift)
	}
	if (btn & 8) != 0 {
		mouseMod.Add(ModAlt)
	}
	if (btn & 16) != 0 {
		mouseMod.Add(ModCtrl)
	}

	mouse := Mouse{
		X:   x - 1, // Convert 1-indexed terminal coords to 0-indexed
		Y:   y - 1,
		Mod: mouseMod,
	}

	if isRelease {
		mouse.Action = MouseRelease
	} else {
		mouse.Action = MousePress
	}

	btnBase := btn & 0x03
	isMotion := (btn & 32) != 0
	isWheel := (btn & 64) != 0

	if isMotion {
		if btnBase == 3 || isRelease {
			mouse.Action = MouseMotion
		} else {
			mouse.Action = MouseDrag
		}
	}

	if isWheel {
		switch btn & 0x03 {
		case 0:
			mouse.Button = MouseWheelUp
		case 1:
			mouse.Button = MouseWheelDown
		case 2:
			mouse.Button = MouseWheelLeft
		case 3:
			mouse.Button = MouseWheelRight
		}
	} else {
		switch btnBase {
		case 0:
			mouse.Button = MouseLeft
		case 1:
			mouse.Button = MouseMiddle
		case 2:
			mouse.Button = MouseRight
		default:
			mouse.Button = MouseNone
		}
	}

	return consumed, Event{Type: EventMouse, Mouse: mouse}, true, false
}

// parseKittyKeyboard decodes CSI <field1> [ ; <field2> [ ; <field3> ] ] u
func parseKittyKeyboard(seq string, consumed int) (int, Event, bool, bool) {
	if len(seq) == 0 {
		return consumed, Event{}, false, false
	}

	// Check for query response: \x1b[?flags u
	if seq[0] == '?' {
		flags := 0
		for i := 1; i < len(seq); i++ {
			if seq[i] >= '0' && seq[i] <= '9' {
				flags = flags*10 + int(seq[i]-'0')
			}
		}
		return consumed, Event{
			Type:       EventKittyMode,
			KittyFlags: flags,
		}, true, false
	}

	// Ignore control / mode-set sequences if echoed
	if seq[0] == '>' || seq[0] == '<' || seq[0] == '=' {
		return consumed, Event{}, false, false
	}

	fields := strings.Split(seq, ";")

	// Field 1: code [: shifted_key [: base_layout_key]]
	sub1 := strings.Split(fields[0], ":")
	code, err := strconv.Atoi(sub1[0])
	if err != nil {
		return consumed, Event{}, false, false
	}

	var shiftedKey rune
	var baseKey rune
	if len(sub1) > 1 && sub1[1] != "" {
		if sk, err := strconv.Atoi(sub1[1]); err == nil {
			shiftedKey = rune(sk)
		}
	}
	if len(sub1) > 2 && sub1[2] != "" {
		if bk, err := strconv.Atoi(sub1[2]); err == nil {
			baseKey = rune(bk)
		}
	}

	// Field 2: modifiers [: event-type]
	modVal := 1
	action := KeyPress
	if len(fields) > 1 && fields[1] != "" {
		sub2 := strings.Split(fields[1], ":")
		if sub2[0] != "" {
			if mv, err := strconv.Atoi(sub2[0]); err == nil {
				modVal = mv
			}
		}
		if len(sub2) > 1 && sub2[1] != "" {
			if act, err := strconv.Atoi(sub2[1]); err == nil {
				switch act {
				case 1:
					action = KeyPress
				case 2:
					action = KeyRepeat
				case 3:
					action = KeyRelease
				}
			}
		}
	}

	// Field 3: text-as-codepoints
	var text string
	if len(fields) > 2 && fields[2] != "" {
		codepointStrs := strings.Split(fields[2], ":")
		var runes []rune
		for _, cpStr := range codepointStrs {
			if cp, err := strconv.Atoi(cpStr); err == nil {
				runes = append(runes, rune(cp))
			}
		}
		if len(runes) > 0 {
			text = string(runes)
		}
	}

	// Parse modifiers: 1 + bitmask
	var mod cell.Modifier
	if modVal > 1 {
		m := modVal - 1
		if m&1 != 0 {
			mod.Add(ModShift)
		}
		if m&2 != 0 {
			mod.Add(ModAlt)
		}
		if m&4 != 0 {
			mod.Add(ModCtrl)
		}
		if m&8 != 0 {
			mod.Add(ModSuper)
		}
		if m&16 != 0 {
			mod.Add(ModHyper)
		}
		if m&32 != 0 {
			mod.Add(ModMeta)
		}
		if m&64 != 0 {
			mod.Add(ModCapsLock)
		}
		if m&128 != 0 {
			mod.Add(ModNumLock)
		}
	}

	var kt KeyType = KeyRune
	var r rune

	switch code {
	case 27: // Escape
		kt = KeyEsc
	case 13: // Enter
		kt = KeyEnter
	case 9: // Tab
		if mod.Has(ModShift) {
			kt = KeyBacktab
		} else {
			kt = KeyTab
		}
	case 127, 8: // Backspace
		kt = KeyBackspace
	case 32: // Space
		kt = KeySpace
		r = ' '

	// Kitty PUA: Cursor keys
	case 57348:
		kt = KeyUp
	case 57349:
		kt = KeyDown
	case 57346:
		kt = KeyLeft
	case 57347:
		kt = KeyRight
	case 57352:
		kt = KeyHome
	case 57353:
		kt = KeyEnd
	case 57350:
		kt = KeyPgUp
	case 57351:
		kt = KeyPgDown
	case 57344:
		kt = KeyInsert
	case 57345:
		kt = KeyDelete

	// Kitty PUA: Lock and System keys
	case 57358:
		kt = KeyCapsLock
	case 57359:
		kt = KeyScrollLock
	case 57360:
		kt = KeyNumLock
	case 57361:
		kt = KeyPrintScreen
	case 57362:
		kt = KeyPause
	case 57363:
		kt = KeyMenu

	// Keypad keys
	case 57399:
		kt = KeyKp0
		r = '0'
	case 57400:
		kt = KeyKp1
		r = '1'
	case 57401:
		kt = KeyKp2
		r = '2'
	case 57402:
		kt = KeyKp3
		r = '3'
	case 57403:
		kt = KeyKp4
		r = '4'
	case 57404:
		kt = KeyKp5
		r = '5'
	case 57405:
		kt = KeyKp6
		r = '6'
	case 57406:
		kt = KeyKp7
		r = '7'
	case 57407:
		kt = KeyKp8
		r = '8'
	case 57408:
		kt = KeyKp9
		r = '9'
	case 57409:
		kt = KeyKpDecimal
		r = '.'
	case 57410:
		kt = KeyKpDivide
		r = '/'
	case 57411:
		kt = KeyKpMultiply
		r = '*'
	case 57412:
		kt = KeyKpSubtract
		r = '-'
	case 57413:
		kt = KeyKpAdd
		r = '+'
	case 57414:
		kt = KeyKpEnter
	case 57415:
		kt = KeyKpEqual
		r = '='
	case 57416:
		kt = KeyKpSeparator
		r = ','
	case 57417:
		kt = KeyKpLeft
	case 57418:
		kt = KeyKpRight
	case 57419:
		kt = KeyKpUp
	case 57420:
		kt = KeyKpDown
	case 57421:
		kt = KeyKpPageUp
	case 57422:
		kt = KeyKpPageDown
	case 57423:
		kt = KeyKpHome
	case 57424:
		kt = KeyKpEnd
	case 57425:
		kt = KeyKpInsert
	case 57426:
		kt = KeyKpDelete
	case 57427:
		kt = KeyKpBegin

	// Media keys
	case 57428:
		kt = KeyMediaPlay
	case 57429:
		kt = KeyMediaPause
	case 57430:
		kt = KeyMediaPlayPause
	case 57431:
		kt = KeyMediaReverse
	case 57432:
		kt = KeyMediaStop
	case 57433:
		kt = KeyMediaFastForward
	case 57434:
		kt = KeyMediaRewind
	case 57435:
		kt = KeyMediaTrackNext
	case 57436:
		kt = KeyMediaTrackPrevious
	case 57437:
		kt = KeyMediaRecord
	case 57438:
		kt = KeyLowerVolume
	case 57439:
		kt = KeyRaiseVolume
	case 57440:
		kt = KeyMuteVolume

	// Modifier keys (reported when individually pressed)
	case 57441:
		kt = KeyLeftShift
	case 57442:
		kt = KeyLeftCtrl
	case 57443:
		kt = KeyLeftAlt
	case 57444:
		kt = KeyLeftSuper
	case 57445:
		kt = KeyLeftHyper
	case 57446:
		kt = KeyLeftMeta
	case 57447:
		kt = KeyRightShift
	case 57448:
		kt = KeyRightCtrl
	case 57449:
		kt = KeyRightAlt
	case 57450:
		kt = KeyRightSuper
	case 57451:
		kt = KeyRightHyper
	case 57452:
		kt = KeyRightMeta
	case 57453:
		kt = KeyIsoLevel3Shift
	case 57454:
		kt = KeyIsoLevel5Shift

	default:
		if code >= 57364 && code <= 57375 {
			kt = KeyType(int(KeyF1) + (code - 57364))
		} else if code >= 57376 && code <= 57398 {
			kt = KeyType(int(KeyF13) + (code - 57376))
		} else {
			kt = KeyRune
			r = rune(code)
			// Pure text event without specific key code
			if code == 0 && text != "" {
				r, _ = utf8.DecodeRuneInString(text)
			}
			// If Shift is pressed without Ctrl/Alt, adjust rune to shiftedKey or uppercase
			if mod.Has(ModShift) && !mod.Has(ModCtrl) && !mod.Has(ModAlt) {
				if shiftedKey != 0 {
					r = shiftedKey
				} else if r >= 'a' && r <= 'z' {
					r = r - 'a' + 'A'
				}
			}
		}
	}

	if text == "" {
		if kt == KeyRune && r != 0 {
			text = string(r)
		} else if kt == KeySpace {
			text = " "
		} else if kt == KeyEnter {
			text = "\r"
		} else if kt == KeyTab {
			text = "\t"
		}
	}

	return consumed, Event{
		Type: EventKey,
		Key: Key{
			Type:       kt,
			Rune:       r,
			Mod:        mod,
			Action:     action,
			ShiftedKey: shiftedKey,
			BaseKey:    baseKey,
			Text:       text,
		},
	}, true, false
}

func parseThreeInts(s string) (int, int, int, bool) {
	var a, b, c int
	idx := 0

	for idx < len(s) && s[idx] != ';' {
		if s[idx] < '0' || s[idx] > '9' {
			return 0, 0, 0, false
		}
		a = a*10 + int(s[idx]-'0')
		idx++
	}
	if idx >= len(s) || s[idx] != ';' {
		return 0, 0, 0, false
	}
	idx++

	for idx < len(s) && s[idx] != ';' {
		if s[idx] < '0' || s[idx] > '9' {
			return 0, 0, 0, false
		}
		b = b*10 + int(s[idx]-'0')
		idx++
	}
	if idx >= len(s) || s[idx] != ';' {
		return 0, 0, 0, false
	}
	idx++

	for idx < len(s) {
		if s[idx] < '0' || s[idx] > '9' {
			return 0, 0, 0, false
		}
		c = c*10 + int(s[idx]-'0')
		idx++
	}

	return a, b, c, true
}

func findPasteEnd(b []byte) int {
	needle := "\x1b[201~"
	for i := 0; i+len(needle) <= len(b); i++ {
		if string(b[i:i+len(needle)]) == needle {
			return i
		}
	}
	return -1
}

func hasPrefix(b []byte, s string) bool {
	if len(b) < len(s) {
		return false
	}
	return string(b[:len(s)]) == s
}
