package buffer

// Rect defines a 2D integer rectangle in terminal screen coordinates.
type Rect struct {
	X      int
	Y      int
	Width  int
	Height int
}

// NewRect creates a Rect with position and size.
func NewRect(x, y, w, h int) Rect {
	if w < 0 {
		w = 0
	}
	if h < 0 {
		h = 0
	}
	return Rect{X: x, Y: y, Width: w, Height: h}
}

// Area returns the number of cells covered by the rectangle.
func (r Rect) Area() int {
	return r.Width * r.Height
}

// IsEmpty returns true if width or height is zero or negative.
func (r Rect) IsEmpty() bool {
	return r.Width <= 0 || r.Height <= 0
}

// Right returns the right boundary coordinate (X + Width).
func (r Rect) Right() int {
	return r.X + r.Width
}

// Bottom returns the bottom boundary coordinate (Y + Height).
func (r Rect) Bottom() int {
	return r.Y + r.Height
}

// Contains returns true if the coordinate (x, y) is within bounds.
func (r Rect) Contains(x, y int) bool {
	return x >= r.X && x < r.X+r.Width && y >= r.Y && y < r.Y+r.Height
}

// Intersection returns the overlapping rectangle between r and other.
func (r Rect) Intersection(other Rect) Rect {
	x1 := max(r.X, other.X)
	y1 := max(r.Y, other.Y)
	x2 := min(r.Right(), other.Right())
	y2 := min(r.Bottom(), other.Bottom())

	if x2 <= x1 || y2 <= y1 {
		return Rect{}
	}
	return Rect{X: x1, Y: y1, Width: x2 - x1, Height: y2 - y1}
}

// Union returns the smallest rectangle enclosing both r and other.
func (r Rect) Union(other Rect) Rect {
	if r.IsEmpty() {
		return other
	}
	if other.IsEmpty() {
		return r
	}
	x1 := min(r.X, other.X)
	y1 := min(r.Y, other.Y)
	x2 := max(r.Right(), other.Right())
	y2 := max(r.Bottom(), other.Bottom())
	return Rect{X: x1, Y: y1, Width: x2 - x1, Height: y2 - y1}
}

// Inset shrinks (or expands if negative) the rectangle by horizontal and vertical amounts.
func (r Rect) Inset(dx, dy int) Rect {
	w := r.Width - 2*dx
	h := r.Height - 2*dy
	if w < 0 {
		w = 0
	}
	if h < 0 {
		h = 0
	}
	return Rect{
		X:      r.X + dx,
		Y:      r.Y + dy,
		Width:  w,
		Height: h,
	}
}

// Clamp constrains coordinates (x, y) to inside the rectangle.
func (r Rect) Clamp(x, y int) (int, int) {
	if r.IsEmpty() {
		return r.X, r.Y
	}
	cx := min(max(x, r.X), r.Right()-1)
	cy := min(max(y, r.Y), r.Bottom()-1)
	return cx, cy
}
