//go:build !windows

package driver

import (
	"bufio"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/baibeicha/goatui/pkg/driver/input"
	"golang.org/x/sys/unix"
)

type unixDriver struct {
	origTermios *unix.Termios
	parser      *input.Parser
	events      chan input.Event
	closeOnce   sync.Once
	stopChan    chan struct{}
	sigChan     chan os.Signal
	outWriter   *bufio.Writer
}

// NewDriver creates an OS terminal driver for POSIX systems (Linux/macOS/BSD).
func NewDriver() (Driver, error) {
	return &unixDriver{
		parser:    input.NewParser(),
		events:    make(chan input.Event, 256),
		stopChan:  make(chan struct{}),
		sigChan:   make(chan os.Signal, 4),
		outWriter: bufio.NewWriterSize(os.Stdout, 32768),
	}, nil
}

func (d *unixDriver) Init() error {
	fd := int(os.Stdin.Fd())
	termios, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	if err == nil {
		d.origTermios = termios
		raw := *termios
		// Disable echo, canonical mode, extended input, signals
		raw.Lflag &^= unix.ECHO | unix.ICANON | unix.IEXTEN | unix.ISIG
		// Disable flow control and CR to NL mapping
		raw.Iflag &^= unix.IXON | unix.ICRNL | unix.BRKINT | unix.INPCK | unix.ISTRIP
		// Set 8-bit chars
		raw.Cflag |= unix.CS8
		// Read at least 1 byte, no timeout
		raw.Cc[unix.VMIN] = 1
		raw.Cc[unix.VTIME] = 0

		_ = unix.IoctlSetTermios(fd, unix.TCSETS, &raw)
	}

	// Register teardown hook
	RegisterTeardown(func() {
		_ = d.Close()
	})

	// Listen for SIGWINCH window resize signals
	signal.Notify(d.sigChan, syscall.SIGWINCH)
	go d.signalLoop()

	// Enter alternate screen, hide cursor, disable auto-wrap, enable mouse & bracketed paste
	initSeq := "\x1b[?1049h" +
		"\x1b[?25l" +
		"\x1b[?7l" + // Disable line auto-wrap (DECAWM)
		"\x1b[?1000h\x1b[?1002h\x1b[?1006h" +
		"\x1b[?2004h"
	_, _ = d.outWriter.WriteString(initSeq)
	_ = d.outWriter.Flush()

	go d.readLoop()

	return nil
}

func (d *unixDriver) Close() error {
	d.closeOnce.Do(func() {
		close(d.stopChan)
		signal.Stop(d.sigChan)

		// Restore normal screen, show cursor, re-enable auto-wrap, disable mouse
		exitSeq := "\x1b[?1006l\x1b[?1002l\x1b[?1000l" +
			"\x1b[?2004l" +
			"\x1b[?7h" + // Re-enable auto-wrap
			"\x1b[?1049l" +
			"\x1b[?25h\x1b[0m"
		_, _ = d.outWriter.WriteString(exitSeq)
		_ = d.outWriter.Flush()

		if d.origTermios != nil {
			_ = unix.IoctlSetTermios(int(os.Stdin.Fd()), unix.TCSETS, d.origTermios)
		}
	})
	return nil
}

func (d *unixDriver) Size() (int, int, error) {
	ws, err := unix.IoctlGetWinsize(int(os.Stdout.Fd()), unix.TIOCGWINSZ)
	if err == nil && ws.Col > 0 && ws.Row > 0 {
		return int(ws.Col), int(ws.Row), nil
	}
	return 80, 24, nil
}

func (d *unixDriver) Events() <-chan input.Event {
	return d.events
}

func (d *unixDriver) Writer() io.Writer {
	return d.outWriter
}

func (d *unixDriver) Flush() error {
	return d.outWriter.Flush()
}

func (d *unixDriver) signalLoop() {
	for {
		select {
		case <-d.stopChan:
			return
		case <-d.sigChan:
			w, h, err := d.Size()
			if err == nil {
				d.events <- input.Event{
					Type:   input.EventResize,
					Width:  w,
					Height: h,
				}
			}
		}
	}
}

func (d *unixDriver) readLoop() {
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
