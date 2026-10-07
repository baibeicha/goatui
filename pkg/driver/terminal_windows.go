//go:build windows

package driver

import (
	"bufio"
	"io"
	"os"
	"sync"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/baibeicha/goatui/pkg/driver/input"
	"golang.org/x/sys/windows"
)

type rawConsoleReader struct {
	handle         windows.Handle
	fallback       io.Reader
	utf16Buf       []uint16
	utf8Buf        []byte
	savedSurrogate rune
}

func newRawConsoleReader(h windows.Handle, fallback io.Reader) io.Reader {
	var mode uint32
	if err := windows.GetConsoleMode(h, &mode); err != nil {
		return fallback
	}
	return &rawConsoleReader{
		handle:   h,
		fallback: fallback,
		utf16Buf: make([]uint16, 256),
		utf8Buf:  make([]byte, 0, 1024),
	}
}

func (r *rawConsoleReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	for len(r.utf8Buf) == 0 {
		var nw uint32
		err := windows.ReadConsole(r.handle, &r.utf16Buf[0], uint32(len(r.utf16Buf)), &nw, nil)
		if err != nil {
			if r.fallback != nil {
				return r.fallback.Read(p)
			}
			return 0, err
		}
		if nw == 0 {
			return 0, nil
		}

		u16 := r.utf16Buf[:nw]
		for i := 0; i < len(u16); i++ {
			rn := rune(u16[i])
			if r.savedSurrogate != 0 {
				rn = utf16.DecodeRune(r.savedSurrogate, rn)
				r.savedSurrogate = 0
			} else if utf16.IsSurrogate(rn) {
				if i+1 < len(u16) {
					rn = utf16.DecodeRune(rn, rune(u16[i+1]))
					i++
				} else {
					r.savedSurrogate = rn
					continue
				}
			}
			// Important: Preserve 0x1A (Ctrl-Z) and all other control characters.
			// Go standard library's internal/poll.fd_windows.go strips 0x1A when reading
			// from *os.File console handles, breaking Ctrl+Z in cmd.exe.
			r.utf8Buf = utf8.AppendRune(r.utf8Buf, rn)
		}
	}

	n := copy(p, r.utf8Buf)
	r.utf8Buf = r.utf8Buf[n:]
	return n, nil
}

type windowsDriver struct {
	hStdin      windows.Handle
	hStdout     windows.Handle
	conInFile   *os.File
	conOutFile  *os.File
	inReader    io.Reader
	origInMode  uint32
	origOutMode uint32
	parser      *input.Parser
	events      chan input.Event
	closeOnce   sync.Once
	stopChan    chan struct{}
	outWriter   *bufio.Writer
	cfg         DriverConfig
}

// NewDriver creates an OS terminal driver for Windows.
func NewDriver(opts ...DriverOption) (Driver, error) {
	cfg := DefaultDriverConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	hIn := windows.Handle(os.Stdin.Fd())
	hOut := windows.Handle(os.Stdout.Fd())
	var conInFile *os.File
	var conOutFile *os.File
	inReader := io.Reader(os.Stdin)
	outWriter := bufio.NewWriterSize(os.Stdout, 32768)

	// If stdin is not a console (e.g. piped like "cd d:\tahr | cls | tahr"),
	// connect directly to CONIN$ so keyboard input remains interactive.
	var mode uint32
	if err := windows.GetConsoleMode(hIn, &mode); err != nil {
		if f, err := os.OpenFile("CONIN$", os.O_RDWR, 0); err == nil {
			conInFile = f
			hIn = windows.Handle(f.Fd())
			inReader = newRawConsoleReader(hIn, f)
		}
	} else {
		inReader = newRawConsoleReader(hIn, os.Stdin)
	}

	// If stdout is not a console, connect to CONOUT$
	if err := windows.GetConsoleMode(hOut, &mode); err != nil {
		if f, err := os.OpenFile("CONOUT$", os.O_RDWR, 0); err == nil {
			conOutFile = f
			hOut = windows.Handle(f.Fd())
			outWriter = bufio.NewWriterSize(f, 32768)
		}
	}

	return &windowsDriver{
		hStdin:     hIn,
		hStdout:    hOut,
		conInFile:  conInFile,
		conOutFile: conOutFile,
		inReader:   inReader,
		parser:     input.NewParser(),
		events:     make(chan input.Event, 256),
		stopChan:   make(chan struct{}),
		outWriter:  outWriter,
		cfg:        cfg,
	}, nil
}

