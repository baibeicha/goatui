package input

import (
	"unicode/utf8"

	"github.com/baibeicha/goatui/pkg/core/cell"
)

// Parser parses raw terminal input bytes into structured events.
type Parser struct {
	buf       []byte
	inPaste   bool
	pasteData []byte
}

// NewParser creates a new input parser.
func NewParser() *Parser {
	return &Parser{
		buf:       make([]byte, 0, 512),
		pasteData: make([]byte, 0, 1024),
	}
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
					Key:  Key{Type: KeyEsc},
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
				consumed, ev, ok, incomplete := parseCSI(p.buf)
				if incomplete {
					// Wait for more bytes
					return
				}
				if ok {
					handler(ev)
					p.buf = p.buf[consumed:]
					continue
				}
			}

			// SS3 sequences (\x1bO...) e.g. F1-F4
			if p.buf[1] == 'O' && len(p.buf) >= 3 {
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
						Type: KeyRune,
						Rune: r,
						Mod:  cell.AttrDim, // Used as Alt modifier representation
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
				Key:  Key{Type: KeySpace, Mod: cell.AttrReverse}, // Ctrl modifier flag
			})
			p.buf = p.buf[1:]
			continue
		case 0x09: // Tab
			handler(Event{Type: EventKey, Key: Key{Type: KeyTab}})
			p.buf = p.buf[1:]
			continue
		case 0x0A, 0x0D: // Enter (LF or CR)
			handler(Event{Type: EventKey, Key: Key{Type: KeyEnter}})
			if b == 0x0D && len(p.buf) > 1 && p.buf[1] == 0x0A {
				p.buf = p.buf[2:]
			} else {
				p.buf = p.buf[1:]
			}
			continue
		case 0x08, 0x7F: // Backspace
			handler(Event{Type: EventKey, Key: Key{Type: KeyBackspace}})
			p.buf = p.buf[1:]
			continue
		}

		// Ctrl+A through Ctrl+Z
		if b >= 1 && b <= 26 && b != 0x09 && b != 0x0A && b != 0x0D {
			handler(Event{
				Type: EventKey,
				Key: Key{
					Type: KeyRune,
					Rune: rune('a' + b - 1),
					Mod:  cell.AttrBold, // Use modifier
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
				Key:  Key{Type: kt, Rune: r},
			})
			p.buf = p.buf[size:]
			continue
		}

		// Invalid byte: skip
		p.buf = p.buf[1:]
	}
}

func parseCSI(b []byte) (consumed int, ev Event, ok bool, incomplete bool) {
	if len(b) < 3 {
		return 0, Event{}, false, true
	}

	// SGR 1006 Mouse: \x1b[<btn;x;yM or \x1b[<btn;x;ym
	if b[2] == '<' {
		return parseSGRMouse(b)
	}

	// Find terminating character
	endIdx := -1
	for i := 2; i < len(b); i++ {
		c := b[i]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || c == '~' {
			endIdx = i
			break
		}
	}
	if endIdx == -1 {
		if len(b) > 32 {
			// Malformed sequence: drop
			return 2, Event{}, false, false
		}
		return 0, Event{}, false, true
	}

	term := b[endIdx]
	seq := string(b[2:endIdx])
	consumed = endIdx + 1

	// Kitty Keyboard: e.g. \x1b[97;5u (Ctrl+a)
	if term == 'u' {
		return parseKittyKeyboard(seq, consumed)
	}

	// Standard functional arrows
	switch term {
	case 'A':
		return consumed, Event{Type: EventKey, Key: Key{Type: KeyUp}}, true, false
	case 'B':
		return consumed, Event{Type: EventKey, Key: Key{Type: KeyDown}}, true, false
	case 'C':
		return consumed, Event{Type: EventKey, Key: Key{Type: KeyRight}}, true, false
	case 'D':
		return consumed, Event{Type: EventKey, Key: Key{Type: KeyLeft}}, true, false
	case 'H':
		return consumed, Event{Type: EventKey, Key: Key{Type: KeyHome}}, true, false
	case 'F':
		return consumed, Event{Type: EventKey, Key: Key{Type: KeyEnd}}, true, false
	case 'Z':
		return consumed, Event{Type: EventKey, Key: Key{Type: KeyBacktab}}, true, false
	}

	// Numeric ~ sequences (e.g. \x1b[3~ = Delete, \x1b[5~ = PgUp)
	if term == '~' {
		switch seq {
		case "1", "7":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyHome}}, true, false
		case "2":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyInsert}}, true, false
		case "3":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyDelete}}, true, false
		case "4", "8":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyEnd}}, true, false
		case "5":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyPgUp}}, true, false
		case "6":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyPgDown}}, true, false
		case "11", "15":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyF5}}, true, false
		case "12", "17":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyF6}}, true, false
		case "13", "18":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyF7}}, true, false
		case "14", "19":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyF8}}, true, false
		case "20":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyF9}}, true, false
		case "21":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyF10}}, true, false
		case "23":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyF11}}, true, false
		case "24":
			return consumed, Event{Type: EventKey, Key: Key{Type: KeyF12}}, true, false
		}
	}

	return consumed, Event{}, false, false
}

