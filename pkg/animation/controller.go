package animation

import (
	"sync"
	"time"
)

// AnimationMode defines how an animation loops or repeats.
type AnimationMode int

const (
	// LoopOnce plays the animation from 0.0 to 1.0 once and stops.
	LoopOnce AnimationMode = iota
	// LoopRepeat restarts the animation from 0.0 once it reaches 1.0.
	LoopRepeat
	// LoopPingPong reverses direction once reaching an end (0 -> 1 -> 0 -> ...).
	LoopPingPong
)

// AnimationStatus defines the current lifecycle state of an animation.
type AnimationStatus int

const (
	StatusStopped AnimationStatus = iota
	StatusRunning
	StatusPaused
	StatusCompleted
)

// AnimationController manages time progression, easing, and state for animations.
// It is thread-safe and can be queried or updated across concurrent ticks and renders.
type AnimationController struct {
	mu sync.RWMutex

	duration time.Duration
	delay    time.Duration
	easing   EasingFunc
	mode     AnimationMode

	elapsed        time.Duration
	delayRemaining time.Duration
	reverse        bool
	status         AnimationStatus
	loops          int
	maxLoops       int // 0 means infinite when LoopRepeat or LoopPingPong

	rawProgress float64
	value       float64

	onUpdate   func(val float64)
	onComplete func()
}

// NewController creates an AnimationController with a given duration and easing curve.
// If easing is nil, Linear easing is used.
func NewController(duration time.Duration, easing EasingFunc) *AnimationController {
	if duration <= 0 {
		duration = time.Millisecond * 300
	}
	if easing == nil {
		easing = Linear
	}
	return &AnimationController{
		duration: duration,
		easing:   easing,
		mode:     LoopOnce,
		status:   StatusStopped,
	}
}

// SetDuration updates the duration of one animation cycle.
func (c *AnimationController) SetDuration(d time.Duration) *AnimationController {
	c.mu.Lock()
	defer c.mu.Unlock()
	if d > 0 {
		c.duration = d
	}
	return c
}

// SetDelay sets an initial delay before the animation starts playing.
func (c *AnimationController) SetDelay(d time.Duration) *AnimationController {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.delay = d
	c.delayRemaining = d
	return c
}

// SetEasing updates the easing curve.
func (c *AnimationController) SetEasing(fn EasingFunc) *AnimationController {
	c.mu.Lock()
	defer c.mu.Unlock()
	if fn != nil {
		c.easing = fn
	}
	return c
}

// SetMode configures the loop mode.
func (c *AnimationController) SetMode(mode AnimationMode) *AnimationController {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.mode = mode
	return c
}

// SetMaxLoops sets the maximum repeat count. 0 means infinite repeating.
func (c *AnimationController) SetMaxLoops(max int) *AnimationController {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.maxLoops = max
	return c
}

// SetOnUpdate registers a listener called whenever value changes.
func (c *AnimationController) SetOnUpdate(fn func(val float64)) *AnimationController {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onUpdate = fn
	return c
}

// SetOnComplete registers a listener called when the animation finishes.
func (c *AnimationController) SetOnComplete(fn func()) *AnimationController {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onComplete = fn
	return c
}

// Play starts or resumes the animation.
func (c *AnimationController) Play() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.status == StatusCompleted {
		c.elapsed = 0
		c.delayRemaining = c.delay
		c.rawProgress = 0
		c.value = c.easing(0)
		c.reverse = false
		c.loops = 0
	}
	c.status = StatusRunning
}

// Pause suspends time advancement without resetting progress.
func (c *AnimationController) Pause() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.status == StatusRunning {
		c.status = StatusPaused
	}
}

// Stop halts the animation and resets its position to the beginning.
func (c *AnimationController) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.status = StatusStopped
	c.elapsed = 0
	c.delayRemaining = c.delay
	c.rawProgress = 0
	c.value = c.easing(0)
	c.reverse = false
	c.loops = 0
}

// Reset resets the animation to time zero without changing running status.
func (c *AnimationController) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.elapsed = 0
	c.delayRemaining = c.delay
	c.rawProgress = 0
	c.value = c.easing(0)
	c.reverse = false
	c.loops = 0
}

// Reverse toggles current playback direction.
func (c *AnimationController) Reverse() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.reverse = !c.reverse
	dur := c.duration
	if dur <= 0 {
		dur = time.Millisecond * 300
	}
	if c.elapsed >= dur {
		c.elapsed = 0
	} else if c.elapsed > 0 {
		c.elapsed = dur - c.elapsed
	}
	if c.status == StatusCompleted {
		c.status = StatusRunning
	}
}

// Advance advances the controller timeline by dt and returns the eased Value().
func (c *AnimationController) Advance(dt time.Duration) float64 {
	c.mu.Lock()
	if c.status != StatusRunning || dt <= 0 {
		val := c.value
		c.mu.Unlock()
		return val
	}

	// Handle initial delay
	if c.delayRemaining > 0 {
		if dt < c.delayRemaining {
			c.delayRemaining -= dt
			val := c.value
			c.mu.Unlock()
			return val
		}
		dt -= c.delayRemaining
		c.delayRemaining = 0
	}

	dur := c.duration
	if dur <= 0 {
		dur = time.Millisecond * 300
	}

	c.elapsed += dt

	completedNow := false
	var updateCb func(val float64)
	var completeCb func()

	for c.elapsed >= dur {
		c.loops++
		hasMax := c.maxLoops > 0 && c.loops >= c.maxLoops

		switch c.mode {
		case LoopOnce:
			c.elapsed = dur
			c.status = StatusCompleted
			completedNow = true
			goto calculation

		case LoopRepeat:
			if hasMax {
				c.elapsed = dur
				c.status = StatusCompleted
				completedNow = true
				goto calculation
			}
			c.elapsed -= dur

		case LoopPingPong:
			c.reverse = !c.reverse
			if hasMax {
				c.elapsed = dur
				c.status = StatusCompleted
				completedNow = true
				goto calculation
			}
			c.elapsed -= dur
		}
	}

calculation:
	progress := float64(c.elapsed) / float64(dur)
	if progress > 1.0 {
		progress = 1.0
	}
	if progress < 0.0 {
		progress = 0.0
	}

	if c.reverse {
		progress = 1.0 - progress
	}

	c.rawProgress = progress
	c.value = c.easing(progress)

	val := c.value
	updateCb = c.onUpdate
	if completedNow {
		completeCb = c.onComplete
	}

	c.mu.Unlock()

	if updateCb != nil {
		updateCb(val)
	}
	if completeCb != nil {
		completeCb()
	}

	return val
}

// Value returns the current eased value (typically in [0.0, 1.0], or exceeding for bounce/elastic).
func (c *AnimationController) Value() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.value
}

// Progress returns the linear, non-eased progression in [0.0, 1.0].
func (c *AnimationController) Progress() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.rawProgress
}

// Status returns the current lifecycle status.
func (c *AnimationController) Status() AnimationStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.status
}

// IsRunning reports whether the animation is currently progressing.
func (c *AnimationController) IsRunning() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.status == StatusRunning
}

// IsCompleted reports whether the animation has finished.
func (c *AnimationController) IsCompleted() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.status == StatusCompleted
}
