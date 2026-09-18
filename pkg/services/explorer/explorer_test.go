package explorer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/driver/input"
	"github.com/baibeicha/goatui/pkg/router"
	"github.com/baibeicha/goatui/pkg/tea"
)

func TestFormatSize(t *testing.T) {
	cases := []struct {
		bytes    int64
		expected string
	}{
		{500, "500 B"},
		{1024, "1.0 KB"},
		{2048, "2.0 KB"},
		{1024 * 1024 * 5, "5.0 MB"},
		{1024 * 1024 * 1024 * 3, "3.0 GB"},
	}

	for _, c := range cases {
		if got := FormatSize(c.bytes); got != c.expected {
			t.Errorf("FormatSize(%d) = %s, expected %s", c.bytes, got, c.expected)
		}
	}
}

func TestExplorerNavigationAndPreview(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "goatui_explorer_test")
	if err != nil {
		t.Fatalf("MkdirTemp failed: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create subfolder and test files
	subDir := filepath.Join(tmpDir, "subdir")
	_ = os.Mkdir(subDir, 0755)
	_ = os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("Hello GoatUI Explorer!\nLine 2"), 0644)
	_ = os.WriteFile(filepath.Join(tmpDir, "binary.dat"), []byte{0x00, 0xFF, 0xAA, 0x55}, 0644)

	exp := NewFileExplorerScreen(tmpDir)
	if len(exp.allEntries) < 3 {
		t.Errorf("Expected at least 3 entries, got %d", len(exp.allEntries))
	}

	// First item should be the directory 'subdir' because directories are sorted first
	if !exp.filteredEntries[0].IsDir || exp.filteredEntries[0].Name != "subdir" {
		t.Errorf("Expected first item to be directory 'subdir', got %+v", exp.filteredEntries[0])
	}

	// Test OnMount with route context
	ctx := &router.RouteContext{
		Path:   "file:///" + filepath.ToSlash(filepath.Join(tmpDir, "test.txt")),
		Params: map[string]string{},
		Query:  map[string]string{},
	}
	exp.OnMount(ctx)

	if exp.previewType != PreviewText {
		t.Errorf("Expected PreviewText after mounting test.txt, got %d", exp.previewType)
	}

	// Test Key Navigation Down
	exp.Update(tea.KeyMsg{Key: input.Key{Type: input.KeyDown}})

	// Test Live Search Filter: trigger '/' and type 'bin'
	exp.Update(tea.KeyMsg{Key: input.Key{Type: input.KeyRune, Rune: '/'}})
	if !exp.isFiltering {
		t.Errorf("Expected isFiltering to be true after pressing '/'")
	}
	exp.Update(tea.KeyMsg{Key: input.Key{Type: input.KeyRune, Rune: 'b'}})
	exp.Update(tea.KeyMsg{Key: input.Key{Type: input.KeyRune, Rune: 'i'}})
	exp.Update(tea.KeyMsg{Key: input.Key{Type: input.KeyRune, Rune: 'n'}})

	if len(exp.filteredEntries) != 1 || exp.filteredEntries[0].Name != "binary.dat" {
		t.Errorf("Expected 1 filtered entry 'binary.dat', got %d: %+v", len(exp.filteredEntries), exp.filteredEntries)
	}

	// Exit search with Esc
	exp.Update(tea.KeyMsg{Key: input.Key{Type: input.KeyEsc}})
	if exp.isFiltering || len(exp.filteredEntries) < 3 {
		t.Errorf("Expected filter to be cleared after Esc")
	}

	// Test Address Bar Editing: trigger 'e' and enter subfolder
	exp.Update(tea.KeyMsg{Key: input.Key{Type: input.KeyRune, Rune: 'e'}})
	if !exp.isEditingPath {
		t.Errorf("Expected isEditingPath to be true after pressing 'e'")
	}
	exp.pathInput.SetValue(subDir)
	exp.Update(tea.KeyMsg{Key: input.Key{Type: input.KeyEnter}})
	if exp.isEditingPath {
		t.Errorf("Expected isEditingPath to be false after Enter")
	}
	if filepath.Clean(exp.CurrentDir()) != filepath.Clean(subDir) {
		t.Errorf("Expected CurrentDir to be %s, got %s", subDir, exp.CurrentDir())
	}

	// Test View rendering into buffer
	buf := buffer.NewBuffer(80, 24)
	frame := &tea.Frame{
		Buffer: buf,
	}
	exp.View(frame)

	// Ensure top border of address bar rendered
	c := buf.Cell(0, 0)
	if c == nil || c.Rune == ' ' {
		t.Errorf("Expected rendered address bar border at (0,0)")
	}
}

