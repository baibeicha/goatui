package ui

import (
	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/layout"
)

// View is the base interface for any renderable UI component.
type View interface {
	Draw(buf *buffer.Buffer, area buffer.Rect)
}

// ViewFunc allows a plain function to satisfy the View interface.
type ViewFunc func(buf *buffer.Buffer, area buffer.Rect)

// Draw invokes the underlying function.
func (f ViewFunc) Draw(buf *buffer.Buffer, area buffer.Rect) {
	if f != nil {
		f(buf, area)
	}
}

// LayoutItem combines a layout sizing constraint with a View component.
type LayoutItem struct {
	Constraint layout.Constraint
	View       View
}

// Fixed creates a layout item with an exact number of cells.
func Fixed(size int, view View) LayoutItem {
	return LayoutItem{
		Constraint: layout.Fixed(size),
		View:       view,
	}
}

// Flex creates a layout item with proportional stretch weight.
func Flex(factor int, view View) LayoutItem {
	return LayoutItem{
		Constraint: layout.Flex(factor),
		View:       view,
	}
}

// Percent creates a layout item with a percentage (0..100) of available space.
func Percent(pct int, view View) LayoutItem {
	return LayoutItem{
		Constraint: layout.Percent(pct),
		View:       view,
	}
}

// Auto creates an auto-flex layout item (Flex factor 1).
func Auto(view View) LayoutItem {
	return LayoutItem{
		Constraint: layout.Flex(1),
		View:       view,
	}
}

// Spacer creates an empty expandable layout item.
func Spacer(factor int) LayoutItem {
	return LayoutItem{
		Constraint: layout.Flex(factor),
		View:       nil,
	}
}

// Container manages layout subdivision and renders child views.
type Container struct {
	direction   layout.Direction
	items       []LayoutItem
	constraints []layout.Constraint
	dstRects    []buffer.Rect
}

// VBox creates a vertical layout container (top to bottom).
func VBox(children ...LayoutItem) *Container {
	c := &Container{
		direction:   layout.Vertical,
		items:       children,
		constraints: make([]layout.Constraint, len(children)),
		dstRects:    make([]buffer.Rect, len(children)),
	}
	for i, item := range children {
		c.constraints[i] = item.Constraint
	}
	return c
}

// HBox creates a horizontal layout container (left to right).
func HBox(children ...LayoutItem) *Container {
	c := &Container{
		direction:   layout.Horizontal,
		items:       children,
		constraints: make([]layout.Constraint, len(children)),
		dstRects:    make([]buffer.Rect, len(children)),
	}
	for i, item := range children {
		c.constraints[i] = item.Constraint
	}
	return c
}

// Add appends a child to the container.
func (c *Container) Add(item LayoutItem) *Container {
	c.items = append(c.items, item)
	c.constraints = append(c.constraints, item.Constraint)
	c.dstRects = append(c.dstRects, buffer.Rect{})
	return c
}

// Draw divides the area according to child constraints and renders each child.
func (c *Container) Draw(buf *buffer.Buffer, area buffer.Rect) {
	if area.IsEmpty() || len(c.items) == 0 {
		return
	}

	// In-place split into c.dstRects eliminates heap allocations
	rects := layout.SplitInto(area, c.direction, c.constraints, c.dstRects)

	for i, item := range c.items {
		if item.View != nil && i < len(rects) && !rects[i].IsEmpty() {
			item.View.Draw(buf, rects[i])
		}
	}
}

// Padding creates a padded view wrapping child.
func Padding(top, right, bottom, left int, child View) View {
	return ViewFunc(func(buf *buffer.Buffer, area buffer.Rect) {
		if child == nil || area.IsEmpty() {
			return
		}
		inner := buffer.NewRect(
			area.X+left,
			area.Y+top,
			max(0, area.Width-(left+right)),
			max(0, area.Height-(top+bottom)),
		)
		if !inner.IsEmpty() {
			child.Draw(buf, inner)
		}
	})
}

// Pad creates an evenly padded view wrapping child.
func Pad(padding int, child View) View {
	return Padding(padding, padding, padding, padding, child)
}

// Center centers a child view of fixed or bounded size within the allocated area.
func Center(w, h int, child View) View {
	return ViewFunc(func(buf *buffer.Buffer, area buffer.Rect) {
		if child == nil || area.IsEmpty() {
			return
		}
		actualW := min(w, area.Width)
		actualH := min(h, area.Height)
		x := area.X + max(0, (area.Width-actualW)/2)
		y := area.Y + max(0, (area.Height-actualH)/2)
		childArea := buffer.NewRect(x, y, actualW, actualH)
		child.Draw(buf, childArea)
	})
}
