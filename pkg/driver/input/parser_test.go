package input

import (
	"testing"
)

func TestParseKeys(t *testing.T) {
	p := NewParser()

	var events []Event
	handler := func(ev Event) {
		events = append(events, ev)
	}

	// Test regular text
	p.Parse([]byte("Goat"), handler)
	if len(events) != 4 {
		t.Fatalf("Expected 4 events, got %d", len(events))
	}
	if events[0].Key.Rune != 'G' || events[3].Key.Rune != 't' {
		t.Errorf("Unexpected runes parsed: %+v", events)
	}

	// Test Arrow Up: \x1b[A
	events = events[:0]
	p.Parse([]byte("\x1b[A"), handler)
	if len(events) != 1 || events[0].Key.Type != KeyUp {
		t.Errorf("Expected KeyUp, got %+v", events)
	}

	// Test Delete: \x1b[3~
	events = events[:0]
	p.Parse([]byte("\x1b[3~"), handler)
	if len(events) != 1 || events[0].Key.Type != KeyDelete {
		t.Errorf("Expected KeyDelete, got %+v", events)
	}

	// Test F1: \x1bOP
	events = events[:0]
	p.Parse([]byte("\x1bOP"), handler)
	if len(events) != 1 || events[0].Key.Type != KeyF1 {
		t.Errorf("Expected KeyF1, got %+v", events)
	}
}

func TestParseSGRMouse(t *testing.T) {
	p := NewParser()

	var events []Event
	handler := func(ev Event) {
		events = append(events, ev)
	}

	// Left click at (10, 20): \x1b[<0;11;21M
	p.Parse([]byte("\x1b[<0;11;21M"), handler)
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}
	m := events[0].Mouse
	if m.Button != MouseLeft || m.Action != MousePress || m.X != 10 || m.Y != 20 {
		t.Errorf("Unexpected mouse event: %+v", m)
	}

	// Wheel Up at (5, 5): \x1b[<64;6;6M
	events = events[:0]
	p.Parse([]byte("\x1b[<64;6;6M"), handler)
	if len(events) != 1 || events[0].Mouse.Button != MouseWheelUp {
		t.Errorf("Expected MouseWheelUp, got %+v", events)
	}
}

func TestParseBracketedPaste(t *testing.T) {
	p := NewParser()

	var events []Event
	handler := func(ev Event) {
		events = append(events, ev)
	}

	p.Parse([]byte("\x1b[200~Pasted Content\x1b[201~"), handler)
	if len(events) != 1 || events[0].Type != EventPaste || events[0].PasteText != "Pasted Content" {
		t.Errorf("Expected EventPaste, got %+v", events)
	}
}

func TestParseIncompleteEscapeSequence(t *testing.T) {
	p := NewParser()

	var events []Event
	handler := func(ev Event) {
		events = append(events, ev)
	}

	// Send partial escape sequence
	p.Parse([]byte("\x1b["), handler)
	if len(events) != 0 {
		t.Fatalf("Expected 0 events for incomplete escape sequence, got %d", len(events))
	}

	// Complete sequence with 'B' (Down arrow)
	p.Parse([]byte("B"), handler)
	if len(events) != 1 || events[0].Key.Type != KeyDown {
		t.Errorf("Expected KeyDown after completion, got %+v", events)
	}

	// Incomplete SS3 sequence
	events = events[:0]
	p.Parse([]byte("\x1bO"), handler)
	if len(events) != 0 {
		t.Fatalf("Expected 0 events for incomplete SS3 sequence, got %d", len(events))
	}
	// Complete SS3 sequence with 'A' (Up arrow in DECCKM mode)
	p.Parse([]byte("A"), handler)
	if len(events) != 1 || events[0].Key.Type != KeyUp {
		t.Errorf("Expected KeyUp for \\x1bOA, got %+v", events)
	}
}

