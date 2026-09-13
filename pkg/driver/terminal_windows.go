//go:build windows

package driver

import (
	"bufio"
	"io"
	"os"
	"sync"

	"github.com/baibeicha/goatui/pkg/driver/input"
	"golang.org/x/sys/windows"
)

type windowsDriver struct {
	hStdin     windows.Handle
	hStdout    windows.Handle
	origInMode uint32
	origOutMode uint32
	parser     *input.Parser
	events     chan input.Event
	closeOnce  sync.Once
	stopChan   chan struct{}
	outWriter  *bufio.Writer
}

// NewDriver creates an OS terminal driver for Windows.
func NewDriver() (Driver, error) {
	hIn := windows.Handle(os.Stdin.Fd())
	hOut := windows.Handle(os.Stdout.Fd())

	return &windowsDriver{
		hStdin:    hIn,
		hStdout:   hOut,
		parser:    input.NewParser(),
		events:    make(chan input.Event, 256),
		stopChan:  make(chan struct{}),
		outWriter: bufio.NewWriterSize(os.Stdout, 32768),
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

	// Enter alternate screen, hide cursor, disable auto-wrap, enable SGR-1006 mouse, focus tracking & bracketed paste
	initSeq := "\x1b[?1049h" + // Enter alternate screen
		"\x1b[?25l" + // Hide cursor
		"\x1b[?7l" + // Disable line auto-wrap (DECAWM)
		"\x1b[?1000h\x1b[?1002h\x1b[?1006h" + // SGR-1006 Mouse
		"\x1b[?1004h" + // Focus tracking
		"\x1b[?2004h" // Bracketed paste
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
		exitSeq := "\x1b[?1006l\x1b[?1002l\x1b[?1000l" +
			"\x1b[?1004l\x1b[?2026l" +
			"\x1b[?2004l" +
			"\x1b[?7h" + // Re-enable auto-wrap
			"\x1b[?1049l" +
			"\x1b[?25h\x1b[0m"
		_, _ = d.outWriter.WriteString(exitSeq)
		_ = d.outWriter.Flush()

		// Restore original console modes
		if d.origInMode != 0 {
			_ = windows.SetConsoleMode(d.hStdin, d.origInMode)
		}
		if d.origOutMode != 0 {
			_ = windows.SetConsoleMode(d.hStdout, d.origOutMode)
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
			n, err := os.Stdin.Read(buf)
			if err != nil || n == 0 {
				return
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
