package style

const (
	dirUp    = 1 << 0
	dirRight = 1 << 1
	dirDown  = 1 << 2
	dirLeft  = 1 << 3
)

func boxRuneToMask(r rune) int {
	switch r {
	case '─':
		return dirLeft | dirRight
	case '│':
		return dirUp | dirDown
	case '┌', '╭':
		return dirDown | dirRight
	case '┐', '╮':
		return dirDown | dirLeft
	case '└', '╰':
		return dirUp | dirRight
	case '┘', '╯':
		return dirUp | dirLeft
	case '├':
		return dirUp | dirDown | dirRight
	case '┤':
		return dirUp | dirDown | dirLeft
	case '┬':
		return dirDown | dirLeft | dirRight
	case '┴':
		return dirUp | dirLeft | dirRight
	case '┼':
		return dirUp | dirDown | dirLeft | dirRight
	default:
		return 0
	}
}

func maskToBoxRune(mask int) rune {
	switch mask {
	case dirLeft | dirRight:
		return '─'
	case dirUp | dirDown:
		return '│'
	case dirDown | dirRight:
		return '┌'
	case dirDown | dirLeft:
		return '┐'
	case dirUp | dirRight:
		return '└'
	case dirUp | dirLeft:
		return '┘'
	case dirUp | dirDown | dirRight:
		return '├'
	case dirUp | dirDown | dirLeft:
		return '┤'
	case dirDown | dirLeft | dirRight:
		return '┬'
	case dirUp | dirLeft | dirRight:
		return '┴'
	case dirUp | dirDown | dirLeft | dirRight:
		return '┼'
	default:
		return 0
	}
}

// MergeBoxRunes combines two overlapping box-drawing characters into an appropriate junction character.
// For example, combining '│' and '─' returns '┼'; combining '│' and '┌' returns '├'.
func MergeBoxRunes(existing, incoming rune) rune {
	m1 := boxRuneToMask(existing)
	m2 := boxRuneToMask(incoming)
	if m1 == 0 {
		return incoming
	}
	if m2 == 0 {
		return existing
	}
	res := maskToBoxRune(m1 | m2)
	if res == 0 {
		return incoming
	}
	return res
}
