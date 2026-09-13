package explorer

import "strings"

// IconMode defines which icon set to use for files and directories.
type IconMode int

const (
	// IconModeASCII uses text labels like [DIR], [SRC], [IMG] — 100% stable width.
	IconModeASCII IconMode = iota
	// IconModeUnicode uses single-width geometric symbols.
	IconModeUnicode
	// IconModeNerdFont uses Nerd Font glyphs (requires patched font).
	IconModeNerdFont
	// IconModeEmoji uses emoji characters (may cause width drift).
	IconModeEmoji
)

// IconModeNames maps mode to display name.
var IconModeNames = [4]string{"ASCII", "Unicode", "NerdFont", "Emoji"}

// String returns the display name of the icon mode.
func (m IconMode) String() string {
	if m >= 0 && int(m) < len(IconModeNames) {
		return IconModeNames[m]
	}
	return "ASCII"
}

// NextMode cycles to the next icon mode.
func (m IconMode) NextMode() IconMode {
	return (m + 1) % 4
}

type iconEntry struct {
	Dir     string
	Source  string
	Image   string
	Video   string
	Config  string
	Doc     string
	Archive string
	Binary  string
	Generic string
}

var iconSets = [4]iconEntry{
	// ASCII
	{Dir: "[DIR]", Source: "[SRC]", Image: "[IMG]", Video: "[VID]", Config: "[CFG]", Doc: "[DOC]", Archive: "[ARC]", Binary: "[BIN]", Generic: "[   ]"},
	// Unicode  
	{Dir: "\u25A0", Source: "\u00B7", Image: "\u00A7", Video: "\u25C6", Config: "\u25B6", Doc: "\u25B2", Archive: "\u25CB", Binary: "\u2666", Generic: "\u00B7"},
	// NerdFont (use common nerd font codepoints)
	{Dir: "\uF07B", Source: "\uF15C", Image: "\uF03E", Video: "\uF008", Config: "\uF013", Doc: "\uF0F6", Archive: "\uF187", Binary: "\uF471", Generic: "\uF016"},
	// Emoji
	{Dir: "\U0001F4C1", Source: "\U0001F4C4", Image: "\U0001F5BC\uFE0F", Video: "\U0001F3AC", Config: "\u2699\uFE0F", Doc: "\U0001F4DD", Archive: "\U0001F4E6", Binary: "\u26A1", Generic: "\U0001F4C4"},
}

// IconWidth returns the display width of icons in the current mode.
// ASCII icons are 5 chars, Unicode/NerdFont are 1, Emoji varies.
func IconWidth(mode IconMode) int {
	switch mode {
	case IconModeASCII:
		return 5
	case IconModeUnicode, IconModeNerdFont:
		return 1
	case IconModeEmoji:
		return 2
	default:
		return 5
	}
}

// GetIcon returns the appropriate icon for a file entry based on the active icon mode.
func GetIcon(mode IconMode, isDir bool, ext string) string {
	set := iconSets[0] // default ASCII
	if int(mode) < len(iconSets) {
		set = iconSets[mode]
	}

	if isDir {
		return set.Dir
	}

	ext = strings.ToLower(ext)
	switch ext {
	case ".png", ".jpg", ".jpeg", ".webp", ".bmp", ".ico", ".svg":
		return set.Image
	case ".gif", ".mp4", ".avi", ".mkv", ".mov", ".webm":
		return set.Video
	case ".go", ".ts", ".js", ".py", ".rs", ".c", ".cpp", ".h", ".java", ".rb", ".php", ".cs", ".swift", ".kt":
		return set.Source
	case ".json", ".yaml", ".yml", ".toml", ".xml", ".ini", ".env", ".conf", ".cfg":
		return set.Config
	case ".md", ".txt", ".rst", ".doc", ".pdf", ".rtf", ".tex", ".log":
		return set.Doc
	case ".zip", ".tar", ".gz", ".7z", ".rar", ".bz2", ".xz", ".zst":
		return set.Archive
	case ".exe", ".dll", ".so", ".bin", ".o", ".a", ".dylib":
		return set.Binary
	default:
		return set.Generic
	}
}
