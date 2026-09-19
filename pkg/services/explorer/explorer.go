package explorer

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
	"github.com/baibeicha/goatui/pkg/driver/input"
	"github.com/baibeicha/goatui/pkg/layout"
	"github.com/baibeicha/goatui/pkg/media"
	"github.com/baibeicha/goatui/pkg/router"
	"github.com/baibeicha/goatui/pkg/style"
	"github.com/baibeicha/goatui/pkg/tea"
	"github.com/baibeicha/goatui/pkg/theme"
	"github.com/baibeicha/goatui/pkg/widgets"
	"github.com/baibeicha/goatui/pkg/window"
)

// PreviewType classifies the inspector preview mode for a selected item.
type PreviewType int

const (
	PreviewNone PreviewType = iota
	PreviewDirectory
	PreviewImage
	PreviewVideo
	PreviewText
	PreviewHex
)

// Entry represents a file or directory item with metadata for presentation and sorting.
type Entry struct {
	Name    string
	Path    string
	IsDir   bool
	Size    int64
	ModTime time.Time
	Mode    os.FileMode
	Ext     string
}

// Icon returns an emoji icon corresponding to file extension or directory status.
// Deprecated: Use GetIcon instead.
func (e Entry) Icon() string {
	if e.IsDir {
		return "[DIR]"
	}
	switch strings.ToLower(e.Ext) {
	case ".png", ".jpg", ".jpeg", ".webp", ".bmp", ".ico":
		return "[IMG]"
	case ".gif":
		return "[GIF]"
	case ".go", ".ts", ".js", ".py", ".rs", ".c", ".cpp", ".h", ".java":
		return "[SRC]"
	case ".json", ".yaml", ".yml", ".toml", ".xml", ".ini", ".env":
		return "[CFG]"
	case ".md", ".txt", ".rst", ".doc", ".pdf":
		return "[DOC]"
	case ".zip", ".tar", ".gz", ".7z", ".rar":
		return "[ZIP]"
	case ".exe", ".dll", ".so", ".bin":
		return "[BIN]"
	default:
		return "[FILE]"
	}
}

// FormatSize returns human-readable formatted file size.
func FormatSize(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024.0)
	} else if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024.0*1024.0))
	}
	return fmt.Sprintf("%.1f GB", float64(bytes)/(1024.0*1024.0*1024.0))
}

// FileExplorerScreen provides an interactive dual-pane file explorer with live media and text inspection.
type FileExplorerScreen struct {
	window.BaseScreen

	currentDir      string
	allEntries      []Entry
	filteredEntries []Entry
	selected        int
	offset          int

	// Interactive Address Bar state
	isEditingPath bool
	pathInput     *widgets.TextInput

	// Interactive In-place Search / Filter state
	isFiltering bool
	filterInput *widgets.TextInput

	// Preview state
	previewType PreviewType
	previewFile string
	imageWidget *media.ImageWidget
	videoPlayer *media.VideoPlayerWidget
	textLines   []string
	hexDump     []string
	dirItemCnt  int
	dirTotalSz  int64

	iconMode IconMode

	selectedMap     map[string]bool
	clipboard       []string
	clipboardIsCut  bool
	statusMsg       string
	listInnerBounds buffer.Rect

	lastWidth  int
	lastHeight int
}

// NewFileExplorerScreen creates a FileExplorerScreen rooted at the given initial directory.
// If dir is empty, the current working directory is used.
func NewFileExplorerScreen(initialDir string) *FileExplorerScreen {
	if initialDir == "" {
		if cwd, err := os.Getwd(); err == nil {
			initialDir = cwd
		} else {
			initialDir = "."
		}
	}
	initialDir = filepath.Clean(initialDir)

	pInput := widgets.NewTextInput()
	pInput.SetPrompt("Path: ")
	pInput.SetValue(initialDir)

	fInput := widgets.NewTextInput()
	fInput.SetPrompt("Search: ")
	fInput.SetPlaceholder("Type to filter files...")

	s := &FileExplorerScreen{
		currentDir:  initialDir,
		pathInput:   pInput,
		filterInput: fInput,
		imageWidget: media.NewImageWidget(),
		videoPlayer: media.NewVideoPlayer(),
		selectedMap: make(map[string]bool),
	}
	s.loadDirectory(initialDir)
	return s
}

// CurrentDir returns the currently active directory path.
func (s *FileExplorerScreen) CurrentDir() string {
	return s.currentDir
}

