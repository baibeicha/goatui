package widgets

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/driver/input"
)

func toColor(val uint32) cell.Color {
	return cell.Color{
		Type:  cell.ColorRGB,
		Value: val,
	}
}

// ColorPickerTheme encapsulates color styling for the ColorPickerModal.
type ColorPickerTheme struct {
	Background cell.Color
	Foreground cell.Color
	Accent     cell.Color
	Comment    cell.Color
}

// DefaultColorPickerTheme returns standard dark theme colors for the picker.
func DefaultColorPickerTheme() ColorPickerTheme {
	return ColorPickerTheme{
		Background: cell.RGB(30, 30, 46),
		Foreground: cell.RGB(205, 214, 244),
		Accent:     cell.RGB(137, 180, 250),
		Comment:    cell.RGB(108, 112, 134),
	}
}

// ColorPickerModal provides an interactive 2D radial color wheel and brightness slider.
type ColorPickerModal struct {
	Open       bool
	ColorKey   string
	ColorLabel string
	OrigHex    string
	CurHex     string
	Hue        float64 // 0 .. 360
	Sat        float64 // 0 .. 1.0
	Val        float64 // 0 .. 1.0 (Brightness)
	FocusMode  int     // 0: Wheel, 1: Value slider, 2: Hex text input
	HexInput   string
	OnApply    func(key, hexVal string)
	OnPreview  func(key, hexVal string)
	Radius     int // Wheel radius in characters

	// Localizable labels
	HintTab    string
	HintCancel string
	HintApply  string
}

// NewColorPickerModal initializes a color picker modal.
func NewColorPickerModal() *ColorPickerModal {
	return &ColorPickerModal{
		Open:       false,
		Hue:        210,
		Sat:        0.8,
		Val:        0.95,
		Radius:     6,
		FocusMode:  0,
		HintTab:    "[Tab] Mode",
		HintCancel: "Cancel",
		HintApply:  "Apply",
	}
}

// OpenForColor activates the modal for a specific color key and hex value.
func (cp *ColorPickerModal) OpenForColor(key, label, hexVal string, applyFn func(key, hexVal string)) {
	cp.Open = true
	cp.ColorKey = key
	cp.ColorLabel = label
	cp.OrigHex = hexVal
	cp.CurHex = hexVal
	cp.HexInput = strings.TrimPrefix(hexVal, "#")
	cp.OnApply = applyFn
	cp.FocusMode = 0

	// Parse current hex into HSV
	if c, ok := HexToRGBColor(hexVal); ok {
		h, s, v := RGBToHSV(c.R(), c.G(), c.B())
		cp.Hue = h
		cp.Sat = s
		cp.Val = v
	}
}

