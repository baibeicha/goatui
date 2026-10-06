package terminal

import (
	"io"
	"os"
	"runtime"
	"unicode/utf8"

	"github.com/baibeicha/goatui/pkg/driver/input"
)

// PTYSession represents an active interactive terminal pseudo-console session.
type PTYSession interface {
	io.Reader
	io.Writer
	io.Closer
	Resize(cols, rows int) error
	Wait() (*os.ProcessState, error)
	Process() *os.Process
}

// KeyToVT translates a GoatUI keyboard event into standard VT / xterm escape byte sequences.
func KeyToVT(k input.Key) []byte {
	// Ctrl combinations
	if k.HasCtrl() {
		r := k.Rune
		if r == 0 {
			r = k.BaseKey
		}
		// On Windows, Ctrl+Z (0x1A) sends EOF to PowerShell/cmd stdin, causing terminal freeze.
		// Never send 0x1A to terminal on Windows.
		if (r == 'z' || r == 'Z' || r == 'я' || r == 'Я' || r == 26) && runtime.GOOS == "windows" {
			return nil
		}
		if r >= 'a' && r <= 'z' {
			return []byte{byte(r - 'a' + 1)}
		}
		if r >= 'A' && r <= 'Z' {
			return []byte{byte(r - 'A' + 1)}
		}
		switch r {
		case ' ', '@':
			return []byte{0x00}
		case '[':
			return []byte{0x1b}
		case '\\':
			return []byte{0x1c}
		case ']':
			return []byte{0x1d}
		case '^':
			return []byte{0x1e}
		case '_':
			return []byte{0x1f}
		}
	}

	switch k.Type {
	case input.KeyEnter:
		return []byte("\r")
	case input.KeyBackspace:
		return []byte("\x08")
	case input.KeyTab:
		if k.HasShift() {
			return []byte("\x1b[Z")
		}
		return []byte("\t")
	case input.KeyBacktab:
		return []byte("\x1b[Z")
	case input.KeyEsc:
		return []byte("\x1b")
	case input.KeyUp:
		if k.HasCtrl() {
			return []byte("\x1b[1;5A")
		}
		if k.HasShift() {
			return []byte("\x1b[1;2A")
		}
		return []byte("\x1b[A")
	case input.KeyDown:
		if k.HasCtrl() {
			return []byte("\x1b[1;5B")
		}
		if k.HasShift() {
			return []byte("\x1b[1;2B")
		}
		return []byte("\x1b[B")
	case input.KeyRight:
		if k.HasCtrl() {
			return []byte("\x1b[1;5C")
		}
		if k.HasShift() {
			return []byte("\x1b[1;2C")
		}
		return []byte("\x1b[C")
	case input.KeyLeft:
		if k.HasCtrl() {
			return []byte("\x1b[1;5D")
		}
		if k.HasShift() {
			return []byte("\x1b[1;2D")
		}
		return []byte("\x1b[D")
	case input.KeyHome:
		return []byte("\x1b[H")
	case input.KeyEnd:
		return []byte("\x1b[F")
	case input.KeyInsert:
		return []byte("\x1b[2~")
	case input.KeyDelete:
		return []byte("\x1b[3~")
	case input.KeyPgUp, input.KeyKpPageUp:
		return []byte("\x1b[5~")
	case input.KeyPgDown, input.KeyKpPageDown:
		return []byte("\x1b[6~")
	case input.KeyF1:
		return []byte("\x1bOP")
	case input.KeyF2:
		return []byte("\x1bOQ")
	case input.KeyF3:
		return []byte("\x1bOR")
	case input.KeyF4:
		return []byte("\x1bOS")
	case input.KeyF5:
		return []byte("\x1b[15~")
	case input.KeyF6:
		return []byte("\x1b[17~")
	case input.KeyF7:
		return []byte("\x1b[18~")
	case input.KeyF8:
		return []byte("\x1b[19~")
	case input.KeyF9:
		return []byte("\x1b[20~")
	case input.KeyF10:
		return []byte("\x1b[21~")
	case input.KeyF11:
		return []byte("\x1b[23~")
	case input.KeyF12:
		return []byte("\x1b[24~")
	}

	if k.Rune > 0 {
		buf := make([]byte, 4)
		n := utf8.EncodeRune(buf, k.Rune)
		return buf[:n]
	}

	return nil
}