// OnMount handles incoming route parameters and queries (e.g. /explorer?path=... or file://... schemes).
func (s *FileExplorerScreen) OnMount(ctx *router.RouteContext) {
	if ctx == nil {
		return
	}

	target := ""
	if p, ok := ctx.Params["path"]; ok && p != "" {
		target = p
	} else if q, ok := ctx.Query["path"]; ok && q != "" {
		target = q
	} else if strings.HasPrefix(ctx.Path, "file://") {
		target = strings.TrimPrefix(ctx.Path, "file://")
		// Handle Windows file:///d:/...
		if strings.HasPrefix(target, "/") && len(target) > 2 && target[2] == ':' {
			target = target[1:]
		}
	}

	if target != "" {
		target = filepath.Clean(target)
		fi, err := os.Stat(target)
		if err == nil {
			if fi.IsDir() {
				s.loadDirectory(target)
			} else {
				dir := filepath.Dir(target)
				s.loadDirectory(dir)
				targetBase := filepath.Base(target)
				for i, e := range s.filteredEntries {
					if e.Name == targetBase {
						s.selected = i
						s.loadPreview(e)
						break
					}
				}
			}
		}
	}
}

// loadDirectory reads, filters, and sorts directory entries.
func (s *FileExplorerScreen) loadDirectory(dirPath string) {
	cleanDir := filepath.Clean(dirPath)
	dirEntries, err := os.ReadDir(cleanDir)
	if err != nil {
		s.allEntries = []Entry{{
			Name:  fmt.Sprintf("[Error reading directory: %v]", err),
			IsDir: false,
		}}
		s.filteredEntries = s.allEntries
		s.selected = 0
		s.offset = 0
		return
	}

	s.currentDir = cleanDir
	s.pathInput.SetValue(cleanDir)
	entries := make([]Entry, 0, len(dirEntries))

	for _, d := range dirEntries {
		info, err := d.Info()
		if err != nil {
			continue
		}
		entries = append(entries, Entry{
			Name:    d.Name(),
			Path:    filepath.Join(cleanDir, d.Name()),
			IsDir:   d.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
			Mode:    info.Mode(),
			Ext:     filepath.Ext(d.Name()),
		})
	}

	// Sort directories first, then alphabetical (case-insensitive)
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].IsDir != entries[j].IsDir {
			return entries[i].IsDir // Dirs first
		}
		return strings.ToLower(entries[i].Name) < strings.ToLower(entries[j].Name)
	})

	s.allEntries = entries
	s.applyFilter()
}

// applyFilter filters allEntries based on filterInput into filteredEntries.
func (s *FileExplorerScreen) applyFilter() {
	query := strings.TrimSpace(strings.ToLower(s.filterInput.Value()))
	if query == "" {
		s.filteredEntries = s.allEntries
	} else {
		filtered := make([]Entry, 0, len(s.allEntries))
		for _, e := range s.allEntries {
			if strings.Contains(strings.ToLower(e.Name), query) {
				filtered = append(filtered, e)
			}
		}
		s.filteredEntries = filtered
	}

	if s.selected >= len(s.filteredEntries) {
		s.selected = max(0, len(s.filteredEntries)-1)
	}
	s.offset = 0

	if len(s.filteredEntries) > 0 {
		s.loadPreview(s.filteredEntries[s.selected])
	} else {
		s.previewType = PreviewNone
	}
}

// loadPreview populates live inspector state for the selected file or directory.
func (s *FileExplorerScreen) loadPreview(entry Entry) {
	s.previewFile = entry.Path

	if entry.IsDir {
		s.previewType = PreviewDirectory
		subEntries, err := os.ReadDir(entry.Path)
		if err == nil {
			s.dirItemCnt = len(subEntries)
			var totalSz int64
			for _, sub := range subEntries {
				if info, err := sub.Info(); err == nil {
					totalSz += info.Size()
				}
			}
			s.dirTotalSz = totalSz
		} else {
			s.dirItemCnt = 0
			s.dirTotalSz = 0
		}
		return
	}

	ext := strings.ToLower(entry.Ext)

	// Animated GIF preview
	if ext == ".gif" {
		s.previewType = PreviewVideo
		if err := s.videoPlayer.LoadGIF(entry.Path); err == nil {
			s.videoPlayer.Play()
			return
		}
	}

	// Image preview
	if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".webp" {
		s.previewType = PreviewImage
		if err := s.imageWidget.LoadFile(entry.Path); err == nil {
			return
		}
	}

	// Text or Code preview
	f, err := os.Open(entry.Path)
	if err != nil {
		s.previewType = PreviewNone
		return
	}
	defer f.Close()

	// Read first 4KB to check if binary
	buf := make([]byte, 4096)
	n, _ := f.Read(buf)
	if n == 0 {
		s.previewType = PreviewText
		s.textLines = []string{"<Empty file>"}
		return
	}

	isBinary := false
	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			isBinary = true
			break
		}
	}

	if isBinary {
		s.previewType = PreviewHex
		s.hexDump = formatHexDump(buf[:n])
	} else {
		s.previewType = PreviewText
		s.textLines = strings.Split(string(buf[:n]), "\n")
		if len(s.textLines) > 100 {
			s.textLines = s.textLines[:100]
		}
	}
}