// HandleKey handles keyboard interaction inside the Color Picker modal.
func (cp *ColorPickerModal) HandleKey(k input.Key) (handled bool, shouldClose bool) {
	if !cp.Open {
		return false, false
	}

	switch k.Type {
	case input.KeyEsc:
		cp.Open = false
		return true, true

	case input.KeyEnter:
		if cp.FocusMode == 2 && cp.HexInput != "" {
			hStr := "#" + strings.TrimPrefix(cp.HexInput, "#")
			if c, ok := HexToRGBColor(hStr); ok {
				cp.CurHex = hStr
				h, s, v := RGBToHSV(c.R(), c.G(), c.B())
				cp.Hue = h
				cp.Sat = s
				cp.Val = v
			}
		}
		if cp.OnApply != nil {
			cp.OnApply(cp.ColorKey, cp.CurHex)
		}
		cp.Open = false
		return true, true

	case input.KeyTab:
		cp.FocusMode = (cp.FocusMode + 1) % 3
		return true, false

	case input.KeyBacktab:
		cp.FocusMode = (cp.FocusMode + 2) % 3
		return true, false

	case input.KeyLeft:
		if cp.FocusMode == 0 { // Wheel: rotate hue counter-clockwise
			cp.Hue -= 10
			if cp.Hue < 0 {
				cp.Hue += 360
			}
			cp.syncHex()
		} else if cp.FocusMode == 1 { // Value slider: decrease brightness
			cp.Val = math.Max(0.0, cp.Val-0.05)
			cp.syncHex()
		}
		return true, false

	case input.KeyRight:
		if cp.FocusMode == 0 { // Wheel: rotate hue clockwise
			cp.Hue += 10
			if cp.Hue >= 360 {
				cp.Hue -= 360
			}
			cp.syncHex()
		} else if cp.FocusMode == 1 { // Value slider: increase brightness
			cp.Val = math.Min(1.0, cp.Val+0.05)
			cp.syncHex()
		}
		return true, false

	case input.KeyUp:
		if cp.FocusMode == 0 { // Wheel: increase saturation
			cp.Sat = math.Min(1.0, cp.Sat+0.08)
			cp.syncHex()
		} else if cp.FocusMode == 1 {
			cp.FocusMode = 0
		}
		return true, false

	case input.KeyDown:
		if cp.FocusMode == 0 { // Wheel: decrease saturation
			cp.Sat = math.Max(0.0, cp.Sat-0.08)
			cp.syncHex()
		} else if cp.FocusMode == 0 && cp.Sat <= 0.05 {
			cp.FocusMode = 1
		}
		return true, false

	case input.KeyBackspace:
		if cp.FocusMode == 2 && len(cp.HexInput) > 0 {
			r := []rune(cp.HexInput)
			cp.HexInput = string(r[:len(r)-1])
			if len(cp.HexInput) == 6 {
				if c, ok := HexToRGBColor("#" + cp.HexInput); ok {
					cp.CurHex = "#" + cp.HexInput
					h, s, v := RGBToHSV(c.R(), c.G(), c.B())
					cp.Hue = h
					cp.Sat = s
					cp.Val = v
					if cp.OnPreview != nil {
						cp.OnPreview(cp.ColorKey, cp.CurHex)
					}
				}
			}
			return true, false
		}

	case input.KeyRune:
		if cp.FocusMode == 2 {
			r := k.Rune
			if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') {
				if len(cp.HexInput) < 6 {
					cp.HexInput += string(r)
					if len(cp.HexInput) == 6 {
						if c, ok := HexToRGBColor("#" + cp.HexInput); ok {
							cp.CurHex = "#" + cp.HexInput
							h, s, v := RGBToHSV(c.R(), c.G(), c.B())
							cp.Hue = h
							cp.Sat = s
							cp.Val = v
							if cp.OnPreview != nil {
								cp.OnPreview(cp.ColorKey, cp.CurHex)
							}
						}
					}
				}
				return true, false
			}
		}
	}

	return false, false
}

func (cp *ColorPickerModal) syncHex() {
	rgb := HSVToRGB(cp.Hue, cp.Sat, cp.Val)
	cp.CurHex = fmt.Sprintf("#%02X%02X%02X", rgb.R(), rgb.G(), rgb.B())
	cp.HexInput = fmt.Sprintf("%02X%02X%02X", rgb.R(), rgb.G(), rgb.B())
	if cp.OnPreview != nil {
		cp.OnPreview(cp.ColorKey, cp.CurHex)
	}
}

