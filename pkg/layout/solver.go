package layout

import "github.com/baibeicha/goatui/pkg/core/buffer"

// Split divides an area rectangle along the specified direction according to constraints.
func Split(area buffer.Rect, dir Direction, constraints ...Constraint) []buffer.Rect {
	dst := make([]buffer.Rect, len(constraints))
	return SplitInto(area, dir, constraints, dst)
}

// SplitHorizontal divides an area into side-by-side columns.
func SplitHorizontal(area buffer.Rect, constraints ...Constraint) []buffer.Rect {
	return Split(area, Horizontal, constraints...)
}

// SplitVertical divides an area into top-to-bottom rows.
func SplitVertical(area buffer.Rect, constraints ...Constraint) []buffer.Rect {
	return Split(area, Vertical, constraints...)
}

// SplitInto divides an area in-place into dst, eliminating heap allocations when dst is reused.
func SplitInto(area buffer.Rect, dir Direction, constraints []Constraint, dst []buffer.Rect) []buffer.Rect {
	n := len(constraints)
	if n == 0 {
		return dst[:0]
	}

	if cap(dst) < n {
		dst = make([]buffer.Rect, n)
	} else {
		dst = dst[:n]
	}

	totalSpace := area.Width
	if dir == Vertical {
		totalSpace = area.Height
	}

	if totalSpace <= 0 || area.IsEmpty() {
		for i := range dst {
			dst[i] = buffer.Rect{X: area.X, Y: area.Y, Width: 0, Height: 0}
		}
		return dst
	}

	// Sizes array (up to 32 elements on stack, or dynamic slice if n > 32)
	var stackSizes [32]int
	var sizes []int
	if n <= 32 {
		sizes = stackSizes[:n]
	} else {
		sizes = make([]int, n)
	}

	remaining := totalSpace
	totalFlex := 0

	// Pass 1: Resolve Fixed, Percent, Ratio, and Min sizes
	for i, c := range constraints {
		switch c.Type {
		case TypeFixed:
			size := min(c.Val, remaining)
			sizes[i] = size
			remaining -= size

		case TypePercent:
			size := (totalSpace * c.Val) / 100
			size = min(size, remaining)
			sizes[i] = size
			remaining -= size

		case TypeRatio:
			size := 0
			if c.Val2 > 0 {
				size = (totalSpace * c.Val) / c.Val2
			}
			size = min(size, remaining)
			sizes[i] = size
			remaining -= size

		case TypeMin:
			size := min(c.Val, remaining)
			sizes[i] = size
			remaining -= size

		case TypeMax:
			// Initialized to max, will be limited if needed
			sizes[i] = 0

		case TypeFlex:
			totalFlex += c.Val
		}
	}

	// Pass 2: Distribute remaining space among Flex constraints
	if totalFlex > 0 && remaining > 0 {
		flexSpace := remaining
		allocated := 0

		for i, c := range constraints {
			if c.Type == TypeFlex {
				share := (flexSpace * c.Val) / totalFlex
				sizes[i] += share
				allocated += share
			}
		}

		// Distribute rounding remainder to first flex items
		rem := flexSpace - allocated
		for i := 0; i < n && rem > 0; i++ {
			if constraints[i].Type == TypeFlex {
				sizes[i]++
				rem--
			}
		}
		remaining = 0
	}

	// Pass 3: Clamp Max constraints
	for i, c := range constraints {
		if c.Type == TypeMax && sizes[i] > c.Val {
			sizes[i] = c.Val
		}
	}

	// Pass 4: Construct destination rectangles
	currentOffset := 0
	for i, size := range sizes {
		if dir == Horizontal {
			dst[i] = buffer.Rect{
				X:      area.X + currentOffset,
				Y:      area.Y,
				Width:  size,
				Height: area.Height,
			}
		} else {
			dst[i] = buffer.Rect{
				X:      area.X,
				Y:      area.Y + currentOffset,
				Width:  area.Width,
				Height: size,
			}
		}
		currentOffset += size
	}

	return dst
}
