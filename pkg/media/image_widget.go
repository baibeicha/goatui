package media

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"sync"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
)

// ImageWidget displays images within terminal boundaries using the highest capability graphics protocol available.
type ImageWidget struct {
	mu        sync.RWMutex
	img       image.Image
	scaleMode ScaleMode
	proto     Protocol
	filePath  string
	err       error
}

// NewImageWidget initializes an empty ImageWidget with ScaleFit mode.
func NewImageWidget() *ImageWidget {
	return &ImageWidget{
		scaleMode: ScaleFit,
		proto:     CurrentProtocol(),
	}
}

// NewImageWidgetFromFile loads an image from disk and initializes the widget.
func NewImageWidgetFromFile(path string) (*ImageWidget, error) {
	w := NewImageWidget()
	if err := w.LoadFile(path); err != nil {
		return nil, err
	}
	return w, nil
}

// LoadFile reads and decodes an image file (PNG, JPEG, GIF) from disk.
func (w *ImageWidget) LoadFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		w.mu.Lock()
		w.err = err
		w.mu.Unlock()
		return fmt.Errorf("open image %s: %w", path, err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		w.mu.Lock()
		w.err = err
		w.mu.Unlock()
		return fmt.Errorf("decode image %s: %w", path, err)
	}

	w.mu.Lock()
	defer w.mu.Unlock()
	w.img = img
	w.filePath = path
	w.err = nil
	return nil
}

// SetImage assigns an in-memory image.Image to the widget.
func (w *ImageWidget) SetImage(img image.Image) *ImageWidget {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.img = img
	w.err = nil
	return w
}

// Image returns the current image.Image, or nil if none loaded.
func (w *ImageWidget) Image() image.Image {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.img
}

// SetScaleMode updates the scaling mode (ScaleFit, ScaleFill, ScaleStretch).
func (w *ImageWidget) SetScaleMode(mode ScaleMode) *ImageWidget {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.scaleMode = mode
	return w
}

// ScaleMode returns the currently configured scaling mode.
func (w *ImageWidget) ScaleMode() ScaleMode {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.scaleMode
}

// SetProtocol explicitly sets the rendering protocol to use for this widget.
func (w *ImageWidget) SetProtocol(proto Protocol) *ImageWidget {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.proto = proto
	return w
}

// Draw renders the image into the target buffer rectangle.
func (w *ImageWidget) Draw(buf *buffer.Buffer, area buffer.Rect) {
	w.mu.RLock()
	img := w.img
	mode := w.scaleMode
	proto := w.proto
	err := w.err
	w.mu.RUnlock()

	if area.IsEmpty() {
		return
	}

	if err != nil {
		buf.SetString(area.X, area.Y, fmt.Sprintf("[Image Error: %v]", err), cell.ColorHex("#FF5555"), cell.DefaultColor(), cell.AttrNone)
		return
	}

	if img == nil {
		return
	}

	if proto == ProtoBraille {
		br := NewBrailleRenderer()
		br.Dithering = true
		br.DrawImage(buf, area, img)
	} else {
		RenderHalfBlock(buf, area, img, mode)
	}
}
