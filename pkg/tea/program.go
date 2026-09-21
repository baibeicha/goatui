package tea

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/baibeicha/goatui/pkg/core/renderer"
	"github.com/baibeicha/goatui/pkg/driver"
	"github.com/baibeicha/goatui/pkg/driver/input"
	"github.com/baibeicha/goatui/pkg/spatial"
)

// Model is the interface implemented by applications running in The Elm Architecture.
type Model interface {
	// Init returns the initial command to run when the program begins.
	Init() Cmd
	// Update processes a message and returns the updated model along with an optional command.
	Update(msg Msg) (Model, Cmd)
	// View renders the application state directly into the provided Frame buffer.
	View(frame *Frame)
}

// ProgramOption configures the program runtime.
type ProgramOption func(*Program)

// WithDriver specifies a custom terminal driver (e.g. testkit.MockDriver).
func WithDriver(d driver.Driver) ProgramOption {
	return func(p *Program) {
		p.driver = d
	}
}

// WithCatchCtrlC configures whether Ctrl+C sends KeyMsg or halts the program immediately.
func WithCatchCtrlC(catch bool) ProgramOption {
	return func(p *Program) {
		p.catchCtrlC = catch
	}
}

// Program runs the main event loop for a Model.
type Program struct {
	model      Model
	driver     driver.Driver
	renderer   *renderer.Renderer
	spatial    *spatial.SpatialMap
	frame      *Frame
	msgs       chan Msg
	stopChan   chan struct{}
	stopOnce   sync.Once
	catchCtrlC bool
}

