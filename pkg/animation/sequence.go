package animation

import (
	"time"
)

// SequenceController orchestrates a slice of controllers sequentially.
type SequenceController struct {
	controllers  []*AnimationController
	currentIndex int
}

// NewSequence creates a new SequenceController.
func NewSequence(controllers ...*AnimationController) *SequenceController {
	return &SequenceController{
		controllers:  controllers,
		currentIndex: 0,
	}
}

// Advance progresses the sequence by dt.
// It sequentially plays controllers, passing residual dt to the next without losing time.
func (s *SequenceController) Advance(dt time.Duration) {
	for dt > 0 && s.currentIndex < len(s.controllers) {
		ctrl := s.controllers[s.currentIndex]

		if ctrl.Status() == StatusStopped || ctrl.Status() == StatusCompleted {
			ctrl.Play()
		}

		ctrl.mu.RLock()
		cDur := ctrl.duration
		cElap := ctrl.elapsed
		cDelay := ctrl.delayRemaining
		ctrl.mu.RUnlock()

		timeNeeded := cDur - cElap + cDelay

		if dt >= timeNeeded {
			ctrl.Advance(timeNeeded)
			dt -= timeNeeded
			s.currentIndex++
		} else {
			ctrl.Advance(dt)
			dt = 0
		}
	}
}

// Progress returns the overall progress of the sequence [0.0, 1.0].
func (s *SequenceController) Progress() float64 {
	if len(s.controllers) == 0 {
		return 1.0
	}
	total := float64(len(s.controllers))
	curr := float64(s.currentIndex)

	if s.IsCompleted() {
		return 1.0
	}

	ctrl := s.controllers[s.currentIndex]
	return (curr + ctrl.Progress()) / total
}

// IsCompleted returns true if all controllers in the sequence have completed.
func (s *SequenceController) IsCompleted() bool {
	return s.currentIndex >= len(s.controllers)
}

// ParallelController orchestrates a slice of controllers concurrently.
type ParallelController struct {
	controllers []*AnimationController
}

// NewParallel creates a new ParallelController.
func NewParallel(controllers ...*AnimationController) *ParallelController {
	return &ParallelController{
		controllers: controllers,
	}
}

// Advance progresses all non-completed controllers by dt.
func (p *ParallelController) Advance(dt time.Duration) {
	for _, ctrl := range p.controllers {
		if ctrl.Status() == StatusStopped {
			ctrl.Play()
		}
		if !ctrl.IsCompleted() {
			ctrl.Advance(dt)
		}
	}
}

// Progress returns the average progress of all controllers.
func (p *ParallelController) Progress() float64 {
	if len(p.controllers) == 0 {
		return 1.0
	}
	var sum float64
	for _, ctrl := range p.controllers {
		sum += ctrl.Progress()
	}
	return sum / float64(len(p.controllers))
}

// IsCompleted returns true if all controllers have completed.
func (p *ParallelController) IsCompleted() bool {
	if len(p.controllers) == 0 {
		return true
	}
	for _, ctrl := range p.controllers {
		if !ctrl.IsCompleted() {
			return false
		}
	}
	return true
}