func TestParseModifiersAndMouse(t *testing.T) {
	p := NewParser()

	var events []Event
	handler := func(ev Event) {
		events = append(events, ev)
	}

	// Ctrl+Space (0x00) should have ModCtrl
	p.Parse([]byte{0x00}, handler)
	if len(events) != 1 || events[0].Key.Type != KeySpace || !events[0].Key.HasCtrl() {
		t.Errorf("Expected KeySpace with ModCtrl, got %+v", events)
	}

	// Ctrl+Click SGR mouse: btn 0 + 16 = 16: \x1b[<16;5;10M
	events = events[:0]
	p.Parse([]byte("\x1b[<16;5;10M"), handler)
	if len(events) != 1 || !events[0].Mouse.HasCtrl() || events[0].Mouse.Button != MouseLeft {
		t.Errorf("Expected Left Click with Ctrl modifier, got %+v", events)
	}

	// Shift+Alt+RightClick SGR mouse: btn 2 + 4 + 8 = 14: \x1b[<14;5;10M
	events = events[:0]
	p.Parse([]byte("\x1b[<14;5;10M"), handler)
	if len(events) != 1 || !events[0].Mouse.HasShift() || !events[0].Mouse.HasAlt() || events[0].Mouse.Button != MouseRight {
		t.Errorf("Expected Right Click with Shift+Alt modifier, got %+v", events)
	}
}

func TestKittyKeyboardBasicKeys(t *testing.T) {
	p := NewParser()
	var events []Event
	handler := func(ev Event) { events = append(events, ev) }

	tests := []struct {
		seq      string
		expected KeyType
		runeVal  rune
	}{
		{"\x1b[13u", KeyEnter, 0},
		{"\x1b[27u", KeyEsc, 0},
		{"\x1b[9u", KeyTab, 0},
		{"\x1b[127u", KeyBackspace, 0},
		{"\x1b[32u", KeySpace, ' '},
	}

	for _, tc := range tests {
		events = events[:0]
		p.Parse([]byte(tc.seq), handler)
		if len(events) != 1 {
			t.Fatalf("For %q expected 1 event, got %d", tc.seq, len(events))
		}
		if events[0].Key.Type != tc.expected {
			t.Errorf("For %q expected KeyType %v, got %v", tc.seq, tc.expected, events[0].Key.Type)
		}
		if tc.runeVal != 0 && events[0].Key.Rune != tc.runeVal {
			t.Errorf("For %q expected rune %c, got %c", tc.seq, tc.runeVal, events[0].Key.Rune)
		}
	}
}

