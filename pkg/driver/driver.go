package driver

import (
	"io"
	"os"
	"sync"

	"github.com/baibeicha/goatui/pkg/driver/input"
)

// Driver defines the contract for OS terminal interactions and event streams.
type Driver interface {
	// Init configures the terminal into raw mode and enables alternate screen buffer.
	Init() error
	// Close restores the terminal to original state and shows the cursor.
	Close() error
	// Size returns the current terminal dimensions (columns, rows).
	Size() (width, height int, err error)
	// Events returns a read-only channel of parsed input events.
	Events() <-chan input.Event
	// Writer returns the output stream where rendered frames are flushed.
	Writer() io.Writer
	// Flush commits all buffered output bytes directly to the OS terminal stream.
	Flush() error
}

var (
	teardownMu    sync.Mutex
	teardownHooks []func()
)

// RegisterTeardown registers a cleanup hook guaranteed to execute on emergency shutdown.
func RegisterTeardown(fn func()) {
	teardownMu.Lock()
	defer teardownMu.Unlock()
	teardownHooks = append(teardownHooks, fn)
}

// TearDown executes all registered cleanup hooks and flushes standard terminal restore sequences.
func TearDown() {
	teardownMu.Lock()
	hooks := make([]func(), len(teardownHooks))
	copy(hooks, teardownHooks)
	teardownHooks = nil
	teardownMu.Unlock()

	// Execute hooks in reverse order
	for i := len(hooks) - 1; i >= 0; i-- {
		hooks[i]()
	}

	// Always emit full terminal restore escape sequence to stdout
	_, _ = os.Stdout.WriteString(
		"\x1b[0m" + // Reset styles
			"\x1b[?1006l\x1b[?1003l\x1b[?1002l\x1b[?1000l" + // Disable mouse
			"\x1b[?2004l" + // Disable bracketed paste
			"\x1b[?1004l" + // Disable focus reporting
			"\x1b[?7h" + // Re-enable line auto-wrap
			"\x1b[?1049l" + // Exit alternate screen
			"\x1b[?25h", // Show cursor
	)
}