func formatHexDump(data []byte) []string {
	lines := make([]string, 0, (len(data)+15)/16)
	for i := 0; i < len(data); i += 16 {
		end := min(i+16, len(data))
		chunk := data[i:end]

		hexPart := hex.EncodeToString(chunk)
		var formattedHex strings.Builder
		for j := 0; j < len(hexPart); j += 2 {
			formattedHex.WriteString(hexPart[j : j+2])
			formattedHex.WriteByte(' ')
			if j == 14 {
				formattedHex.WriteByte(' ')
			}
		}
		// Pad hex string to 48 chars
		for formattedHex.Len() < 49 {
			formattedHex.WriteByte(' ')
		}

		var asciiPart strings.Builder
		for _, b := range chunk {
			if b >= 32 && b <= 126 {
				asciiPart.WriteByte(b)
			} else {
				asciiPart.WriteByte('.')
			}
		}

		lines = append(lines, fmt.Sprintf("%08x  %s |%s|", i, formattedHex.String(), asciiPart.String()))
		if len(lines) >= 60 {
			break
		}
	}
	return lines
}

// SelectedFiles returns all selected file paths, or the currently focused file if no multi-selection.
func (s *FileExplorerScreen) SelectedFiles() []string {
	var res []string
	for p := range s.selectedMap {
		res = append(res, p)
	}
	if len(res) == 0 && len(s.filteredEntries) > 0 && s.selected < len(s.filteredEntries) {
		res = append(res, s.filteredEntries[s.selected].Path)
	}
	return res
}

// ClearSelection clears the multi-selection map.
func (s *FileExplorerScreen) ClearSelection() {
	s.selectedMap = make(map[string]bool)
}

// CopySelection copies the selected files to clipboard.
func (s *FileExplorerScreen) CopySelection() {
	files := s.SelectedFiles()
	if len(files) == 0 {
		return
	}
	s.clipboard = files
	s.clipboardIsCut = false
	s.statusMsg = fmt.Sprintf("Copied %d item(s) to clipboard", len(files))
}

// CutSelection marks the selected files for moving.
func (s *FileExplorerScreen) CutSelection() {
	files := s.SelectedFiles()
	if len(files) == 0 {
		return
	}
	s.clipboard = files
	s.clipboardIsCut = true
	s.statusMsg = fmt.Sprintf("Cut %d item(s) to clipboard", len(files))
}

// PasteClipboard pastes files from clipboard into current directory.
func (s *FileExplorerScreen) PasteClipboard() {
	if len(s.clipboard) == 0 {
		s.statusMsg = "Clipboard is empty"
		return
	}
	count := 0
	for _, src := range s.clipboard {
		dst := filepath.Join(s.currentDir, filepath.Base(src))
		if src == dst {
			dst = filepath.Join(s.currentDir, "copy_"+filepath.Base(src))
		}
		if s.clipboardIsCut {
			if err := os.Rename(src, dst); err == nil {
				count++
			} else if errCopy := copyPath(src, dst); errCopy == nil {
				_ = os.RemoveAll(src)
				count++
			}
		} else {
			if err := copyPath(src, dst); err == nil {
				count++
			}
		}
	}
	if s.clipboardIsCut {
		s.clipboard = nil
	}
	s.statusMsg = fmt.Sprintf("Pasted %d item(s)", count)
	s.loadDirectory(s.currentDir)
}

// DeleteSelection deletes currently selected files or focused file.
func (s *FileExplorerScreen) DeleteSelection() {
	files := s.SelectedFiles()
	if len(files) == 0 {
		return
	}
	count := 0
	for _, f := range files {
		if err := os.RemoveAll(f); err == nil {
			count++
			delete(s.selectedMap, f)
		}
	}
	s.statusMsg = fmt.Sprintf("Deleted %d item(s)", count)
	s.loadDirectory(s.currentDir)
}

