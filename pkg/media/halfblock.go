package media

import (
	"image"
	"image/color"
	"math"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
)

// ScaleMode dictates how an image is resized to fit terminal rectangle constraints.
type ScaleMode int

const (
	// ScaleFit scales the image to fit within bounds while maintaining aspect ratio (letterboxed).
	ScaleFit ScaleMode = iota
	// ScaleFill scales the image to fill the entire bounds, cropping excess dimensions.
	ScaleFill
	// ScaleStretch resizes the image to precisely match target width and height without maintaining ratio.
	ScaleStretch
)

// HalfBlockRenderer renders images into terminal half-blocks.
type HalfBlockRenderer struct {
	FontAspectRatio float64
}

// NewHalfBlockRenderer creates a new HalfBlockRenderer with default settings.
func NewHalfBlockRenderer() *HalfBlockRenderer {
	return &HalfBlockRenderer{FontAspectRatio: 0.50}
}

// RenderHalfBlock renders an image using the default HalfBlockRenderer.
// Preserved for backward compatibility.
func RenderHalfBlock(buf *buffer.Buffer, area buffer.Rect, img image.Image, mode ScaleMode) {
	renderer := NewHalfBlockRenderer()
	renderer.DrawImage(buf, area, img, mode)
}

// sampleAreaAvg returns the average RGBA color of all source pixels mapping to destination (dstX, dstY).
func sampleAreaAvg(img image.Image, srcX0, srcY0, srcX1, srcY1 float64) (uint8, uint8, uint8) {
	bounds := img.Bounds()
	x0 := max(int(srcX0), bounds.Min.X)
	y0 := max(int(srcY0), bounds.Min.Y)
	x1 := min(int(math.Ceil(srcX1)), bounds.Max.X)
	y1 := min(int(math.Ceil(srcY1)), bounds.Max.Y)

	if x0 >= x1 || y0 >= y1 {
		r, g, b, _ := img.At(max(bounds.Min.X, min(int(srcX0), bounds.Max.X-1)), max(bounds.Min.Y, min(int(srcY0), bounds.Max.Y-1))).RGBA()
		return uint8(r >> 8), uint8(g >> 8), uint8(b >> 8)
	}

	var rSum, gSum, bSum float64
	count := 0
	for py := y0; py < y1; py++ {
		for px := x0; px < x1; px++ {
			r, g, b, _ := img.At(px, py).RGBA()
			rSum += float64(r >> 8)
			gSum += float64(g >> 8)
			bSum += float64(b >> 8)
			count++
		}
	}
	if count == 0 {
		return 0, 0, 0
	}
	return uint8(rSum / float64(count)), uint8(gSum / float64(count)), uint8(bSum / float64(count))
}

// sampleBilinear returns bilinearly interpolated color at fractional source coordinates.
func sampleBilinear(img image.Image, sx, sy float64) (uint8, uint8, uint8) {
	bounds := img.Bounds()
	x0 := int(math.Floor(sx))
	y0 := int(math.Floor(sy))
	x1 := x0 + 1
	y1 := y0 + 1

	fx := sx - float64(x0)
	fy := sy - float64(y0)

	clampX := func(x int) int { return max(bounds.Min.X, min(x, bounds.Max.X-1)) }
	clampY := func(y int) int { return max(bounds.Min.Y, min(y, bounds.Max.Y-1)) }

	getC := func(x, y int) (float64, float64, float64) {
		r, g, b, _ := img.At(clampX(x), clampY(y)).RGBA()
		return float64(r >> 8), float64(g >> 8), float64(b >> 8)
	}

	r00, g00, b00 := getC(x0, y0)
	r10, g10, b10 := getC(x1, y0)
	r01, g01, b01 := getC(x0, y1)
	r11, g11, b11 := getC(x1, y1)

	lerp := func(a, b, t float64) float64 { return a*(1-t) + b*t }

	r := lerp(lerp(r00, r10, fx), lerp(r01, r11, fx), fy)
	g := lerp(lerp(g00, g10, fx), lerp(g01, g11, fx), fy)
	b := lerp(lerp(b00, b10, fx), lerp(b01, b11, fx), fy)

	return uint8(max(0, min(255, r))), uint8(max(0, min(255, g))), uint8(max(0, min(255, b)))
}