func (d *windowsDriver) Init() error {
	// Save initial console modes
	if err := windows.GetConsoleMode(d.hStdin, &d.origInMode); err != nil {
		// Not a real console (e.g. pipes in test or IDE): proceed gracefully
		d.origInMode = 0
	}
	if err := windows.GetConsoleMode(d.hStdout, &d.origOutMode); err != nil {
		d.origOutMode = 0
	}

	// Configure stdout for Virtual Terminal Processing
	outMode := d.origOutMode | windows.ENABLE_PROCESSED_OUTPUT | windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
	_ = windows.SetConsoleMode(d.hStdout, outMode)

	// Configure stdin: disable echo and line input, enable VT input and mouse input
	inMode := uint32(windows.ENABLE_EXTENDED_FLAGS | windows.ENABLE_VIRTUAL_TERMINAL_INPUT | windows.ENABLE_WINDOW_INPUT | windows.ENABLE_MOUSE_INPUT)
	_ = windows.SetConsoleMode(d.hStdin, inMode)

	// Register emergency teardown hook
	RegisterTeardown(func() {
		_ = d.Close()
	})

	// Enter alternate screen, clear screen, hide cursor, disable auto-wrap, enable SGR-1006 mouse, focus tracking & bracketed paste
	initSeq := "\x1b[?1049h" + // Enter alternate screen
		"\x1b[2J\x1b[H" + // Clear screen and position cursor at home (1,1)
		"\x1b[?25l" + // Hide cursor
		"\x1b[?7l" + // Disable line auto-wrap (DECAWM)
		"\x1b[?1000h\x1b[?1003h\x1b[?1006h" + // SGR-1006 Mouse (1003h any-event tracking for hover)
		"\x1b[?1004h" + // Focus tracking
		"\x1b[?2004h" // Bracketed paste

	if !d.cfg.DisableKitty {
		// Push Kitty Keyboard Protocol flags onto the alternate screen stack and query status
		initSeq += input.KittyPushFlags(d.cfg.KittyFlags) + input.KittyQuery()
	}

	_, _ = d.outWriter.WriteString(initSeq)
	_ = d.outWriter.Flush()

	// Start asynchronous reader
	go d.readLoop()

	return nil
}

func (d *windowsDriver) Close() error {
	d.closeOnce.Do(func() {
		close(d.stopChan)

		// Restore normal screen, show cursor, re-enable auto-wrap, disable mouse, disable focus, disable sync
		exitSeq := "\x1b[?1006l\x1b[?1003l\x1b[?1002l\x1b[?1000l" +
			"\x1b[?1004l\x1b[?2026l" +
			"\x1b[?2004l" +
			"\x1b[?7h" // Re-enable auto-wrap

		if !d.cfg.DisableKitty {
			exitSeq += input.KittyDisable()
		}

		exitSeq += "\x1b[?1049l" + // Exit alternate screen
			"\x1b[?25h\x1b[0m" // Show cursor and reset styles

		_, _ = d.outWriter.WriteString(exitSeq)
		_ = d.outWriter.Flush()

		// Restore original console modes
		if d.origInMode != 0 {
			_ = windows.SetConsoleMode(d.hStdin, d.origInMode)
		}
		if d.origOutMode != 0 {
			_ = windows.SetConsoleMode(d.hStdout, d.origOutMode)
		}
		if d.conInFile != nil {
			_ = d.conInFile.Close()
		}
		if d.conOutFile != nil {
			_ = d.conOutFile.Close()
		}
	})
	return nil
}

func (d *windowsDriver) Size() (int, int, error) {
	var csbi windows.ConsoleScreenBufferInfo
	if err := windows.GetConsoleScreenBufferInfo(d.hStdout, &csbi); err == nil {
		w := int(csbi.Window.Right - csbi.Window.Left + 1)
		h := int(csbi.Window.Bottom - csbi.Window.Top + 1)
		if w > 0 && h > 0 {
			return w, h, nil
		}
	}
	// Fallback to standard 80x24 terminal dimensions
	return 80, 24, nil
}

func (d *windowsDriver) Events() <-chan input.Event {
	return d.events
}

func (d *windowsDriver) Writer() io.Writer {
	return d.outWriter
}

func (d *windowsDriver) Flush() error {
	return d.outWriter.Flush()
}

func (d *windowsDriver) readLoop() {
	buf := make([]byte, 1024)
	for {
		select {
		case <-d.stopChan:
			return
		default:
			n, err := d.inReader.Read(buf)
			if err != nil {
				// If reading from pipe returned EOF, switch to CONIN$ if not already done
				if d.conInFile == nil {
					if conIn, ferr := os.OpenFile("CONIN$", os.O_RDWR, 0); ferr == nil {
						d.conInFile = conIn
						d.hStdin = windows.Handle(conIn.Fd())
						inMode := uint32(windows.ENABLE_EXTENDED_FLAGS | windows.ENABLE_VIRTUAL_TERMINAL_INPUT | windows.ENABLE_WINDOW_INPUT | windows.ENABLE_MOUSE_INPUT)
						_ = windows.SetConsoleMode(d.hStdin, inMode)
						d.inReader = newRawConsoleReader(d.hStdin, conIn)
						continue
					}
				}
				return
			}
			if n == 0 {
				continue
			}
			d.parser.Parse(buf[:n], func(ev input.Event) {
				select {
				case d.events <- ev:
				case <-d.stopChan:
				}
			})
		}
	}
}