// CreateNewFolder creates a new folder in current directory.
func (s *FileExplorerScreen) CreateNewFolder(name string) {
	if name == "" {
		name = "new_folder"
	}
	target := filepath.Join(s.currentDir, name)
	idx := 1
	for {
		if _, err := os.Stat(target); os.IsNotExist(err) {
			break
		}
		target = filepath.Join(s.currentDir, fmt.Sprintf("%s_%d", name, idx))
		idx++
	}
	if err := os.MkdirAll(target, 0755); err == nil {
		s.statusMsg = fmt.Sprintf("Created: %s", filepath.Base(target))
		s.loadDirectory(s.currentDir)
	}
}

func copyPath(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return copyDir(src, dst)
	}
	return copyFile(src, dst)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		sChild := filepath.Join(src, entry.Name())
		dChild := filepath.Join(dst, entry.Name())
		if err := copyPath(sChild, dChild); err != nil {
			return err
		}
	}
	return nil
}

// IsEditingPath returns whether the address bar is currently focused and being edited.
func (s *FileExplorerScreen) IsEditingPath() bool {
	return s.isEditingPath
}

// IsFiltering returns whether the in-place search / filter is active.
func (s *FileExplorerScreen) IsFiltering() bool {
	return s.isFiltering
}

// Update handles interactive keyboard events and timer ticks.
func (s *FileExplorerScreen) Update(msg tea.Msg) (window.Screen, tea.Cmd) {
	switch m := msg.(type) {
	case tea.KeyMsg:
		// 1. If currently editing the Address Bar path
		if s.isEditingPath {
			switch m.Key.Type {
			case input.KeyEsc:
				s.isEditingPath = false
				s.pathInput.SetValue(s.currentDir)
				return s, nil
			case input.KeyEnter:
				s.isEditingPath = false
				newPath := strings.TrimSpace(s.pathInput.Value())
				if newPath != "" {
					// Handle relative or absolute paths
					resolved := newPath
					if !filepath.IsAbs(resolved) {
						resolved = filepath.Join(s.currentDir, resolved)
					}
					s.loadDirectory(resolved)
				}
				return s, nil
			default:
				s.pathInput.HandleKey(m.Key)
				return s, nil
			}
		}

		// 2. If currently searching/filtering files
		if s.isFiltering {
			switch m.Key.Type {
			case input.KeyEsc:
				s.isFiltering = false
				s.filterInput.SetValue("")
				s.applyFilter()
				return s, nil
			case input.KeyEnter:
				s.isFiltering = false
				if len(s.filteredEntries) > 0 && s.selected < len(s.filteredEntries) {
					sel := s.filteredEntries[s.selected]
					if sel.IsDir {
						s.loadDirectory(sel.Path)
					}
				}
				return s, nil
			case input.KeyUp:
				if s.selected > 0 {
					s.selected--
					if s.selected < s.offset {
						s.offset = s.selected
					}
					s.loadPreview(s.filteredEntries[s.selected])
				}
				return s, nil
			case input.KeyDown:
				if s.selected < len(s.filteredEntries)-1 {
					s.selected++
					s.loadPreview(s.filteredEntries[s.selected])
				}
				return s, nil
			default:
				if s.filterInput.HandleKey(m.Key) {
					s.applyFilter()
					return s, nil
				}
			}
		}

		// 3. Normal Explorer Navigation mode
		// Hotkeys for entering Address Bar editing: Ctrl+L, 'e', or 'l'
		if (m.Key.Type == input.KeyRune && m.Key.Rune == 'l' && m.Key.Mod.Has(cell.AttrBold)) || // Ctrl+L
			(m.Key.Type == input.KeyRune && (m.Key.Rune == 'e' || m.Key.Rune == 'E')) {
			s.isEditingPath = true
			s.pathInput.SetValue(s.currentDir)
			s.pathInput.Focus()
			return s, nil
		}

		// Hotkey for live search/filter: '/'
		if m.Key.Type == input.KeyRune && m.Key.Rune == '/' {
			s.isFiltering = true
			s.filterInput.Focus()
			return s, nil
		}

		switch m.Key.Type {
		case input.KeyUp:
			if s.selected > 0 {
				s.selected--
				if s.selected < s.offset {
					s.offset = s.selected
				}
				s.loadPreview(s.filteredEntries[s.selected])
			}
		case input.KeyDown:
			if s.selected < len(s.filteredEntries)-1 {
				s.selected++
				s.loadPreview(s.filteredEntries[s.selected])
			}
		case input.KeyEnter:
			if len(s.filteredEntries) > 0 && s.selected < len(s.filteredEntries) {
				sel := s.filteredEntries[s.selected]
				if sel.IsDir {
					s.loadDirectory(sel.Path)
				}
			}
		case input.KeyBackspace, input.KeyLeft:
			parent := filepath.Dir(s.currentDir)
			if parent != s.currentDir {
				s.loadDirectory(parent)
			}
		case input.KeyHome:
			s.selected = 0
			s.offset = 0
			if len(s.filteredEntries) > 0 {
				s.loadPreview(s.filteredEntries[0])
			}
		case input.KeyEnd:
			if len(s.filteredEntries) > 0 {
				s.selected = len(s.filteredEntries) - 1
				s.loadPreview(s.filteredEntries[s.selected])
			}
		case input.KeyPgUp:
			s.selected = max(0, s.selected-10)
			if s.selected < s.offset {
				s.offset = s.selected
			}
			if len(s.filteredEntries) > 0 {
				s.loadPreview(s.filteredEntries[s.selected])
			}
		case input.KeyPgDown:
			if len(s.filteredEntries) > 0 {
				s.selected = min(len(s.filteredEntries)-1, s.selected+10)
				s.loadPreview(s.filteredEntries[s.selected])
			}
		case input.KeySpace:
			if len(s.filteredEntries) > 0 && s.selected < len(s.filteredEntries) {
				curPath := s.filteredEntries[s.selected].Path
				s.selectedMap[curPath] = !s.selectedMap[curPath]
				if !s.selectedMap[curPath] {
					delete(s.selectedMap, curPath)
				}
				return s, nil
			}
		case input.KeyDelete:
			s.DeleteSelection()
			return s, nil
		case input.KeyRune:
			switch m.Key.Rune {
			case 'k':
				if s.selected > 0 {
					s.selected--
					if s.selected < s.offset {
						s.offset = s.selected
					}
					s.loadPreview(s.filteredEntries[s.selected])
				}
			case 'j':
				if s.selected < len(s.filteredEntries)-1 {
					s.selected++
					s.loadPreview(s.filteredEntries[s.selected])
				}
			case 'h':
				parent := filepath.Dir(s.currentDir)
				if parent != s.currentDir {
					s.loadDirectory(parent)
				}
			case 'r', 'R':
				s.loadDirectory(s.currentDir)
			case 'c', 'C':
				s.CopySelection()
			case 'x', 'X':
				s.CutSelection()
			case 'v', 'V':
				s.PasteClipboard()
			case 'd', 'D':
				s.DeleteSelection()
			case 'n', 'N':
				s.CreateNewFolder("new_folder")
			case 'p', 'P':
				if s.previewType == PreviewVideo {
					s.videoPlayer.Toggle()
				}
			case 'f', 'F':
				mode := (s.imageWidget.ScaleMode() + 1) % 3
				s.imageWidget.SetScaleMode(mode)
				s.videoPlayer.SetScaleMode(mode)
			case 'i', 'I':
				s.iconMode = s.iconMode.NextMode()
			}
		}

	case tea.MouseMsg:
		if m.Button == input.MouseWheelUp {
			if s.selected > 0 {
				s.selected--
				if s.selected < s.offset {
					s.offset = s.selected
				}
				s.loadPreview(s.filteredEntries[s.selected])
			}
			return s, nil
		}
		if m.Button == input.MouseWheelDown {
			if s.selected < len(s.filteredEntries)-1 {
				s.selected++
				s.loadPreview(s.filteredEntries[s.selected])
			}
			return s, nil
		}
		if m.Button == input.MouseLeft && m.Action == input.MousePress {
			if !s.listInnerBounds.IsEmpty() && s.listInnerBounds.Contains(m.X, m.Y) {
				relY := m.Y - s.listInnerBounds.Y
				clickedIdx := s.offset + relY
				if clickedIdx >= 0 && clickedIdx < len(s.filteredEntries) {
					if s.selected == clickedIdx && s.filteredEntries[clickedIdx].IsDir {
						// Second click on directory enters it
						s.loadDirectory(s.filteredEntries[clickedIdx].Path)
					} else {
						s.selected = clickedIdx
						s.loadPreview(s.filteredEntries[clickedIdx])
					}
					return s, nil
				}
			}
		}

	case time.Time:
		// Tick event for video/gif playback
		if s.previewType == PreviewVideo && s.videoPlayer.IsPlaying() {
			s.videoPlayer.Advance(33 * time.Millisecond) // ~30 FPS
		}
	}

	return s, nil
}