// DrawImage renders an image.Image into the destination buffer area using 24-bit TrueColor
// half-blocks (▀, U+2580). Each terminal cell displays two vertical pixels (top=Fg, bottom=Bg).
func (r *HalfBlockRenderer) DrawImage(buf *buffer.Buffer, area buffer.Rect, img image.Image, mode ScaleMode) {
	if buf == nil || img == nil || area.IsEmpty() {
		return
	}

	bounds := img.Bounds()
	imgW := bounds.Dx()
	imgH := bounds.Dy()
	if imgW <= 0 || imgH <= 0 {
		return
	}

	aspect := r.FontAspectRatio
	if aspect <= 0 {
		aspect = 0.50
	}
	canvasW := area.Width
	canvasH := int(float64(area.Height) / aspect)

	var dstW, dstH int
	var startX, startY int

	switch mode {
	case ScaleStretch:
		dstW = canvasW
		dstH = canvasH
		startX = 0
		startY = 0

	case ScaleFit:
		scaleX := float64(canvasW) / float64(imgW)
		scaleY := float64(canvasH) / float64(imgH)
		scale := min(scaleX, scaleY)

		dstW = int(math.Round(float64(imgW) * scale))
		dstH = int(math.Round(float64(imgH) * scale))
		if dstW <= 0 {
			dstW = 1
		}
		if dstH <= 0 {
			dstH = 1
		}

		startX = (canvasW - dstW) / 2
		startY = (canvasH - dstH) / 2

	case ScaleFill:
		scaleX := float64(canvasW) / float64(imgW)
		scaleY := float64(canvasH) / float64(imgH)
		scale := max(scaleX, scaleY)

		dstW = int(math.Round(float64(imgW) * scale))
		dstH = int(math.Round(float64(imgH) * scale))
		if dstW <= 0 {
			dstW = 1
		}
		if dstH <= 0 {
			dstH = 1
		}

		startX = (canvasW - dstW) / 2
		startY = (canvasH - dstH) / 2
	}

	scaleX := float64(imgW) / float64(dstW)
	scaleY := float64(imgH) / float64(dstH)

	samplePixel := func(x, y int) (uint8, uint8, uint8, bool) {
		if x < 0 || x >= dstW || y < 0 || y >= dstH {
			return 0, 0, 0, false
		}
		
		if scaleX < 1.0 || scaleY < 1.0 {
			// Downscaling: use area average
			srcX0 := float64(bounds.Min.X) + float64(x)*scaleX
			srcY0 := float64(bounds.Min.Y) + float64(y)*scaleY
			srcX1 := float64(bounds.Min.X) + float64(x+1)*scaleX
			srcY1 := float64(bounds.Min.Y) + float64(y+1)*scaleY
			r, g, b := sampleAreaAvg(img, srcX0, srcY0, srcX1, srcY1)
			return r, g, b, true
		} else {
			// Upscaling: use bilinear with center alignment
			srcX := float64(bounds.Min.X) + (float64(x)+0.5)*scaleX - 0.5
			srcY := float64(bounds.Min.Y) + (float64(y)+0.5)*scaleY - 0.5
			r, g, b := sampleBilinear(img, srcX, srcY)
			return r, g, b, true
		}
	}

	for row := 0; row < area.Height; row++ {
		screenY := area.Y + row
		// Map terminal row to two canvas rows based on aspect ratio
		canvasYTop := int(float64(row) / aspect) - startY
		canvasYBot := int((float64(row) + 0.5) / aspect) - startY

		for col := 0; col < area.Width; col++ {
			screenX := area.X + col
			px := col - startX

			rTop, gTop, bTop, hasTop := samplePixel(px, canvasYTop)
			rBot, gBot, bBot, hasBot := samplePixel(px, canvasYBot)

			if !hasTop && !hasBot {
				buf.Set(screenX, screenY, cell.Cell{
					Rune:     ' ',
					Width:    1,
					FgType:   cell.ColorDefault,
					BgType:   cell.ColorDefault,
					Modifier: cell.AttrNone,
				})
				continue
			}

			fg := cell.RGB(rTop, gTop, bTop)
			bg := cell.RGB(rBot, gBot, bBot)

			if !hasTop && hasBot {
				buf.Set(screenX, screenY, cell.Cell{
					Rune:     '▄',
					Width:    1,
					FgType:   cell.ColorRGB,
					Fg:       bg.Value,
					BgType:   cell.ColorDefault,
					Modifier: cell.AttrNone,
				})
				continue
			}

			if hasTop && !hasBot {
				buf.Set(screenX, screenY, cell.Cell{
					Rune:     '▀',
					Width:    1,
					FgType:   cell.ColorRGB,
					Fg:       fg.Value,
					BgType:   cell.ColorDefault,
					Modifier: cell.AttrNone,
				})
				continue
			}

			buf.Set(screenX, screenY, cell.Cell{
				Rune:     '▀',
				Width:    1,
				FgType:   cell.ColorRGB,
				BgType:   cell.ColorRGB,
				Fg:       fg.Value,
				Bg:       bg.Value,
				Modifier: cell.AttrNone,
			})
		}
	}
}

// ConvertToRGBA ensures an image is in fast *image.RGBA format.
func ConvertToRGBA(src image.Image) *image.RGBA {
	if rgba, ok := src.(*image.RGBA); ok {
		return rgba
	}
	b := src.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			c := color.RGBAModel.Convert(src.At(x, y)).(color.RGBA)
			dst.SetRGBA(x-b.Min.X, y-b.Min.Y, c)
		}
	}
	return dst
}
