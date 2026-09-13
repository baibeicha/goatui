package buffer

// RuneWidth returns the terminal cell display width for a given rune.
// Returns 0 for non-spacing/combining marks, 1 for normal characters,
// and 2 for East Asian Wide / Emoji characters.
func RuneWidth(r rune) int {
	// Fast-path: printable ASCII
	if r >= 0x20 && r < 0x7F {
		return 1
	}
	if r < 0x20 || (r >= 0x7F && r < 0xA0) {
		return 0
	}

	// Zero-width combining and formatting characters
	if isZeroWidth(r) {
		return 0
	}

	if r == 0xFE0E || r == 0xFE0F {
		return 0
	}

	// Wide characters (CJK, Fullwidth, Emoji)
	if isWide(r) {
		return 2
	}

	return 1
}

// StringWidth returns the total visual cell width of a UTF-8 string.
func StringWidth(s string) int {
	w := 0
	for _, r := range s {
		w += RuneWidth(r)
	}
	return w
}

func isZeroWidth(r rune) bool {
	if r >= 0x0300 && r <= 0x036F { // Combining Diacritical Marks
		return true
	}
	if r >= 0x0483 && r <= 0x0489 { // Cyrillic Combining Marks
		return true
	}
	if r >= 0x0591 && r <= 0x05BD || r == 0x05BF || r >= 0x05C1 && r <= 0x05C2 || r >= 0x05C4 && r <= 0x05C5 || r == 0x05C7 { // Hebrew
		return true
	}
	if r >= 0x0610 && r <= 0x061A || r >= 0x064B && r <= 0x065F || r == 0x0670 { // Arabic
		return true
	}
	if r >= 0x1DC0 && r <= 0x1DFF || r >= 0x20D0 && r <= 0x20FF || r >= 0xFE20 && r <= 0xFE2F { // Combining marks
		return true
	}
	if r >= 0x200B && r <= 0x200F { // Zero-width spaces and formatting
		return true
	}
	if r == 0x2028 || r == 0x2029 || (r >= 0x202A && r <= 0x202E) { // Directional formatting
		return true
	}
	if r == 0x2060 || r == 0xFEFF { // Word joiner / BOM
		return true
	}
	return false
}

func isWide(r rune) bool {
	// Hangul Jamo, CJK Radicals, Kangxi, Ideographic Description
	if r >= 0x1100 && (r <= 0x115F || r == 0x2329 || r == 0x232A ||
		(r >= 0x2E80 && r <= 0xA4CF && r != 0x303F) ||
		(r >= 0xAC00 && r <= 0xD7A3)) {
		return true
	}

	// CJK Compatibility, Enclosed, Ideographs
	if r >= 0xF900 && r <= 0xFAFF || (r >= 0xFE10 && r <= 0xFE19) ||
		(r >= 0xFE30 && r <= 0xFE6F) || (r >= 0xFF01 && r <= 0xFF60) ||
		(r >= 0xFFE0 && r <= 0xFFE6) {
		return true
	}

	// Plane 2 & 3 CJK Unified Ideographs
	if r >= 0x20000 && r <= 0x3FFFD {
		return true
	}

	// Emoji ranges
	if (r >= 0x1F300 && r <= 0x1F64F) || // Misc Symbols & Pictographs, Emoticons
		(r >= 0x1F680 && r <= 0x1F6FF) || // Transport and Map
		(r >= 0x1F900 && r <= 0x1F9FF) || // Supplemental Symbols
		(r >= 0x1FA00 && r <= 0x1FAFF) { // Chess, Symbols & Pictographs Extended-A
		return true
	}

	return false
}