// HandleMouse processes mouse clicks and drags inside the ColorPicker modal.
func (cp *ColorPickerModal) HandleMouse(ev input.Mouse, screenW, screenH int) (handled bool, shouldClose bool) {
	if !cp.Open {
		return false, false
	}

	modalW := 52
	modalH := 20
	startX := (screenW - modalW) / 2
	startY := (screenH - modalH) / 2

	if ev.X < startX || ev.X >= startX+modalW || ev.Y < startY || ev.Y >= startY+modalH {
		if ev.Button == input.MouseLeft && ev.Action == input.MousePress {
			cp.Open = false
			return true, true
		}
		return true, false
	}

	// Bottom action bar buttons
	instY := startY + modalH - 1
	if ev.Y == instY && ev.Button == input.MouseLeft && ev.Action == input.MousePress {
		cancelStr := fmt.Sprintf(" %s ", cp.HintCancel)
		cancelStartX := startX + modalW - 2 - len([]rune(cancelStr))
		cancelEndX := cancelStartX + len([]rune(cancelStr))

		applyStr := fmt.Sprintf(" %s ", cp.HintApply)
		applyStartX := cancelStartX - len([]rune(applyStr)) - 1
		applyEndX := applyStartX + len([]rune(applyStr))

		if ev.X >= applyStartX && ev.X < applyEndX {
			if cp.OnApply != nil {
				cp.OnApply(cp.ColorKey, cp.CurHex)
			}
			cp.Open = false
			return true, true
		} else if ev.X >= cancelStartX && ev.X < cancelEndX {
			cp.Open = false
			return true, true
		}
	}

	wheelCenterX := startX + 13
	wheelCenterY := startY + 8
	dx := float64(ev.X - wheelCenterX)
	dy := float64(ev.Y - wheelCenterY)
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist <= float64(cp.Radius)+1.5 && ev.Y >= startY+2 && ev.Y <= startY+14 && ev.X < startX+26 {
		if ev.Button == input.MouseLeft && (ev.Action == input.MousePress || ev.Action == input.MouseMotion) {
			cp.FocusMode = 0
			angle := math.Atan2(dy, dx) * 180.0 / math.Pi
			if angle < 0 {
				angle += 360.0
			}
			cp.Hue = angle
			cp.Sat = math.Min(1.0, dist/float64(cp.Radius))
			cp.syncHex()
			return true, false
		}
	}

	sliderY := startY + 15
	sliderStartX := startX + 18
	sliderLen := 20
	if ev.Y == sliderY && ev.X >= sliderStartX && ev.X <= sliderStartX+sliderLen {
		if ev.Button == input.MouseLeft && (ev.Action == input.MousePress || ev.Action == input.MouseMotion) {
			cp.FocusMode = 1
			relX := ev.X - sliderStartX
			cp.Val = math.Max(0.0, math.Min(1.0, float64(relX)/float64(sliderLen)))
			cp.syncHex()
			return true, false
		}
	}

	hexY := startY + 11
	if ev.Y == hexY && ev.X >= startX+28 && ev.X <= startX+48 {
		if ev.Button == input.MouseLeft && ev.Action == input.MousePress {
			cp.FocusMode = 2
			return true, false
		}
	}

	return true, false
}

// Render draws the ColorPicker modal into the buffer.
func (cp *ColorPickerModal) Render(buf *buffer.Buffer, w, h int, th ColorPickerTheme) {
	if !cp.Open || buf == nil {
		return
	}

	modalW := 52
	modalH := 20
	startX := (w - modalW) / 2
	startY := (h - modalH) / 2

	themeBg := th.Background
	themeFg := th.Foreground
	accentFg := th.Accent
	commentFg := th.Comment

	// Fill background
	for y := 0; y < modalH; y++ {
		for x := 0; x < modalW; x++ {
			buf.SetRune(startX+x, startY+y, ' ', themeFg, themeBg, cell.AttrNone)
		}
	}

	// Border
	for x := 0; x < modalW; x++ {
		buf.SetRune(startX+x, startY, '─', commentFg, themeBg, cell.AttrNone)
		buf.SetRune(startX+x, startY+modalH-1, '─', commentFg, themeBg, cell.AttrNone)
	}
	for y := 0; y < modalH; y++ {
		buf.SetRune(startX, startY+y, '│', commentFg, themeBg, cell.AttrNone)
		buf.SetRune(startX+modalW-1, startY+y, '│', commentFg, themeBg, cell.AttrNone)
	}
	buf.SetRune(startX, startY, '┌', commentFg, themeBg, cell.AttrNone)
	buf.SetRune(startX+modalW-1, startY, '┐', commentFg, themeBg, cell.AttrNone)
	buf.SetRune(startX, startY+modalH-1, '└', commentFg, themeBg, cell.AttrNone)
	buf.SetRune(startX+modalW-1, startY+modalH-1, '┘', commentFg, themeBg, cell.AttrNone)

	// Title
	title := fmt.Sprintf(" 🎨 Color Picker: %s ", cp.ColorLabel)
	for i, r := range []rune(title) {
		buf.SetRune(startX+2+i, startY, r, accentFg, themeBg, cell.AttrBold)
	}

	// Instructions
	instY := startY + modalH - 1
	tabModeStr := fmt.Sprintf(" %s ", cp.HintTab)
	for i, r := range []rune(tabModeStr) {
		buf.SetRune(startX+2+i, instY, r, commentFg, themeBg, cell.AttrNone)
	}

	cancelStr := fmt.Sprintf(" %s ", cp.HintCancel)
	cancelStartX := startX + modalW - 2 - len([]rune(cancelStr))
	for i, r := range []rune(cancelStr) {
		buf.SetRune(cancelStartX+i, instY, r, themeFg, themeBg, cell.AttrNone)
	}

	applyStr := fmt.Sprintf(" %s ", cp.HintApply)
	applyStartX := cancelStartX - len([]rune(applyStr)) - 1
	for i, r := range []rune(applyStr) {
		buf.SetRune(applyStartX+i, instY, r, accentFg, themeBg, cell.AttrBold)
	}
}

