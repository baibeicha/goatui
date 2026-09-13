package animation

import (
	"math"
	"sync"
	"testing"
	"time"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
)

func TestEasingEndpoints(t *testing.T) {
	curves := []struct {
		name string
		fn   EasingFunc
	}{
		{"Linear", Linear},
		{"EaseInQuad", EaseInQuad},
		{"EaseOutQuad", EaseOutQuad},
		{"EaseInOutQuad", EaseInOutQuad},
		{"EaseInCubic", EaseInCubic},
		{"EaseOutCubic", EaseOutCubic},
		{"EaseInOutCubic", EaseInOutCubic},
		{"EaseInQuart", EaseInQuart},
		{"EaseOutQuart", EaseOutQuart},
		{"EaseInOutQuart", EaseInOutQuart},
		{"EaseInSine", EaseInSine},
		{"EaseOutSine", EaseOutSine},
		{"EaseInOutSine", EaseInOutSine},
		{"EaseInExpo", EaseInExpo},
		{"EaseOutExpo", EaseOutExpo},
		{"EaseInOutExpo", EaseInOutExpo},
		{"EaseInElastic", EaseInElastic},
		{"EaseOutElastic", EaseOutElastic},
		{"EaseInOutElastic", EaseInOutElastic},
		{"EaseInBounce", EaseInBounce},
		{"EaseOutBounce", EaseOutBounce},
		{"EaseInOutBounce", EaseInOutBounce},
		{"EaseInBack", EaseInBack},
		{"EaseOutBack", EaseOutBack},
		{"EaseInOutBack", EaseInOutBack},
	}

	for _, tc := range curves {
		t.Run(tc.name, func(t *testing.T) {
			start := tc.fn(0.0)
			end := tc.fn(1.0)
			if math.Abs(start) > 1e-4 {
				t.Errorf("%s: expected f(0) ~ 0, got %f", tc.name, start)
			}
			if math.Abs(end-1.0) > 1e-4 {
				t.Errorf("%s: expected f(1) ~ 1, got %f", tc.name, end)
			}
		})
	}
}

func TestSpringSimulation(t *testing.T) {
	spring := NewSpringSimulation(0.0, 100.0, DefaultSpring())
	var pos float64
	var finished bool

	dt := 1.0 / 60.0 // 60 FPS
	for i := 0; i < 300; i++ {
		pos, finished = spring.Update(dt)
		if finished {
			break
		}
	}

	if !finished {
		t.Errorf("Spring did not settle within 5 seconds, final pos: %f", pos)
	}
	if math.Abs(pos-100.0) > 0.05 {
		t.Errorf("Spring settled at %f, expected ~100.0", pos)
	}
}

func TestLerpFunctions(t *testing.T) {
	// Float
	if v := LerpFloat(10.0, 20.0, 0.5); v != 15.0 {
		t.Errorf("LerpFloat expected 15, got %f", v)
	}
	// Int
	if v := LerpInt(0, 10, 0.4); v != 4 {
		t.Errorf("LerpInt expected 4, got %d", v)
	}
	// Rect
	r1 := buffer.Rect{X: 0, Y: 0, Width: 10, Height: 20}
	r2 := buffer.Rect{X: 100, Y: 50, Width: 30, Height: 40}
	rMid := LerpRect(r1, r2, 0.5)
	if rMid.X != 50 || rMid.Y != 25 || rMid.Width != 20 || rMid.Height != 30 {
		t.Errorf("LerpRect unexpected mid rect: %+v", rMid)
	}
	// Color
	c1 := cell.RGB(0, 0, 0)
	c2 := cell.RGB(100, 200, 50)
	cMid := LerpColor(c1, c2, 0.5)
	r := uint8((cMid.Value >> 16) & 0xFF)
	g := uint8((cMid.Value >> 8) & 0xFF)
	b := uint8(cMid.Value & 0xFF)
	if r != 50 || g != 100 || b != 25 {
		t.Errorf("LerpColor unexpected RGB: %d, %d, %d", r, g, b)
	}
}

func TestAnimationController(t *testing.T) {
	dur := 100 * time.Millisecond
	ctrl := NewController(dur, Linear)

	if ctrl.IsRunning() {
		t.Errorf("Controller should be stopped initially")
	}

	ctrl.Play()
	if !ctrl.IsRunning() {
		t.Errorf("Controller should be running after Play()")
	}

	// Advance 50ms (halfway)
	v := ctrl.Advance(50 * time.Millisecond)
	if math.Abs(v-0.5) > 0.01 {
		t.Errorf("Expected ~0.5 at 50ms, got %f", v)
	}

	// Advance another 60ms (over duration)
	completed := false
	ctrl.SetOnComplete(func() {
		completed = true
	})

	v = ctrl.Advance(60 * time.Millisecond)
	if v != 1.0 {
		t.Errorf("Expected 1.0 at completion, got %f", v)
	}
	if !completed {
		t.Errorf("Expected onComplete callback to fire")
	}
	if !ctrl.IsCompleted() {
		t.Errorf("Expected IsCompleted to be true")
	}
}

func TestAnimationPingPong(t *testing.T) {
	dur := 100 * time.Millisecond
	ctrl := NewController(dur, Linear)
	ctrl.SetMode(LoopPingPong)
	ctrl.Play()

	// 100ms: reaches 1.0
	ctrl.Advance(100 * time.Millisecond)
	// 50ms more: should be at 0.5 going reverse
	v := ctrl.Advance(50 * time.Millisecond)
	if math.Abs(v-0.5) > 0.02 {
		t.Errorf("Expected ~0.5 on reverse stroke of ping pong, got %f", v)
	}
}

func TestTransitions(t *testing.T) {
	// SlideRect
	target := buffer.Rect{X: 10, Y: 10, Width: 20, Height: 10}
	slid := SlideRect(target, 20, SlideFromLeft, 0.5)
	if slid.X != 0 {
		t.Errorf("Expected X=0 (10 - 10), got %d", slid.X)
	}

	// ShimmerOffset
	shimmer := ShimmerOffset(50, 0.5)
	if shimmer < 20 || shimmer > 30 {
		t.Errorf("Expected shimmer to be centered around 25, got %d", shimmer)
	}

	// SmoothFloat
	sf := NewSmoothFloat(0.0, 20.0)
	sf.SetTarget(100.0)
	for i := 0; i < 60; i++ {
		sf.Update(1.0 / 60.0)
	}
	if !sf.IsSettled() || math.Abs(sf.Value()-100.0) > 0.01 {
		t.Errorf("SmoothFloat failed to settle to 100, got %f", sf.Value())
	}
}

func TestConcurrentController(t *testing.T) {
	ctrl := NewController(50*time.Millisecond, EaseOutQuad)
	ctrl.SetMode(LoopRepeat)
	ctrl.Play()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = ctrl.Advance(time.Millisecond)
				_ = ctrl.Value()
				_ = ctrl.Progress()
			}
		}()
	}
	wg.Wait()
}