func TestKittyKeyboardModifiers(t *testing.T) {
	p := NewParser()
	var events []Event
	handler := func(ev Event) { events = append(events, ev) }

	// Shift+Enter: \x1b[13;2u
	p.Parse([]byte("\x1b[13;2u"), handler)
	if len(events) != 1 || events[0].Key.Type != KeyEnter || !events[0].Key.HasShift() {
		t.Fatalf("Expected Shift+Enter, got %+v", events)
	}

	// Ctrl+Enter: \x1b[13;5u
	events = events[:0]
	p.Parse([]byte("\x1b[13;5u"), handler)
	if len(events) != 1 || events[0].Key.Type != KeyEnter || !events[0].Key.HasCtrl() {
		t.Fatalf("Expected Ctrl+Enter, got %+v", events)
	}

	// Alt+Enter: \x1b[13;3u
	events = events[:0]
	p.Parse([]byte("\x1b[13;3u"), handler)
	if len(events) != 1 || events[0].Key.Type != KeyEnter || !events[0].Key.HasAlt() {
		t.Fatalf("Expected Alt+Enter, got %+v", events)
	}

	// Ctrl+Shift+Enter: \x1b[13;6u (1 + 1(Shift) + 4(Ctrl) = 6)
	events = events[:0]
	p.Parse([]byte("\x1b[13;6u"), handler)
	if len(events) != 1 || events[0].Key.Type != KeyEnter || !events[0].Key.HasCtrl() || !events[0].Key.HasShift() {
		t.Fatalf("Expected Ctrl+Shift+Enter, got %+v", events)
	}

	// Shift+Tab: \x1b[9;2u -> KeyBacktab
	events = events[:0]
	p.Parse([]byte("\x1b[9;2u"), handler)
	if len(events) != 1 || events[0].Key.Type != KeyBacktab || !events[0].Key.HasShift() {
		t.Fatalf("Expected KeyBacktab with Shift, got %+v", events)
	}

	// Ctrl+Tab: \x1b[9;5u -> KeyTab with Ctrl
	events = events[:0]
	p.Parse([]byte("\x1b[9;5u"), handler)
	if len(events) != 1 || events[0].Key.Type != KeyTab || !events[0].Key.HasCtrl() {
		t.Fatalf("Expected KeyTab with Ctrl, got %+v", events)
	}

	// Ctrl+Shift+Tab: \x1b[9;6u -> KeyBacktab with Ctrl and Shift
	events = events[:0]
	p.Parse([]byte("\x1b[9;6u"), handler)
	if len(events) != 1 || events[0].Key.Type != KeyBacktab || !events[0].Key.HasCtrl() || !events[0].Key.HasShift() {
		t.Fatalf("Expected KeyBacktab with Ctrl+Shift, got %+v", events)
	}

	// Ctrl+Backspace: \x1b[127;5u
	events = events[:0]
	p.Parse([]byte("\x1b[127;5u"), handler)
	if len(events) != 1 || events[0].Key.Type != KeyBackspace || !events[0].Key.HasCtrl() {
		t.Fatalf("Expected Ctrl+Backspace, got %+v", events)
	}

	// Super+Space: \x1b[32;9u (1 + 8(Super) = 9)
	events = events[:0]
	p.Parse([]byte("\x1b[32;9u"), handler)
	if len(events) != 1 || events[0].Key.Type != KeySpace || !events[0].Key.HasSuper() {
		t.Fatalf("Expected Super+Space, got %+v", events)
	}
}

func TestKittyDisambiguatedCtrlCodes(t *testing.T) {
	p := NewParser()
	var events []Event
	handler := func(ev Event) { events = append(events, ev) }

	// Ctrl+I: \x1b[105;5u (must be 'i' with Ctrl, not Tab!)
	p.Parse([]byte("\x1b[105;5u"), handler)
	if len(events) != 1 || events[0].Key.Type != KeyRune || events[0].Key.Rune != 'i' || !events[0].Key.HasCtrl() {
		t.Fatalf("Expected Ctrl+I rune event, got %+v", events)
	}

	// Ctrl+M: \x1b[109;5u (must be 'm' with Ctrl, not Enter!)
	events = events[:0]
	p.Parse([]byte("\x1b[109;5u"), handler)
	if len(events) != 1 || events[0].Key.Type != KeyRune || events[0].Key.Rune != 'm' || !events[0].Key.HasCtrl() {
		t.Fatalf("Expected Ctrl+M rune event, got %+v", events)
	}

	// Ctrl+[: \x1b[91;5u (must be '[' with Ctrl, not Esc!)
	events = events[:0]
	p.Parse([]byte("\x1b[91;5u"), handler)
	if len(events) != 1 || events[0].Key.Type != KeyRune || events[0].Key.Rune != '[' || !events[0].Key.HasCtrl() {
		t.Fatalf("Expected Ctrl+[ rune event, got %+v", events)
	}

	// Ctrl+H: \x1b[104;5u (must be 'h' with Ctrl, not Backspace!)
	events = events[:0]
	p.Parse([]byte("\x1b[104;5u"), handler)
	if len(events) != 1 || events[0].Key.Type != KeyRune || events[0].Key.Rune != 'h' || !events[0].Key.HasCtrl() {
		t.Fatalf("Expected Ctrl+H rune event, got %+v", events)
	}
}

