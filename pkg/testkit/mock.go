package testkit

import (
	"bytes"
	"io"
	"sync"

	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/driver/input"
)

// MockDriver is an in-memory headless terminal driver for automated testing and CI.
type MockDriver struct {
	mu     sync.Mutex
	width  int
	height int
	events chan input.Event
	out    bytes.Buffer
	closed bool
}

// NewMockDriver creates a headless mock driver with the specified dimensions.
func NewMockDriver(width, height int) *MockDriver {
	return &MockDriver{
		width:  width,
		height: height,
		events: make(chan input.Event, 512),
	}
}

func (m *MockDriver) Init() error {
	return nil
}

func (m *MockDriver) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func (m *MockDriver) Size() (int, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.width, m.height, nil
}

func (m *MockDriver) Events() <-chan input.Event {
	return m.events
}

func (m *MockDriver) Writer() io.Writer {
	return &mockWriter{driver: m}
}

func (m *MockDriver) Flush() error {
	return nil
}

// SendKey pushes a Key event into the mock driver's event queue.
func (m *MockDriver) SendKey(kt input.KeyType, r rune, mod cell.Modifier) {
	m.events <- input.Event{
		Type: input.EventKey,
		Key: input.Key{
			Type: kt,
			Rune: r,
			Mod:  mod,
		},
	}
}

// SendRune pushes a single rune key event.
func (m *MockDriver) SendRune(r rune) {
	m.SendKey(input.KeyRune, r, cell.AttrNone)
}

// SendText pushes each character of s as a rune event.
func (m *MockDriver) SendText(s string) {
	for _, r := range s {
		m.SendRune(r)
	}
}

// SendMouse pushes a mouse event into the queue.
func (m *MockDriver) SendMouse(x, y int, btn input.MouseButton, act input.MouseAction, mod cell.Modifier) {
	m.events <- input.Event{
		Type: input.EventMouse,
		Mouse: input.Mouse{
			X:      x,
			Y:      y,
			Button: btn,
			Action: act,
			Mod:    mod,
		},
	}
}

// SendResize pushes a resize event and updates the driver's internal dimensions.
func (m *MockDriver) SendResize(width, height int) {
	m.mu.Lock()
	m.width = width
	m.height = height
	m.mu.Unlock()

	m.events <- input.Event{
		Type:   input.EventResize,
		Width:  width,
		Height: height,
	}
}

// OutBytes returns all bytes written to the mock driver.
func (m *MockDriver) OutBytes() []byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]byte(nil), m.out.Bytes()...)
}

// OutString returns all written output as a string.
func (m *MockDriver) OutString() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.out.String()
}

// ClearOut resets the output buffer.
func (m *MockDriver) ClearOut() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.out.Reset()
}

type mockWriter struct {
	driver *MockDriver
}

func (w *mockWriter) Write(p []byte) (int, error) {
	w.driver.mu.Lock()
	defer w.driver.mu.Unlock()
	return w.driver.out.Write(p)
}