// View draws the complete dual-pane File Explorer layout across the whole terminal screen.
func (s *FileExplorerScreen) View(f *tea.Frame) {
	s.DrawInArea(f, f.Area())
}

// DrawInArea renders the File Explorer inside the specified bounding rectangle.
// Completely clears the background on every frame to prevent ghosting or residual text.
func (s *FileExplorerScreen) DrawInArea(f *tea.Frame, area buffer.Rect) {
	if area.IsEmpty() {
		return
	}

	s.lastWidth = area.Width
	s.lastHeight = area.Height

	curTheme := theme.Current()
	p := curTheme.Colors

	// GUARANTEED ZERO GHOSTING: clear the entire container area with theme background
	f.Buffer.Fill(area, cell.Cell{
		Rune:   ' ',
		Width:  1,
		FgType: cell.ColorDefault,
		BgType: p.Background.Type,
		Bg:     p.Background.Value,
	})

	// 1. Address Bar Element (3 rows tall, bounded with box/border)
	addressArea := buffer.NewRect(area.X, area.Y, area.Width, 3)

	addrBorderColor := p.Secondary
	addrTitle := " Location (Press [E] to Edit, [/] to Search) "
	if s.isEditingPath {
		addrBorderColor = p.Primary
		addrTitle = " Edit Address (Enter=Navigate, Esc=Cancel) "
	}

	addrBox := style.NewStyle().
		Border(curTheme.Borders.Border).
		BorderForeground(addrBorderColor).
		Background(p.Background)
	addrBox.Draw(f.Buffer, addressArea, addrTitle)

	innerAddr := addressArea.Inset(1, 1)
	if !innerAddr.IsEmpty() {
		// Clear address bar inner line
		for cx := innerAddr.X; cx < innerAddr.Right(); cx++ {
			f.Buffer.Set(cx, innerAddr.Y, cell.Cell{
				Rune:   ' ',
				Width:  1,
				FgType: cell.ColorDefault,
				BgType: p.Background.Type,
				Bg:     p.Background.Value,
			})
		}

		if s.isEditingPath {
			s.pathInput.Draw(f.Buffer, innerAddr)
		} else {
			disp := s.currentDir
			hints := "[E] Edit  [/] Search  [Enter] Open  [Backspace] Up  [R] Reload"
			hintsW := buffer.StringWidth(hints)

			showHints := innerAddr.Width > hintsW+20
			maxDisp := innerAddr.Width - 2
			if showHints {
				maxDisp = innerAddr.Width - hintsW - 4
			}
			if maxDisp > 4 && buffer.StringWidth(disp) > maxDisp {
				runes := []rune(disp)
				targetW := maxDisp - 3
				w := 0
				startIdx := len(runes)
				for i := len(runes) - 1; i >= 0; i-- {
					rw := buffer.RuneWidth(runes[i])
					if rw == 0 {
						rw = 1
					}
					if w+rw > targetW {
						break
					}
					w += rw
					startIdx = i
				}
				disp = "..." + string(runes[startIdx:])
			}
			addrArea := buffer.NewRect(innerAddr.X+1, innerAddr.Y, max(0, maxDisp), 1)
			f.Buffer.SetStringAligned(addrArea, disp, buffer.AlignLeft, p.Foreground, p.Background, cell.AttrBold)

			if showHints {
				hintsX := innerAddr.Right() - hintsW - 1
				f.Buffer.SetString(hintsX, innerAddr.Y, hints, p.Muted, p.Background, cell.AttrDim)
			}
		}
	}

	// 2. Main Dual-Pane split below address bar
	bodyArea := buffer.NewRect(area.X, area.Y+3, area.Width, max(0, area.Height-3))
	if bodyArea.Height <= 0 {
		return
	}

	cols := layout.SplitHorizontal(bodyArea,
		layout.Percent(45),
		layout.Percent(55),
	)

	s.drawFileList(f.Buffer, cols[0], curTheme)
	s.drawPreview(f.Buffer, cols[1], curTheme)
}

