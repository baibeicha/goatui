package layout

// Direction specifies the layout splitting axis.
type Direction uint8

const (
	Horizontal Direction = iota // Split into columns side-by-side
	Vertical                    // Split into rows top-to-bottom
)

// ConstraintType denotes the sizing rule applied to an element.
type ConstraintType uint8

const (
	TypeFixed ConstraintType = iota
	TypePercent
	TypeFlex
	TypeMin
	TypeMax
	TypeRatio
)

// Constraint represents a dimension constraint for layout subdivision.
type Constraint struct {
	Type ConstraintType
	Val  int
	Val2 int
}

// Fixed allocates an exact number of cells.
func Fixed(size int) Constraint {
	if size < 0 {
		size = 0
	}
	return Constraint{Type: TypeFixed, Val: size}
}

// Percent allocates a percentage (0..100) of the total available length.
func Percent(pct int) Constraint {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	return Constraint{Type: TypePercent, Val: pct}
}

// Flex allocates space proportionally according to its weight factor relative to other Flex children.
func Flex(factor int) Constraint {
	if factor <= 0 {
		factor = 1
	}
	return Constraint{Type: TypeFlex, Val: factor}
}

// Min ensures the element receives at least minSize cells if space permits.
func Min(minSize int) Constraint {
	if minSize < 0 {
		minSize = 0
	}
	return Constraint{Type: TypeMin, Val: minSize}
}

// Max restricts the element to at most maxSize cells.
func Max(maxSize int) Constraint {
	if maxSize < 0 {
		maxSize = 0
	}
	return Constraint{Type: TypeMax, Val: maxSize}
}

// Ratio allocates a fraction (numerator / denominator) of the available length.
func Ratio(num, den int) Constraint {
	if den <= 0 {
		den = 1
	}
	if num < 0 {
		num = 0
	}
	return Constraint{Type: TypeRatio, Val: num, Val2: den}
}
