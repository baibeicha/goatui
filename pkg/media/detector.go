package media

import (
	"os"
	"strings"
	"sync"
)

// Protocol identifies the terminal graphics rendering protocol.
type Protocol int

const (
	// ProtoKitty represents Kitty Graphics Protocol (GPU-accelerated inline graphics).
	ProtoKitty Protocol = iota
	// ProtoITerm2 represents iTerm2 Inline Images Protocol.
	ProtoITerm2
	// ProtoSixel represents DEC Sixel raster graphics.
	ProtoSixel
	// ProtoHalfBlock represents TrueColor 24-bit half-block (▀) rendering (2 pixels per cell).
	// Universally compatible across Windows Terminal, cmd.exe, PowerShell, VS Code, Linux console.
	ProtoHalfBlock
	// ProtoBraille represents 2x4 dot matrix Braille unicode character fallback.
	ProtoBraille
)

func (p Protocol) String() string {
	switch p {
	case ProtoKitty:
		return "Kitty Graphics"
	case ProtoITerm2:
		return "iTerm2 Inline"
	case ProtoSixel:
		return "Sixel"
	case ProtoHalfBlock:
		return "Half-Block TrueColor (▀)"
	case ProtoBraille:
		return "Braille Matrix"
	default:
		return "Unknown"
	}
}

var (
	detectOnce      sync.Once
	currentProtocol Protocol = ProtoHalfBlock
	protocolMu      sync.RWMutex
)

// DetectProtocol performs single-pass capability detection to select the best available
// graphics protocol supported by the host terminal. Evaluation order is strictly:
// Kitty -> iTerm2 -> Sixel -> Half-Block TrueColor -> Braille.
// Once evaluated, the result is cached for the lifetime of the process.
func DetectProtocol() Protocol {
	detectOnce.Do(func() {
		currentProtocol = evaluateCapabilities()
	})

	protocolMu.RLock()
	defer protocolMu.RUnlock()
	return currentProtocol
}

// CurrentProtocol returns the currently active protocol without triggering detection if not done yet.
func CurrentProtocol() Protocol {
	return DetectProtocol()
}

// SetProtocol explicitly overrides the active graphics protocol.
func SetProtocol(p Protocol) {
	protocolMu.Lock()
	defer protocolMu.Unlock()
	currentProtocol = p
}

// evaluateCapabilities inspects terminal environment variables and capabilities from best to worst.
func evaluateCapabilities() Protocol {
	term := strings.ToLower(os.Getenv("TERM"))
	termProg := strings.ToLower(os.Getenv("TERM_PROGRAM"))
	colorTerm := strings.ToLower(os.Getenv("COLORTERM"))
	lcTerm := strings.ToLower(os.Getenv("LC_TERMINAL"))

	// 1. Kitty Graphics Protocol
	if os.Getenv("KITTY_WINDOW_ID") != "" ||
		os.Getenv("GHOSTTY_RESOURCES_DIR") != "" ||
		strings.Contains(term, "kitty") ||
		strings.Contains(termProg, "kitty") ||
		strings.Contains(termProg, "ghostty") {
		return ProtoKitty
	}

	// 2. iTerm2 Inline Protocol
	if strings.Contains(termProg, "iterm") ||
		strings.Contains(lcTerm, "iterm") ||
		strings.Contains(termProg, "wezterm") {
		return ProtoITerm2
	}

	// 3. Sixel Graphics
	if strings.Contains(term, "sixel") ||
		strings.Contains(term, "foot") ||
		strings.Contains(term, "mlterm") {
		return ProtoSixel
	}

	// 4. Half-Block TrueColor (▀)
	// Supported by Windows Terminal, modern cmd.exe/pwsh, VS Code terminal, modern Linux xterms
	if colorTerm == "truecolor" || colorTerm == "24bit" ||
		os.Getenv("WT_SESSION") != "" ||
		termProg == "vscode" ||
		term == "xterm-256color" ||
		term == "screen-256color" ||
		term == "tmux-256color" ||
		term == "alacritty" {
		return ProtoHalfBlock
	}

	// On Windows, Windows Terminal or Win10+ conhost supports ANSI TrueColor
	if os.Getenv("WT_SESSION") != "" || os.Getenv("OS") == "Windows_NT" {
		return ProtoHalfBlock
	}

	// 5. Fallback
	return ProtoBraille
}