func (s *FileExplorerScreen) drawFileList(buf *buffer.Buffer, area buffer.Rect, th theme.Theme) {
	p := th.Colors

	title := fmt.Sprintf(" Explorer (%d items) ", len(s.filteredEntries))
	if s.isFiltering {
		title = fmt.Sprintf(" Explorer [Search: %s (%d matches)] ", s.filterInput.Value(), len(s.filteredEntries))
	}

	borderColor := p.Secondary
	if s.isFiltering {
		borderColor = p.Accent
	}

	box := style.NewStyle().
		Border(th.Borders.Border).
		BorderForeground(borderColor)
	box.Draw(buf, area, title)

	inner := area.Inset(1, 1)
	if inner.IsEmpty() {
		return
	}

	// CRITICAL BUGFIX: Pre-clear the ENTIRE inner list area with spaces!
	// This completely prevents ghost text from previous directories or files!
	for cy := inner.Y; cy < inner.Bottom(); cy++ {
		for cx := inner.X; cx < inner.Right(); cx++ {
			buf.Set(cx, cy, cell.Cell{
				Rune:   ' ',
				Width:  1,
				FgType: cell.ColorDefault,
				BgType: p.Background.Type,
				Bg:     p.Background.Value,
			})
		}
	}

	s.listInnerBounds = inner

	startY := inner.Y
	availHeight := inner.Height

	// If live filtering is active, render search input on first row
	if s.isFiltering {
		filterRow := buffer.NewRect(inner.X, inner.Y, inner.Width, 1)
		s.filterInput.Draw(buf, filterRow)
		startY++
		availHeight--
		if availHeight <= 0 {
			return
		}
	}

	// Reserve 1 row at bottom for operations status / hotkey help if height permits
	hasFooter := availHeight >= 4
	if hasFooter {
		availHeight--
	}

	visibleRows := availHeight
	if s.selected >= s.offset+visibleRows {
		s.offset = s.selected - visibleRows + 1
	}
	if s.selected < s.offset {
		s.offset = s.selected
	}

	for i := 0; i < visibleRows; i++ {
		idx := s.offset + i
		y := startY + i
		if idx >= len(s.filteredEntries) {
			break
		}

		entry := s.filteredEntries[idx]
		isSelected := (idx == s.selected)
		isMultiChecked := s.selectedMap[entry.Path]

		fg := p.Foreground
		bg := p.Background
		mod := cell.AttrNone

		if isSelected {
			fg = p.Background
			bg = p.Primary
			mod = cell.AttrBold

			// Fill entire row width with selected background color
			for cx := inner.X; cx < inner.Right(); cx++ {
				buf.Set(cx, y, cell.Cell{
					Rune:   ' ',
					Width:  1,
					FgType: fg.Type,
					BgType: bg.Type,
					Fg:     fg.Value,
					Bg:     bg.Value,
				})
			}
		} else if entry.IsDir {
			fg = p.Accent
		}

		colX := inner.X + 1

		// Checkbox for selection: [x] or [ ]
		checkStr := "[ ] "
		if isMultiChecked {
			checkStr = "[x] "
		}
		buf.SetString(colX, y, checkStr, p.Secondary, bg, mod)
		colX += 4

		// Icon
		icon := GetIcon(s.iconMode, entry.IsDir, entry.Ext)
		buf.SetString(colX, y, icon, fg, bg, mod)
		colX += IconWidth(s.iconMode) + 1

		nameMaxW := inner.Right() - colX - 9
		name := entry.Name
		if buffer.StringWidth(name) > nameMaxW && nameMaxW > 3 {
			runes := []rune(name)
			targetW := nameMaxW - 3
			curW := 0
			endRune := 0
			for i, r := range runes {
				rw := buffer.RuneWidth(r)
				if rw == 0 {
					rw = 1
				}
				if curW+rw > targetW {
					break
				}
				curW += rw
				endRune = i + 1
			}
			name = string(runes[:endRune]) + "..."
		}
		nameArea := buffer.NewRect(colX, y, max(0, nameMaxW), 1)
		buf.SetStringAligned(nameArea, name, buffer.AlignLeft, fg, bg, mod)

		// File size aligned right
		sizeStr := ""
		if !entry.IsDir {
			sizeStr = FormatSize(entry.Size)
		} else {
			sizeStr = "<DIR>"
		}
		sizeX := inner.Right() - buffer.StringWidth(sizeStr) - 1
		if sizeX > colX+buffer.StringWidth(name) {
			buf.SetString(sizeX, y, sizeStr, p.Secondary, bg, cell.AttrDim)
		}
	}

	// Render bottom status bar
	if hasFooter {
		statusY := inner.Bottom() - 1
		statusText := s.statusMsg
		statusCol := p.Success
		if statusText == "" {
			statusText = "[Space] Select  [C]opy  [X]Cut  [V]Paste  [D]el  [N]ew"
			statusCol = p.Muted
		}
		buf.SetString(inner.X+1, statusY, statusText, statusCol, p.Background, cell.AttrDim)
	}
}