// HSVToRGB converts hue (0..360), saturation (0..1), and value (0..1) to 24-bit TrueColor.
func HSVToRGB(h, s, v float64) cell.Color {
	if s <= 0 {
		val := uint8(math.Round(v * 255.0))
		colorVal := (uint32(val) << 16) | (uint32(val) << 8) | uint32(val)
		return toColor(colorVal)
	}

	hh := h / 60.0
	i := int(math.Floor(hh)) % 6
	ff := hh - math.Floor(hh)
	p := v * (1.0 - s)
	q := v * (1.0 - (s * ff))
	t := v * (1.0 - (s * (1.0 - ff)))

	var r, g, b float64
	switch i {
	case 0:
		r, g, b = v, t, p
	case 1:
		r, g, b = q, v, p
	case 2:
		r, g, b = p, v, t
	case 3:
		r, g, b = p, q, v
	case 4:
		r, g, b = t, p, v
	default:
		r, g, b = v, p, q
	}

	rUint := uint8(math.Round(r * 255.0))
	gUint := uint8(math.Round(g * 255.0))
	bUint := uint8(math.Round(b * 255.0))
	return toColor((uint32(rUint) << 16) | (uint32(gUint) << 8) | uint32(bUint))
}

// RGBToHSV converts 8-bit RGB components to Hue (0..360), Saturation (0..1), and Value (0..1).
func RGBToHSV(r, g, b uint8) (h, s, v float64) {
	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0

	maxVal := math.Max(rf, math.Max(gf, bf))
	minVal := math.Min(rf, math.Min(gf, bf))
	delta := maxVal - minVal

	v = maxVal
	if maxVal > 0 {
		s = delta / maxVal
	} else {
		s = 0
		h = 0
		return
	}

	if delta == 0 {
		h = 0
		return
	}

	if rf == maxVal {
		h = (gf - bf) / delta
	} else if gf == maxVal {
		h = 2.0 + (bf-rf)/delta
	} else {
		h = 4.0 + (rf-gf)/delta
	}

	h *= 60.0
	if h < 0 {
		h += 360.0
	}
	return
}

// HexToRGBColor parses a hex color string into cell.Color.
func HexToRGBColor(hexStr string) (cell.Color, bool) {
	clean := strings.TrimPrefix(strings.TrimSpace(hexStr), "#")
	if len(clean) == 3 {
		r, _ := strconv.ParseUint(string(clean[0])+string(clean[0]), 16, 8)
		g, _ := strconv.ParseUint(string(clean[1])+string(clean[1]), 16, 8)
		b, _ := strconv.ParseUint(string(clean[2])+string(clean[2]), 16, 8)
		colorVal := (uint32(r) << 16) | (uint32(g) << 8) | uint32(b)
		return toColor(colorVal), true
	}
	if len(clean) >= 6 {
		val, err := strconv.ParseUint(clean[:6], 16, 32)
		if err == nil {
			return toColor(uint32(val)), true
		}
	}
	return cell.Color{}, false
}
