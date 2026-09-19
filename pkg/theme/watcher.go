package theme

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/baibeicha/goatui/pkg/tea"
)

// ThemeChangedMsg is emitted when a theme is modified or switched.
type ThemeChangedMsg struct {
	Theme Theme
}

// ThemeManager coordinates registered themes, active theme selection, and hot-reload.
type ThemeManager struct {
	mu       sync.RWMutex
	current  Theme
	themes   map[string]Theme
	watchers []chan struct{}
}

var (
	defaultManagerOnce sync.Once
	defaultManager     *ThemeManager
)

// Default returns the global singleton ThemeManager.
func Default() *ThemeManager {
	defaultManagerOnce.Do(func() {
		defaultManager = NewThemeManager()
	})
	return defaultManager
}

// Current returns the active global theme.
func Current() Theme {
	return Default().Current()
}

// NewThemeManager creates a new theme manager with default presets registered.
func NewThemeManager() *ThemeManager {
	tm := &ThemeManager{
		current: SynodicDark,
		themes:  make(map[string]Theme),
	}

	tm.Register(SynodicDark)
	tm.Register(GoatDark)
	tm.Register(Dracula)
	tm.Register(CatppuccinMocha)
	tm.Register(Nord)
	tm.Register(Monokai)
	tm.Register(Cyberpunk)
	tm.Register(Matrix)
	tm.Register(Forest)

	return tm
}

func normalizeThemeKey(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "_", "")
	return s
}

// Register adds a theme to the manager's library.
func (tm *ThemeManager) Register(t Theme) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.themes[strings.ToLower(t.Name)] = t
	tm.themes[normalizeThemeKey(t.Name)] = t
}

// SetTheme switches the active theme by name.
func (tm *ThemeManager) SetTheme(name string) (Theme, bool) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if t, ok := tm.themes[strings.ToLower(name)]; ok {
		tm.current = t
		return t, true
	}
	if t, ok := tm.themes[normalizeThemeKey(name)]; ok {
		tm.current = t
		return t, true
	}
	return tm.current, false
}

// Current returns the active theme in a thread-safe manner.
func (tm *ThemeManager) Current() Theme {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.current
}

// AvailableThemes returns a slice of all registered unique theme names.
func (tm *ThemeManager) AvailableThemes() []string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	seen := make(map[string]bool)
	res := make([]string, 0, len(tm.themes))
	for _, t := range tm.themes {
		if !seen[t.Name] {
			seen[t.Name] = true
			res = append(res, t.Name)
		}
	}
	return res
}

// LoadFile parses a YAML theme file from disk and registers it.
func (tm *ThemeManager) LoadFile(filePath string) (Theme, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return Theme{}, err
	}

	t, err := ParseThemeYAML(string(data))
	if err != nil {
		return Theme{}, err
	}

	if t.Name == "" {
		base := filepath.Base(filePath)
		t.Name = strings.TrimSuffix(base, filepath.Ext(base))
	}

	tm.Register(t)
	return t, nil
}

// LoadDirectory scans a folder for .yaml/.yml files and registers all found themes.
func (tm *ThemeManager) LoadDirectory(dir string) ([]Theme, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var loaded []Theme
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext == ".yaml" || ext == ".yml" {
			path := filepath.Join(dir, entry.Name())
			if t, err := tm.LoadFile(path); err == nil {
				loaded = append(loaded, t)
			}
		}
	}
	return loaded, nil
}

// WatchDirectory polls a themes directory in the background and hot-reloads themes on changes.
// Returns a cancel function to stop the background watcher.
func (tm *ThemeManager) WatchDirectory(dir string, interval time.Duration, onChange func(Theme)) func() {
	stopChan := make(chan struct{})
	tm.mu.Lock()
	tm.watchers = append(tm.watchers, stopChan)
	tm.mu.Unlock()

	lastModMap := make(map[string]time.Time)

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-stopChan:
				return
			case <-ticker.C:
				entries, err := os.ReadDir(dir)
				if err != nil {
					continue
				}

				for _, e := range entries {
					if e.IsDir() {
						continue
					}
					ext := strings.ToLower(filepath.Ext(e.Name()))
					if ext != ".yaml" && ext != ".yml" {
						continue
					}

					fullPath := filepath.Join(dir, e.Name())
					info, err := e.Info()
					if err != nil {
						continue
					}

					modTime := info.ModTime()
					lastMod, exists := lastModMap[fullPath]
					if !exists || modTime.After(lastMod) {
						lastModMap[fullPath] = modTime
						if t, err := tm.LoadFile(fullPath); err == nil {
							// If current theme matches modified theme, update active theme
							if strings.EqualFold(tm.Current().Name, t.Name) {
								tm.SetTheme(t.Name)
								if onChange != nil {
									onChange(t)
								}
							}
						}
					}
				}
			}
		}
	}()

	return func() {
		close(stopChan)
	}
}

// SwitchTheme returns a tea.Cmd that dispatches ThemeChangedMsg into the GoatUI event loop.
func SwitchTheme(name string) tea.Cmd {
	return func() tea.Msg {
		t, ok := Default().SetTheme(name)
		if ok {
			return ThemeChangedMsg{Theme: t}
		}
		return nil
	}
}
