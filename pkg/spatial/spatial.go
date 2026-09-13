package spatial

import (
	"sort"
	"sync"

	"github.com/baibeicha/goatui/pkg/core/buffer"
)

// HitTarget represents an interactive screen region registered during frame rendering.
type HitTarget struct {
	ID       string
	Area     buffer.Rect
	ZIndex   int
	UserData any
}

// SpatialMap tracks interactive rectangular areas on screen and routes mouse interactions.
type SpatialMap struct {
	mu      sync.RWMutex
	targets []HitTarget
}

// NewSpatialMap creates a new empty spatial map.
func NewSpatialMap() *SpatialMap {
	return &SpatialMap{
		targets: make([]HitTarget, 0, 64),
	}
}

// Reset clears all registered targets while retaining underlying slice capacity.
func (sm *SpatialMap) Reset() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.targets = sm.targets[:0]
}

// Register adds an interactive rectangular target with a given ID and Z-Index.
func (sm *SpatialMap) Register(id string, area buffer.Rect, zIndex int, userData any) {
	if area.IsEmpty() {
		return
	}
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.targets = append(sm.targets, HitTarget{
		ID:       id,
		Area:     area,
		ZIndex:   zIndex,
		UserData: userData,
	})
}

// HitTest finds the topmost target (highest Z-index) containing coordinates (x, y).
func (sm *SpatialMap) HitTest(x, y int) (HitTarget, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	bestIdx := -1
	bestZ := -1000000

	for i := len(sm.targets) - 1; i >= 0; i-- {
		t := &sm.targets[i]
		if t.Area.Contains(x, y) {
			if t.ZIndex > bestZ {
				bestZ = t.ZIndex
				bestIdx = i
			}
		}
	}

	if bestIdx != -1 {
		return sm.targets[bestIdx], true
	}
	return HitTarget{}, false
}

// HitTestAll returns all registered targets containing coordinates (x, y),
// ordered from highest Z-index (front) to lowest (back).
func (sm *SpatialMap) HitTestAll(x, y int) []HitTarget {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var hits []HitTarget
	for _, t := range sm.targets {
		if t.Area.Contains(x, y) {
			hits = append(hits, t)
		}
	}
	if len(hits) > 1 {
		sort.SliceStable(hits, func(i, j int) bool {
			return hits[i].ZIndex > hits[j].ZIndex
		})
	}
	return hits
}

// Targets returns a snapshot copy of all registered targets.
func (sm *SpatialMap) Targets() []HitTarget {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return append([]HitTarget(nil), sm.targets...)
}