func parseSS3(b byte) (Event, bool) {
	switch b {
	case 'P':
		return Event{Type: EventKey, Key: Key{Type: KeyF1}}, true
	case 'Q':
		return Event{Type: EventKey, Key: Key{Type: KeyF2}}, true
	case 'R':
		return Event{Type: EventKey, Key: Key{Type: KeyF3}}, true
	case 'S':
		return Event{Type: EventKey, Key: Key{Type: KeyF4}}, true
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

	mouse := Mouse{
		X: x - 1, // Convert 1-indexed terminal coords to 0-indexed
		Y: y - 1,
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
		if isRelease {
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

func parseKittyKeyboard(seq string, consumed int) (int, Event, bool, bool) {
	code, mod, _, parsed := parseThreeInts(seq)
	if !parsed {
		var err bool
		code, mod, err = parseTwoInts(seq)
		if !err {
			return consumed, Event{}, false, false
		}
	}

	var m cell.Modifier
	if mod > 1 {
		// Kitty modifier: 1=none, 2=Shift, 3=Alt, 4=Shift+Alt, 5=Ctrl...
		if (mod-1)&1 != 0 {
			m.Add(cell.AttrUnderline) // Shift
		}
		if (mod-1)&2 != 0 {
			m.Add(cell.AttrDim) // Alt
		}
		if (mod-1)&4 != 0 {
			m.Add(cell.AttrBold) // Ctrl
		}
	}

	return consumed, Event{
		Type: EventKey,
		Key: Key{
			Type: KeyRune,
			Rune: rune(code),
			Mod:  m,
		},
	}, true, false
}

func parseThreeInts(s string) (int, int, int, bool) {
	var a, b, c int
	n := 0
	idx := 0

	for idx < len(s) && s[idx] != ';' {
		if s[idx] < '0' || s[idx] > '9' {
			return 0, 0, 0, false
		}
		a = a*10 + int(s[idx]-'0')
		idx++
		n++
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

func parseTwoInts(s string) (int, int, bool) {
	var a, b int
	idx := 0

	for idx < len(s) && s[idx] != ';' {
		if s[idx] < '0' || s[idx] > '9' {
			return 0, 0, false
		}
		a = a*10 + int(s[idx]-'0')
		idx++
	}
	if idx >= len(s) || s[idx] != ';' {
		return a, 1, true
	}
	idx++

	for idx < len(s) {
		if s[idx] < '0' || s[idx] > '9' {
			return 0, 0, false
		}
		b = b*10 + int(s[idx]-'0')
		idx++
	}
	return a, b, true
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
