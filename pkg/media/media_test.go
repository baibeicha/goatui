package media

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"testing"
	"time"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
)

func TestDetectProtocol(t *testing.T) {
	proto := DetectProtocol()
	if proto < ProtoKitty || proto > ProtoBraille {
		t.Errorf("Unexpected protocol enum: %d", proto)
	}

	SetProtocol(ProtoHalfBlock)
	if CurrentProtocol() != ProtoHalfBlock {
		t.Errorf("SetProtocol failed: expected ProtoHalfBlock, got %v", CurrentProtocol())
	}
}

func TestRenderHalfBlock(t *testing.T) {
	// Create a 2x4 RGBA image
	// Row 0: Red, Green
	// Row 1: Blue, Yellow
	// Row 2: Cyan, Magenta
	// Row 3: White, Black
	img := image.NewRGBA(image.Rect(0, 0, 2, 4))
	img.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	img.Set(1, 0, color.RGBA{R: 0, G: 255, B: 0, A: 255})
	img.Set(0, 1, color.RGBA{R: 0, G: 0, B: 255, A: 255})
	img.Set(1, 1, color.RGBA{R: 255, G: 255, B: 0, A: 255})
	img.Set(0, 2, color.RGBA{R: 0, G: 255, B: 255, A: 255})
	img.Set(1, 2, color.RGBA{R: 255, G: 0, B: 255, A: 255})
	img.Set(0, 3, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	img.Set(1, 3, color.RGBA{R: 0, G: 0, B: 0, A: 255})

	// Render into a 2x2 terminal cell buffer (2 columns, 2 rows = 2 cols, 4 subpixels)
	buf := buffer.NewBuffer(2, 2)
	RenderHalfBlock(buf, buffer.NewRect(0, 0, 2, 2), img, ScaleStretch)

	// Check top-left cell (x=0, y=0)
	// Subpixel top should be Red, bottom should be Blue
	c00 := buf.Cell(0, 0)
	if c00 == nil || c00.Rune != '▀' {
		t.Fatalf("Expected rune '▀', got %+v", c00)
	}
	if c00.FgType != cell.ColorRGB || c00.BgType != cell.ColorRGB {
		t.Errorf("Expected ColorRGB for both Fg and Bg, got Fg=%d, Bg=%d", c00.FgType, c00.BgType)
	}

	rTop := uint8((c00.Fg >> 16) & 0xFF)
	bBot := uint8(c00.Bg & 0xFF)
	if rTop != 255 {
		t.Errorf("Expected top red=255, got %d", rTop)
	}
	if bBot != 255 {
		t.Errorf("Expected bottom blue=255, got %d", bBot)
	}
}

func TestImageWidget(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	w := NewImageWidget().SetImage(img).SetScaleMode(ScaleFit)

	buf := buffer.NewBuffer(20, 10)
	w.Draw(buf, buffer.NewRect(0, 0, 20, 10))

	if w.Image() == nil {
		t.Errorf("Expected non-nil image in ImageWidget")
	}
	if w.ScaleMode() != ScaleFit {
		t.Errorf("Expected ScaleFit mode")
	}
}

func TestVideoPlayerWidget(t *testing.T) {
	// Create a minimal 2-frame animated GIF in memory
	pal := []color.Color{color.Black, color.White}
	f1 := image.NewPaletted(image.Rect(0, 0, 2, 2), pal)
	f2 := image.NewPaletted(image.Rect(0, 0, 2, 2), pal)
	f2.Set(0, 0, color.White)

	g := &gif.GIF{
		Image: []*image.Paletted{f1, f2},
		Delay: []int{5, 5}, // 50ms per frame
	}

	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, g); err != nil {
		t.Fatalf("Failed to encode test GIF: %v", err)
	}

	player := NewVideoPlayer()
	if err := player.LoadGIFReader(&buf, "test.gif"); err != nil {
		t.Fatalf("LoadGIFReader failed: %v", err)
	}

	if player.TotalFrames() != 2 {
		t.Fatalf("Expected 2 frames, got %d", player.TotalFrames())
	}
	if player.CurrentFrame() != 0 {
		t.Errorf("Expected initial frame 0, got %d", player.CurrentFrame())
	}

	// Advance 25ms: should remain at frame 0
	changed := player.Advance(25 * time.Millisecond)
	if changed || player.CurrentFrame() != 0 {
		t.Errorf("Frame should not change before delay expires")
	}

	// Advance 30ms more (55ms total > 50ms): should advance to frame 1
	changed = player.Advance(30 * time.Millisecond)
	if !changed || player.CurrentFrame() != 1 {
		t.Errorf("Expected transition to frame 1, got frame %d, changed=%v", player.CurrentFrame(), changed)
	}

	// Advance 60ms: should loop back to frame 0
	player.Advance(60 * time.Millisecond)
	if player.CurrentFrame() != 0 {
		t.Errorf("Expected loop back to frame 0, got frame %d", player.CurrentFrame())
	}

	// Test NextFrame and PrevFrame
	player.NextFrame()
	if player.CurrentFrame() != 1 {
		t.Errorf("Expected NextFrame to advance to 1, got %d", player.CurrentFrame())
	}
	player.PrevFrame()
	if player.CurrentFrame() != 0 {
		t.Errorf("Expected PrevFrame to go back to 0, got %d", player.CurrentFrame())
	}

	// Test pause and draw
	player.Pause()
	if player.IsPlaying() {
		t.Errorf("Expected player to be paused")
	}

	tBuf := buffer.NewBuffer(10, 5)
	player.Draw(tBuf, buffer.NewRect(0, 0, 10, 5))
}

func TestBrailleRenderer(t *testing.T) {
	// Create a 2x4 image where only top-left pixel (0,0) is white
	img := image.NewRGBA(image.Rect(0, 0, 2, 4))
	img.Set(0, 0, color.RGBA{R: 255, G: 255, B: 255, A: 255})

	buf := buffer.NewBuffer(1, 1)
	br := NewBrailleRenderer()
	br.Threshold = 128
	br.DrawImage(buf, buffer.NewRect(0, 0, 1, 1), img)

	c := buf.Cell(0, 0)
	if c == nil {
		t.Fatalf("Cell (0,0) is nil")
	}
	// Dot at (0,0) is bit 0 -> 0x2800 + 1 = 0x2801 ('⠁')
	if c.Rune != '⠁' {
		t.Errorf("Expected Braille rune '⠁' (0x2801), got %q (0x%X)", c.Rune, c.Rune)
	}
}

func TestHalfBlockRendererAspectRatio(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	img.Set(5, 5, color.RGBA{R: 255, G: 0, B: 0, A: 255})

	buf := buffer.NewBuffer(10, 10)
	hbr := NewHalfBlockRenderer()
	hbr.FontAspectRatio = 0.50
	hbr.DrawImage(buf, buffer.NewRect(0, 0, 10, 10), img, ScaleFit)

	// Buffer must contain at least one half-block cell
	found := false
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			if cell := buf.Cell(x, y); cell != nil && cell.Rune == '▀' {
				found = true
				break
			}
		}
	}
	if !found {
		t.Errorf("Expected halfblock rune '▀' to be rendered")
	}
}
