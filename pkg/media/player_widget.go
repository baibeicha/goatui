package media

import (
	"fmt"
	"image"
	"image/gif"
	"io"
	"os"
	"sync"
	"time"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
)

// VideoPlayerWidget displays and animates frame sequences and animated GIFs at precise frame timings.
type VideoPlayerWidget struct {
	mu sync.RWMutex

	frames []image.Image
	delays []time.Duration

	currentFrame int
	frameElapsed time.Duration
	playing      bool
	loop         bool
	scaleMode    ScaleMode
	proto        Protocol

	filePath string
	err      error
}

// NewVideoPlayer creates an empty VideoPlayerWidget configured to loop by default.
func NewVideoPlayer() *VideoPlayerWidget {
	return &VideoPlayerWidget{
		loop:      true,
		scaleMode: ScaleFit,
		proto:     CurrentProtocol(),
	}
}

// NewVideoPlayerFromGIF loads an animated GIF file from the given path.
func NewVideoPlayerFromGIF(path string) (*VideoPlayerWidget, error) {
	player := NewVideoPlayer()
	if err := player.LoadGIF(path); err != nil {
		return nil, err
	}
	return player, nil
}

// LoadGIF opens and decodes an animated GIF from a file path.
func (p *VideoPlayerWidget) LoadGIF(path string) error {
	f, err := os.Open(path)
	if err != nil {
		p.mu.Lock()
		p.err = err
		p.mu.Unlock()
		return fmt.Errorf("open gif %s: %w", path, err)
	}
	defer f.Close()

	return p.LoadGIFReader(f, path)
}

// LoadGIFReader decodes an animated GIF from any io.Reader.
func (p *VideoPlayerWidget) LoadGIFReader(r io.Reader, label string) error {
	g, err := gif.DecodeAll(r)
	if err != nil {
		p.mu.Lock()
		p.err = err
		p.mu.Unlock()
		return fmt.Errorf("decode gif %s: %w", label, err)
	}

	frames := make([]image.Image, len(g.Image))
	delays := make([]time.Duration, len(g.Image))

	for i, frame := range g.Image {
		frames[i] = frame
		// gif.Delay is in 100ths of a second (10ms units)
		delayMs := g.Delay[i] * 10
		if delayMs <= 0 {
			delayMs = 100 // Default to 10 FPS (100ms) if zero
		}
		delays[i] = time.Duration(delayMs) * time.Millisecond
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.frames = frames
	p.delays = delays
	p.currentFrame = 0
	p.frameElapsed = 0
	p.filePath = label
	p.err = nil
	p.playing = true

	return nil
}

// SetFrames manually sets an image frame sequence with corresponding per-frame durations.
func (p *VideoPlayerWidget) SetFrames(frames []image.Image, delays []time.Duration) *VideoPlayerWidget {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.frames = frames
	p.delays = delays
	p.currentFrame = 0
	p.frameElapsed = 0
	p.err = nil
	return p
}

// Play starts or resumes playback.
func (p *VideoPlayerWidget) Play() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.playing = true
}

// Pause pauses playback at the current frame.
func (p *VideoPlayerWidget) Pause() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.playing = false
}

// Stop pauses playback and rewinds to frame 0.
func (p *VideoPlayerWidget) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.playing = false
	p.currentFrame = 0
	p.frameElapsed = 0
}

// Toggle flips between playing and paused states.
func (p *VideoPlayerWidget) Toggle() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.playing = !p.playing
}

// NextFrame steps forward by one frame.
func (p *VideoPlayerWidget) NextFrame() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.frames) == 0 {
		return
	}
	p.currentFrame = (p.currentFrame + 1) % len(p.frames)
	p.frameElapsed = 0
}

// PrevFrame steps backward by one frame.
func (p *VideoPlayerWidget) PrevFrame() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.frames) == 0 {
		return
	}
	p.currentFrame = (p.currentFrame - 1 + len(p.frames)) % len(p.frames)
	p.frameElapsed = 0
}

// Advance progresses playback time by dt. Returns true if the active frame changed.
func (p *VideoPlayerWidget) Advance(dt time.Duration) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.playing || len(p.frames) <= 1 || dt <= 0 {
		return false
	}

	frameChanged := false
	p.frameElapsed += dt

	targetDelay := time.Millisecond * 100 // Fallback
	if p.currentFrame < len(p.delays) && p.delays[p.currentFrame] > 0 {
		targetDelay = p.delays[p.currentFrame]
	}

	for p.frameElapsed >= targetDelay {
		p.frameElapsed -= targetDelay
		p.currentFrame++
		frameChanged = true

		if p.currentFrame >= len(p.frames) {
			if p.loop {
				p.currentFrame = 0
			} else {
				p.currentFrame = len(p.frames) - 1
				p.playing = false
				break
			}
		}

		if p.currentFrame < len(p.delays) && p.delays[p.currentFrame] > 0 {
			targetDelay = p.delays[p.currentFrame]
		}
	}

	return frameChanged
}

// CurrentFrame returns the 0-indexed position of the active frame.
func (p *VideoPlayerWidget) CurrentFrame() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.currentFrame
}

// TotalFrames returns the total number of frames in the sequence.
func (p *VideoPlayerWidget) TotalFrames() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.frames)
}

// IsPlaying reports whether playback is actively progressing.
func (p *VideoPlayerWidget) IsPlaying() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.playing
}

// SetScaleMode updates the scaling mode for frame rendering.
func (p *VideoPlayerWidget) SetScaleMode(mode ScaleMode) *VideoPlayerWidget {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.scaleMode = mode
	return p
}

// SetProtocol explicitly sets the rendering protocol to use for this widget.
func (p *VideoPlayerWidget) SetProtocol(proto Protocol) *VideoPlayerWidget {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.proto = proto
	return p
}

// Draw renders the active frame into the target buffer rectangle using TrueColor half-blocks.
func (p *VideoPlayerWidget) Draw(buf *buffer.Buffer, area buffer.Rect) {
	p.mu.RLock()
	total := len(p.frames)
	cur := p.currentFrame
	mode := p.scaleMode
	proto := p.proto
	err := p.err
	var curImg image.Image
	if total > 0 && cur >= 0 && cur < total {
		curImg = p.frames[cur]
	}
	p.mu.RUnlock()

	if area.IsEmpty() {
		return
	}

	if err != nil {
		buf.SetString(area.X, area.Y, fmt.Sprintf("[Video Error: %v]", err), cell.ColorHex("#FF5555"), cell.DefaultColor(), cell.AttrNone)
		return
	}

	if curImg == nil {
		buf.SetString(area.X, area.Y, "[No Video Frames Loaded]", cell.ColorHex("#888888"), cell.DefaultColor(), cell.AttrDim)
		return
	}

	if proto == ProtoBraille {
		br := NewBrailleRenderer()
		br.Dithering = true
		br.DrawImage(buf, area, curImg)
	} else {
		RenderHalfBlock(buf, area, curImg, mode)
	}
}
