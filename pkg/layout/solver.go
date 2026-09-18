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

	// Pass 1: Resolve Fixed constraints first (invariant non-negotiable sizes)
	for i, c := range constraints {
		if c.Type == TypeFixed {
			size := min(c.Val, remaining)
			sizes[i] = size
			remaining -= size
		}
	}

	// Pass 2: Resolve Percent and Ratio constraints (proportional fractions of total space)
	for i, c := range constraints {
		switch c.Type {
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
		}
	}

	// Pass 3: Resolve Min and Max constraints
	for i, c := range constraints {
		switch c.Type {
		case TypeMin, TypeMax:
			size := min(c.Val, remaining)
			sizes[i] = size
			remaining -= size
		}
	}

	// Pass 4: Distribute remaining space among Flex constraints
	totalFlex := 0
	for _, c := range constraints {
		if c.Type == TypeFlex {
			totalFlex += c.Val
		}
	}
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