// NewProgram initializes a new TEA program with options.
func NewProgram(initial Model, opts ...ProgramOption) *Program {
	p := &Program{
		model:    initial,
		msgs:     make(chan Msg, 256),
		stopChan: make(chan struct{}),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Send injects a message into the program's event loop from any goroutine.
func (p *Program) Send(msg Msg) {
	if msg == nil {
		return
	}
	select {
	case p.msgs <- msg:
	case <-p.stopChan:
	}
}

// Quit terminates the program event loop.
func (p *Program) Quit() {
	p.Send(QuitMsg{})
}

// Run executes the application event loop until termination.
// In case of an unhandled panic, terminal restore hooks are executed before re-panicking.
func (p *Program) Run(ctx ...context.Context) (Model, error) {
	defer func() {
		if r := recover(); r != nil {
			p.stop()
			// Emergency terminal restore
			driver.TearDown()
			panic(r)
		}
	}()

	var appCtx context.Context
	if len(ctx) > 0 && ctx[0] != nil {
		appCtx = ctx[0]
	} else {
		appCtx = context.Background()
	}

	// Default to OS driver if none provided
	if p.driver == nil {
		drv, err := driver.NewDriver()
		if err != nil {
			return p.model, fmt.Errorf("failed to initialize terminal driver: %w", err)
		}
		p.driver = drv
	}

	if err := p.driver.Init(); err != nil {
		return p.model, fmt.Errorf("driver init failed: %w", err)
	}
	defer func() {
		_ = p.driver.Close()
	}()

	w, h, err := p.driver.Size()
	if err != nil || w <= 0 || h <= 0 {
		w, h = 80, 24
	}

	p.renderer = renderer.NewRenderer(w, h)
	p.spatial = spatial.NewSpatialMap()
	p.frame = &Frame{
		Buffer:  p.renderer.Back(),
		Spatial: p.spatial,
		Arena:   NewFrameArena(256, 4096),
	}

	// Listen for OS interrupt signals
	sigChan := make(chan os.Signal, 2)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	go func() {
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()
		lastW, lastH, _ := p.driver.Size()
		for {
			select {
			case <-appCtx.Done():
				return
			case <-p.stopChan:
				return
			case <-ticker.C:
				w, h, _ := p.driver.Size()
				if w > 0 && h > 0 && (w != lastW || h != lastH) {
					lastW, lastH = w, h
					p.Send(WindowSizeMsg{Width: w, Height: h})
				}
			}
		}
	}()

	// Ingest driver input events
	go p.eventLoop()

	// Initial WindowSizeMsg dispatch so model and screens have proper initial dimensions
	var initSizeCmd Cmd
	p.model, initSizeCmd = p.model.Update(WindowSizeMsg{Width: w, Height: h})
	if initSizeCmd != nil {
		p.execCmd(initSizeCmd)
	}

	// Initial render
	p.renderFrame()

	// Execute model Init()
	if initCmd := p.model.Init(); initCmd != nil {
		p.execCmd(initCmd)
	}

	for {
		select {
		case <-appCtx.Done():
			p.stop()
			return p.model, appCtx.Err()

		case <-sigChan:
			p.stop()
			return p.model, nil

		case <-p.stopChan:
			return p.model, nil

		case msg := <-p.msgs:
			if msg == nil {
				continue
			}

			// Handle QuitMsg
			if _, ok := msg.(QuitMsg); ok {
				p.stop()
				return p.model, nil
			}

			// Handle batch / sequence messages
			if b, ok := msg.(batchMsg); ok {
				execBatch(b, p.Send)
				continue
			}
			if s, ok := msg.(sequenceMsg); ok {
				go func(seq []Cmd) {
					for _, c := range seq {
						if c != nil {
							if res := c(); res != nil {
								p.Send(res)
							}
						}
					}
				}(s)
				continue
			}

			// Handle WindowSizeMsg
			if ws, ok := msg.(WindowSizeMsg); ok {
				p.renderer.Resize(ws.Width, ws.Height)
			}

			// Dispatch to Model
			var cmd Cmd
			p.model, cmd = p.model.Update(msg)
			if cmd != nil {
				p.execCmd(cmd)
			}

			// Redraw frame
			p.renderFrame()
		}
	}
}

func (p *Program) renderFrame() {
	p.renderer.Back().Reset()
	p.spatial.Reset()
	if p.frame.Arena != nil {
		p.frame.Arena.Reset()
	}

	p.model.View(p.frame)
	if p.frame.Buffer != nil {
		p.frame.Buffer.RenderOverlays()
	}

	_ = p.renderer.Render(p.driver.Writer())
	_ = p.driver.Flush()
}

func (p *Program) execCmd(cmd Cmd) {
	if cmd == nil {
		return
	}
	go func() {
		if msg := cmd(); msg != nil {
			p.Send(msg)
		}
	}()
}

func (p *Program) eventLoop() {
	for {
		select {
		case <-p.stopChan:
			return
		case ev, ok := <-p.driver.Events():
			if !ok {
				return
			}
			p.handleDriverEvent(ev)
		}
	}
}

func (p *Program) handleDriverEvent(ev input.Event) {
	switch ev.Type {
	case input.EventKey:
		// Default exit on Ctrl+C unless explicitly caught
		if !p.catchCtrlC && ev.Key.Type == input.KeyRune && ev.Key.Rune == 'c' && ev.Key.Mod.Has(1<<0) {
			p.Quit()
			return
		}
		p.Send(KeyMsg{Key: ev.Key})

	case input.EventMouse:
		// HitMsg is dispatched only on left mouse press (a deliberate click action).
		// This guarantees that 1 physical click = exactly 1 HitMsg, eliminating duplicate triggering on mouse release.
		if ev.Mouse.Action == input.MousePress && ev.Mouse.Button == input.MouseLeft {
			if hit, ok := p.spatial.HitTest(ev.Mouse.X, ev.Mouse.Y); ok {
				p.Send(HitMsg{Target: hit, Mouse: ev.Mouse})
			}
		}
		p.Send(MouseMsg{Mouse: ev.Mouse})

	case input.EventResize:
		p.Send(WindowSizeMsg{Width: ev.Width, Height: ev.Height})

	case input.EventPaste:
		p.Send(PasteMsg{Text: ev.PasteText})

	case input.EventFocus:
		p.Send(FocusMsg{})

	case input.EventBlur:
		p.Send(BlurMsg{})
	}
}

func (p *Program) stop() {
	p.stopOnce.Do(func() {
		close(p.stopChan)
	})
}