func TestKittyFunctionalPUAKeys(t *testing.T) {
	p := NewParser()
	var events []Event
	handler := func(ev Event) { events = append(events, ev) }

	tests := []struct {
		seq      string
		expected KeyType
	}{
		{"\x1b[57348u", KeyUp},
		{"\x1b[57349u", KeyDown},
		{"\x1b[57346u", KeyLeft},
		{"\x1b[57347u", KeyRight},
		{"\x1b[57352u", KeyHome},
		{"\x1b[57353u", KeyEnd},
		{"\x1b[57350u", KeyPgUp},
		{"\x1b[57351u", KeyPgDown},
		{"\x1b[57344u", KeyInsert},
		{"\x1b[57345u", KeyDelete},
		{"\x1b[57358u", KeyCapsLock},
		{"\x1b[57359u", KeyScrollLock},
		{"\x1b[57360u", KeyNumLock},
		{"\x1b[57361u", KeyPrintScreen},
		{"\x1b[57362u", KeyPause},
		{"\x1b[57363u", KeyMenu},
		{"\x1b[57364u", KeyF1},
		{"\x1b[57375u", KeyF12},
		{"\x1b[57376u", KeyF13},
		{"\x1b[57387u", KeyF24},
		{"\x1b[57399u", KeyKp0},
		{"\x1b[57408u", KeyKp9},
		{"\x1b[57413u", KeyKpAdd},
		{"\x1b[57414u", KeyKpEnter},
		{"\x1b[57428u", KeyMediaPlay},
		{"\x1b[57429u", KeyMediaPause},
		{"\x1b[57441u", KeyLeftShift},
		{"\x1b[57442u", KeyLeftCtrl},
		{"\x1b[57444u", KeyLeftSuper},
	}

	for _, tc := range tests {
		events = events[:0]
		p.Parse([]byte(tc.seq), handler)
		if len(events) != 1 {
			t.Fatalf("For %q expected 1 event, got %d", tc.seq, len(events))
		}
		if events[0].Key.Type != tc.expected {
			t.Errorf("For %q expected KeyType %v, got %v", tc.seq, tc.expected, events[0].Key.Type)
		}
	}
}

func TestKittyEventTypes(t *testing.T) {
	p := NewParser()
	var events []Event
	handler := func(ev Event) { events = append(events, ev) }

	// Key Press: \x1b[97;1:1u
	p.Parse([]byte("\x1b[97;1:1u"), handler)
	if len(events) != 1 || !events[0].Key.IsPress() || events[0].Key.Action != KeyPress {
		t.Fatalf("Expected KeyPress, got %+v", events)
	}

	// Key Repeat: \x1b[97;1:2u
	events = events[:0]
	p.Parse([]byte("\x1b[97;1:2u"), handler)
	if len(events) != 1 || !events[0].Key.IsRepeat() || events[0].Key.Action != KeyRepeat {
		t.Fatalf("Expected KeyRepeat, got %+v", events)
	}

	// Key Release: \x1b[97;1:3u
	events = events[:0]
	p.Parse([]byte("\x1b[97;1:3u"), handler)
	if len(events) != 1 || !events[0].Key.IsRelease() || events[0].Key.Action != KeyRelease {
		t.Fatalf("Expected KeyRelease, got %+v", events)
	}

	// Legacy functional key with Kitty release: \x1b[1;5:3A (Up arrow, Ctrl, Release)
	events = events[:0]
	p.Parse([]byte("\x1b[1;5:3A"), handler)
	if len(events) != 1 || events[0].Key.Type != KeyUp || !events[0].Key.HasCtrl() || !events[0].Key.IsRelease() {
		t.Fatalf("Expected Up with Ctrl Release, got %+v", events)
	}

	// Delete key with Kitty release: \x1b[3:3~
	events = events[:0]
	p.Parse([]byte("\x1b[3:3~"), handler)
	if len(events) != 1 || events[0].Key.Type != KeyDelete || !events[0].Key.IsRelease() {
		t.Fatalf("Expected Delete Release, got %+v", events)
	}
}