func (s *FileExplorerScreen) drawPreview(buf *buffer.Buffer, area buffer.Rect, th theme.Theme) {
	p := th.Colors
	title := " Live Inspector "
	if s.previewFile != "" {
		base := filepath.Base(s.previewFile)
		if buffer.StringWidth(base) > 24 {
			runes := []rune(base)
			targetW := 21
			curW := 0
			endRune := 0
			for i, r := range runes {
				rw := buffer.RuneWidth(r)
				if rw == 0 {
					rw = 1
				}
				if curW+rw > targetW {
					break
				}
				curW += rw
				endRune = i + 1
			}
			base = string(runes[:endRune]) + "..."
		}
		title = fmt.Sprintf(" Preview: %s ", base)
	}

	box := style.NewStyle().
		Border(th.Borders.Border).
		BorderForeground(p.Secondary)
	box.Draw(buf, area, title)

	inner := area.Inset(1, 1)
	if inner.IsEmpty() {
		return
	}

	// CRITICAL BUGFIX: Pre-clear the ENTIRE inner preview area with background cells!
	// Eliminates dangling vertical bars '|' and old frame fragments.
	for cy := inner.Y; cy < inner.Bottom(); cy++ {
		for cx := inner.X; cx < inner.Right(); cx++ {
			buf.Set(cx, cy, cell.Cell{
				Rune:   ' ',
				Width:  1,
				FgType: cell.ColorDefault,
				BgType: p.Background.Type,
				Bg:     p.Background.Value,
			})
		}
	}

	switch s.previewType {
	case PreviewDirectory:
		buf.SetString(inner.X+2, inner.Y+1, "Directory Information", p.Accent, p.Background, cell.AttrBold)
		buf.SetString(inner.X+2, inner.Y+3, fmt.Sprintf("• Path: %s", s.previewFile), p.Foreground, p.Background, 0)
		buf.SetString(inner.X+2, inner.Y+4, fmt.Sprintf("• Total Items: %d", s.dirItemCnt), p.Foreground, p.Background, 0)
		buf.SetString(inner.X+2, inner.Y+5, fmt.Sprintf("• Total Content Size: %s", FormatSize(s.dirTotalSz)), p.Foreground, p.Background, 0)
		buf.SetString(inner.X+2, inner.Y+7, "Press [Enter] to browse inside this directory.", p.Primary, p.Background, cell.AttrItalic)

	case PreviewImage:
		// Render image with media.ImageWidget
		s.imageWidget.Draw(buf, inner)

	case PreviewVideo:
		// Render animated GIF with media.VideoPlayerWidget
		s.videoPlayer.Draw(buf, inner)
		status := "PAUSED"
		if s.videoPlayer.IsPlaying() {
			status = "PLAYING"
		}
		info := fmt.Sprintf("[%s Frame %d/%d - Press 'P' to toggle]",
			status, s.videoPlayer.CurrentFrame()+1, s.videoPlayer.TotalFrames())
		buf.SetString(inner.X+1, inner.Y+inner.Height-1, info, p.Accent, p.Background, cell.AttrBold)

	case PreviewText:
		buf.SetString(inner.X+1, inner.Y, "Text / Source Preview (First 100 lines):", p.Accent, p.Background, cell.AttrBold)
		textInner := buffer.NewRect(inner.X, inner.Y+2, inner.Width, inner.Height-2)
		for i := 0; i < textInner.Height && i < len(s.textLines); i++ {
			lineNum := fmt.Sprintf("%3d │ ", i+1)
			lineNumW := buffer.StringWidth(lineNum)
			buf.SetString(textInner.X+1, textInner.Y+i, lineNum, p.Secondary, p.Background, cell.AttrDim)
			lineContent := s.textLines[i]
			contentX := textInner.X + 1 + lineNumW
			maxLen := textInner.Right() - contentX - 1
			if maxLen > 0 {
				lineArea := buffer.NewRect(contentX, textInner.Y+i, maxLen, 1)
				buf.SetStringAligned(lineArea, lineContent, buffer.AlignLeft, p.Foreground, p.Background, 0)
			}
		}

	case PreviewHex:
		buf.SetString(inner.X+1, inner.Y, "Binary HEX Dump:", p.Accent, p.Background, cell.AttrBold)
		hexInner := buffer.NewRect(inner.X, inner.Y+2, inner.Width, inner.Height-2)
		for i := 0; i < hexInner.Height && i < len(s.hexDump); i++ {
			buf.SetString(hexInner.X+1, hexInner.Y+i, s.hexDump[i], p.Foreground, p.Background, 0)
		}

	default:
		buf.SetString(inner.X+2, inner.Y+2, "No preview available for selected item.", p.Secondary, p.Background, cell.AttrDim)
	}
}

// IconMode returns the current icon mode.
func (s *FileExplorerScreen) IconMode() IconMode {
	return s.iconMode
}

// SetIconMode sets the active icon mode.
func (s *FileExplorerScreen) SetIconMode(mode IconMode) {
	s.iconMode = mode
}
