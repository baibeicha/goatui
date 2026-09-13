package style

// Border defines the characters used for drawing rectangular borders.
type Border struct {
	Top         rune
	Bottom      rune
	Left        rune
	Right       rune
	TopLeft     rune
	TopRight    rune
	BottomLeft  rune
	BottomRight rune
}

var (
	// BorderNormal is standard single-line box drawing border: ┌─┐│└─┘
	BorderNormal = Border{
		Top:         '─',
		Bottom:      '─',
		Left:        '│',
		Right:       '│',
		TopLeft:     '┌',
		TopRight:    '┐',
		BottomLeft:  '└',
		BottomRight: '┘',
	}

	// BorderRounded uses smooth rounded corners: ╭─╮│╰─╯
	BorderRounded = Border{
		Top:         '─',
		Bottom:      '─',
		Left:        '│',
		Right:       '│',
		TopLeft:     '╭',
		TopRight:    '╮',
		BottomLeft:  '╰',
		BottomRight: '╯',
	}

	// BorderThick uses heavy box drawing lines: ┏━┓┃┗━┛
	BorderThick = Border{
		Top:         '━',
		Bottom:      '━',
		Left:        '┃',
		Right:       '┃',
		TopLeft:     '┏',
		TopRight:    '┓',
		BottomLeft:  '┗',
		BottomRight: '┛',
	}

	// BorderDouble uses double lines: ╔═╗║╚═╝
	BorderDouble = Border{
		Top:         '═',
		Bottom:      '═',
		Left:        '║',
		Right:       '║',
		TopLeft:     '╔',
		TopRight:    '╗',
		BottomLeft:  '╚',
		BottomRight: '╝',
	}

	// BorderASCII uses standard ASCII characters: +-+|+-+
	BorderASCII = Border{
		Top:         '-',
		Bottom:      '-',
		Left:        '|',
		Right:       '|',
		TopLeft:     '+',
		TopRight:    '+',
		BottomLeft:  '+',
		BottomRight: '+',
	}

	// BorderBlock uses full block characters: █
	BorderBlock = Border{
		Top:         '█',
		Bottom:      '█',
		Left:        '█',
		Right:       '█',
		TopLeft:     '█',
		TopRight:    '█',
		BottomLeft:  '█',
		BottomRight: '█',
	}
)