func TestKittyAlternateKeysAndText(t *testing.T) {
	p := NewParser()
	var events []Event
	handler := func(ev Event) { events = append(events, ev) }

	// Russian 'я' with Shift: code 1103, shifted 1071 ('Я'), base 'a' (97), mod 2 (Shift)
	// \x1b[1103:1071:97;2u
	p.Parse([]byte("\x1b[1103:1071:97;2u"), handler)
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}
	k := events[0].Key
	if k.Rune != 1071 || k.ShiftedKey != 1071 || k.BaseKey != 97 || !k.HasShift() {
		t.Errorf("Unexpected alternate key parsing: %+v", k)
	}

	// Associated text: \x1b[97;1;104:101:108:108:111u ("hello")
	events = events[:0]
	p.Parse([]byte("\x1b[97;1;104:101:108:108:111u"), handler)
	if len(events) != 1 || events[0].Key.Text != "hello" {
		t.Errorf("Expected associated text 'hello', got %+v", events)
	}
}

func TestKittyQueryAndMode(t *testing.T) {
	p := NewParser()
	var events []Event
	handler := func(ev Event) { events = append(events, ev) }

	if p.KittySupported() {
		t.Error("Expected kittySupported to initially be false")
	}

	// Terminal responds to \x1b[?u with \x1b[?3u
	p.Parse([]byte("\x1b[?3u"), handler)
	if len(events) != 1 || events[0].Type != EventKittyMode || events[0].KittyFlags != 3 {
		t.Fatalf("Expected EventKittyMode with flags 3, got %+v", events)
	}
	if !p.KittySupported() || p.KittyFlags() != 3 {
		t.Errorf("Parser state not updated: supported=%v, flags=%d", p.KittySupported(), p.KittyFlags())
	}

	// Mode sequences should be ignored gracefully
	events = events[:0]
	p.Parse([]byte("\x1b[>1u\x1b[<u\x1b[=0u"), handler)
	if len(events) != 0 {
		t.Errorf("Expected 0 events for control sequences, got %d", len(events))
	}
}

func TestKittyKeyString(t *testing.T) {
	k1 := Key{Type: KeyRune, Rune: 'c', Mod: ModCtrl, Action: KeyPress}
	if k1.String() != "ctrl+c" {
		t.Errorf("Expected 'ctrl+c', got %q", k1.String())
	}

	k2 := Key{Type: KeyEnter, Mod: ModCtrl | ModShift, Action: KeyPress}
	if k2.String() != "ctrl+shift+enter" {
		t.Errorf("Expected 'ctrl+shift+enter', got %q", k2.String())
	}

	k3 := Key{Type: KeyUp, Mod: ModAlt, Action: KeyRelease}
	if k3.String() != "alt+up:release" {
		t.Errorf("Expected 'alt+up:release', got %q", k3.String())
	}

	// Keypad, media, and modifier keys should have descriptive names
	kKpAdd := Key{Type: KeyKpAdd}
	if kKpAdd.String() != "kp_add" {
		t.Errorf("Expected 'kp_add', got %q", kKpAdd.String())
	}

	kMediaPlay := Key{Type: KeyMediaPlay}
	if kMediaPlay.String() != "media_play" {
		t.Errorf("Expected 'media_play', got %q", kMediaPlay.String())
	}

	kLeftShift := Key{Type: KeyLeftShift}
	if kLeftShift.String() != "left_shift" {
		t.Errorf("Expected 'left_shift', got %q", kLeftShift.String())
	}

	kF35 := Key{Type: KeyF35}
	if kF35.String() != "f35" {
		t.Errorf("Expected 'f35', got %q", kF35.String())
	}
}

func TestKittyPureTextCodeZero(t *testing.T) {
	p := NewParser()
	var events []Event
	handler := func(ev Event) { events = append(events, ev) }

	// Pure text event from IME/OS with code 0: \x1b[0;;229u ('å')
	p.Parse([]byte("\x1b[0;;229u"), handler)
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got %d", len(events))
	}
	ev := events[0]
	if ev.Key.Type != KeyRune || ev.Key.Rune != 'å' || ev.Key.Text != "å" {
		t.Errorf("Expected rune 'å' (229), got %+v", ev.Key)
	}
}