func TestExplorerIconModes(t *testing.T) {
	exp := NewFileExplorerScreen(".")

	// Default should be ASCII
	if exp.IconMode() != IconModeASCII {
		t.Errorf("Expected default IconModeASCII, got %v", exp.IconMode())
	}

	// Verify ASCII icon
	dirIcon := GetIcon(IconModeASCII, true, "")
	if dirIcon != "[DIR]" {
		t.Errorf("Expected '[DIR]', got %q", dirIcon)
	}
	srcIcon := GetIcon(IconModeASCII, false, ".go")
	if srcIcon != "[SRC]" {
		t.Errorf("Expected '[SRC]', got %q", srcIcon)
	}
	if IconWidth(IconModeASCII) != 5 {
		t.Errorf("Expected ASCII IconWidth = 5, got %d", IconWidth(IconModeASCII))
	}

	// Cycle mode using 'i' key
	exp.Update(tea.KeyMsg{Key: input.Key{Type: input.KeyRune, Rune: 'i'}})
	if exp.IconMode() != IconModeUnicode {
		t.Errorf("Expected IconModeUnicode after pressing 'i', got %v", exp.IconMode())
	}

	exp.Update(tea.KeyMsg{Key: input.Key{Type: input.KeyRune, Rune: 'i'}})
	if exp.IconMode() != IconModeNerdFont {
		t.Errorf("Expected IconModeNerdFont, got %v", exp.IconMode())
	}

	exp.Update(tea.KeyMsg{Key: input.Key{Type: input.KeyRune, Rune: 'i'}})
	if exp.IconMode() != IconModeEmoji {
		t.Errorf("Expected IconModeEmoji, got %v", exp.IconMode())
	}

	exp.Update(tea.KeyMsg{Key: input.Key{Type: input.KeyRune, Rune: 'i'}})
	if exp.IconMode() != IconModeASCII {
		t.Errorf("Expected cycle back to IconModeASCII, got %v", exp.IconMode())
	}
}

func TestExplorerFileOperationsAndMultiSelect(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "goatui_ops_test")
	if err != nil {
		t.Fatalf("MkdirTemp failed: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	f1 := filepath.Join(tmpDir, "file1.txt")
	f2 := filepath.Join(tmpDir, "file2.txt")
	_ = os.WriteFile(f1, []byte("Content 1"), 0644)
	_ = os.WriteFile(f2, []byte("Content 2"), 0644)

	exp := NewFileExplorerScreen(tmpDir)

	// Test multi-select with Space
	if len(exp.selectedMap) != 0 {
		t.Errorf("Expected empty selectedMap initially")
	}
	exp.Update(tea.KeyMsg{Key: input.Key{Type: input.KeySpace}})
	if len(exp.selectedMap) != 1 {
		t.Errorf("Expected 1 selected file after Space, got %d", len(exp.selectedMap))
	}

	// Test copy & paste
	exp.CopySelection()
	if len(exp.clipboard) != 1 {
		t.Errorf("Expected 1 clipboard item")
	}

	// Paste in same directory should create copy_file1.txt
	exp.PasteClipboard()
	copiedFile := filepath.Join(tmpDir, "copy_file1.txt")
	if _, err := os.Stat(copiedFile); os.IsNotExist(err) {
		t.Errorf("Expected pasted file %s to exist", copiedFile)
	}

	// Test CreateNewFolder
	exp.CreateNewFolder("my_new_dir")
	newDir := filepath.Join(tmpDir, "my_new_dir")
	if info, err := os.Stat(newDir); os.IsNotExist(err) || !info.IsDir() {
		t.Errorf("Expected created dir %s to exist", newDir)
	}

	// Test DeleteSelection
	exp.selectedMap = map[string]bool{copiedFile: true}
	exp.DeleteSelection()
	if _, err := os.Stat(copiedFile); !os.IsNotExist(err) {
		t.Errorf("Expected copiedFile to be deleted")
	}
}

func TestExplorerUnicodeSafety(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "goatui_unicode_test")
	if err != nil {
		t.Fatalf("MkdirTemp failed: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create long Cyrillic and CJK files
	cyrFile := filepath.Join(tmpDir, "Очень_Длинный_Файл_С_Русскими_Буквами_2026.txt")
	cyrContent := "Строка 1: Привет, Мир! Проверка рендеринга русских букв без паники и разделения байт UTF-8.\nСтрока 2: Вторая длинная строка для проверки ограничения ширины текста."
	_ = os.WriteFile(cyrFile, []byte(cyrContent), 0644)

	exp := NewFileExplorerScreen(tmpDir)
	frame := &tea.Frame{
		Buffer: buffer.NewBuffer(40, 15),
	}

	// Rendering in narrow 40x15 terminal area should not panic or split multi-byte characters
	exp.View(frame)

	// Mount the unicode file and view
	ctx := &router.RouteContext{
		Path:   "file:///" + filepath.ToSlash(cyrFile),
		Params: map[string]string{},
		Query:  map[string]string{},
	}
	exp.OnMount(ctx)
	exp.View(frame)
}