func TestKittyExtendedFKeys(t *testing.T) {
	p := NewParser()
	var events []Event
	handler := func(ev Event) { events = append(events, ev) }

	// F25: \x1b[57388u
	p.Parse([]byte("\x1b[57388u"), handler)
	if len(events) != 1 || events[0].Key.Type != KeyF25 {
		t.Errorf("Expected KeyF25, got %+v", events)
	}

	// F35: \x1b[57398u
	events = events[:0]
	p.Parse([]byte("\x1b[57398u"), handler)
	if len(events) != 1 || events[0].Key.Type != KeyF35 {
		t.Errorf("Expected KeyF35, got %+v", events)
	}
}

func TestLegacyFunctionalKeysAndModifiers(t *testing.T) {
	p := NewParser()
	var events []Event
	handler := func(ev Event) { events = append(events, ev) }

	tests := []struct {
		seq     string
		keyType KeyType
		hasCtrl bool
		hasAlt  bool
		hasShft bool
	}{
		{"\x1b[1;2P", KeyF1, false, false, true},
		{"\x1b[1;5P", KeyF1, true, false, false},
		{"\x1b[1;3Q", KeyF2, false, true, false},
		{"\x1b[1;2R", KeyF3, false, false, true},
		{"\x1b[1;2S", KeyF4, false, false, true},
		{"\x1b[1;5E", KeyKpBegin, true, false, false},
		{"\x1b[11~", KeyF1, false, false, false},
		{"\x1b[12~", KeyF2, false, false, false},
		{"\x1b[13~", KeyF3, false, false, false},
		{"\x1b[14~", KeyF4, false, false, false},
		{"\x1b[15~", KeyF5, false, false, false},
		{"\x1b[29~", KeyMenu, false, false, false},
		{"\x1b[57427~", KeyKpBegin, false, false, false},
	}

	for _, tc := range tests {
		events = events[:0]
		p.Parse([]byte(tc.seq), handler)
		if len(events) != 1 {
			t.Fatalf("For %q expected 1 event, got %d", tc.seq, len(events))
		}
		k := events[0].Key
		if k.Type != tc.keyType {
			t.Errorf("For %q expected KeyType %v, got %v", tc.seq, tc.keyType, k.Type)
		}
		if tc.hasCtrl && !k.HasCtrl() {
			t.Errorf("For %q expected Ctrl modifier", tc.seq)
		}
		if tc.hasAlt && !k.HasAlt() {
			t.Errorf("For %q expected Alt modifier", tc.seq)
		}
		if tc.hasShft && !k.HasShift() {
			t.Errorf("For %q expected Shift modifier", tc.seq)
		}
	}
}

func TestKittySequencesBuilders(t *testing.T) {
	if KittyPushFlags(1) != "\x1b[>1u" {
		t.Errorf("KittyPushFlags failed: %q", KittyPushFlags(1))
	}
	if KittyPopFlags(1) != "\x1b[<u" {
		t.Errorf("KittyPopFlags(1) failed: %q", KittyPopFlags(1))
	}
	if KittyPopFlags(3) != "\x1b[<3u" {
		t.Errorf("KittyPopFlags(3) failed: %q", KittyPopFlags(3))
	}
	if KittySetFlags(1, 2) != "\x1b[=1;2u" {
		t.Errorf("KittySetFlags failed: %q", KittySetFlags(1, 2))
	}
	if KittyQuery() != "\x1b[?u" {
		t.Errorf("KittyQuery failed: %q", KittyQuery())
	}
	if KittyDisable() != "\x1b[<u\x1b[=0u" {
		t.Errorf("KittyDisable failed: %q", KittyDisable())
	}
}

