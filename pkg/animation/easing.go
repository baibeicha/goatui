package animation

import (
	"math"
)

// EasingFunc defines a function mapping normalized progress t in [0.0, 1.0] to an interpolated value.
type EasingFunc func(t float64) float64

// Linear is constant velocity easing.
func Linear(t float64) float64 {
	return clamp01(t)
}

// -----------------------------------------------------------------------------
// Quadratic Curves (t^2)
// -----------------------------------------------------------------------------

func EaseInQuad(t float64) float64 {
	t = clamp01(t)
	return t * t
}

func EaseOutQuad(t float64) float64 {
	t = clamp01(t)
	return t * (2 - t)
}

func EaseInOutQuad(t float64) float64 {
	t = clamp01(t)
	if t < 0.5 {
		return 2 * t * t
	}
	return -1 + (4-2*t)*t
}

// -----------------------------------------------------------------------------
// Cubic Curves (t^3)
// -----------------------------------------------------------------------------

func EaseInCubic(t float64) float64 {
	t = clamp01(t)
	return t * t * t
}

func EaseOutCubic(t float64) float64 {
	t = clamp01(t)
	t--
	return t*t*t + 1
}

func EaseInOutCubic(t float64) float64 {
	t = clamp01(t)
	if t < 0.5 {
		return 4 * t * t * t
	}
	p := 2*t - 2
	return 0.5*p*p*p + 1
}

// -----------------------------------------------------------------------------
// Quartic Curves (t^4)
// -----------------------------------------------------------------------------

func EaseInQuart(t float64) float64 {
	t = clamp01(t)
	return t * t * t * t
}

func EaseOutQuart(t float64) float64 {
	t = clamp01(t)
	t--
	return 1 - t*t*t*t
}

func EaseInOutQuart(t float64) float64 {
	t = clamp01(t)
	if t < 0.5 {
		return 8 * t * t * t * t
	}
	t--
	return 1 - 8*t*t*t*t
}

// -----------------------------------------------------------------------------
// Sinusoidal Curves
// -----------------------------------------------------------------------------

func EaseInSine(t float64) float64 {
	t = clamp01(t)
	return 1 - math.Cos((t*math.Pi)/2)
}

func EaseOutSine(t float64) float64 {
	t = clamp01(t)
	return math.Sin((t * math.Pi) / 2)
}

func EaseInOutSine(t float64) float64 {
	t = clamp01(t)
	return -(math.Cos(math.Pi*t) - 1) / 2
}

// -----------------------------------------------------------------------------
// Exponential Curves
// -----------------------------------------------------------------------------

func EaseInExpo(t float64) float64 {
	t = clamp01(t)
	if t == 0 {
		return 0
	}
	return math.Pow(2, 10*(t-1))
}

func EaseOutExpo(t float64) float64 {
	t = clamp01(t)
	if t == 1 {
		return 1
	}
	return 1 - math.Pow(2, -10*t)
}

func EaseInOutExpo(t float64) float64 {
	t = clamp01(t)
	if t == 0 {
		return 0
	}
	if t == 1 {
		return 1
	}
	if t < 0.5 {
		return math.Pow(2, 20*t-10) / 2
	}
	return (2 - math.Pow(2, -20*t+10)) / 2
}

// -----------------------------------------------------------------------------
// Elastic Curves
// -----------------------------------------------------------------------------

func EaseInElastic(t float64) float64 {
	t = clamp01(t)
	if t == 0 {
		return 0
	}
	if t == 1 {
		return 1
	}
	return -math.Pow(2, 10*(t-1)) * math.Sin((t-1.1)*(2*math.Pi)/0.4)
}

func EaseOutElastic(t float64) float64 {
	t = clamp01(t)
	if t == 0 {
		return 0
	}
	if t == 1 {
		return 1
	}
	return math.Pow(2, -10*t)*math.Sin((t-0.1)*(2*math.Pi)/0.4) + 1
}

func EaseInOutElastic(t float64) float64 {
	t = clamp01(t)
	if t == 0 {
		return 0
	}
	if t == 1 {
		return 1
	}
	t *= 2
	if t < 1 {
		return -0.5 * math.Pow(2, 10*(t-1)) * math.Sin((t-1.1)*(2*math.Pi)/0.4)
	}
	return 0.5*math.Pow(2, -10*(t-1))*math.Sin((t-1.1)*(2*math.Pi)/0.4) + 1
}

// -----------------------------------------------------------------------------
// Bounce Curves
// -----------------------------------------------------------------------------

func EaseOutBounce(t float64) float64 {
	t = clamp01(t)
	const (
		n1 = 7.5625
		d1 = 2.75
	)
	if t < 1/d1 {
		return n1 * t * t
	} else if t < 2/d1 {
		t -= 1.5 / d1
		return n1*t*t + 0.75
	} else if t < 2.5/d1 {
		t -= 2.25 / d1
		return n1*t*t + 0.9375
	}
	t -= 2.625 / d1
	return n1*t*t + 0.984375
}

func EaseInBounce(t float64) float64 {
	t = clamp01(t)
	return 1 - EaseOutBounce(1-t)
}

func EaseInOutBounce(t float64) float64 {
	t = clamp01(t)
	if t < 0.5 {
		return (1 - EaseOutBounce(1-2*t)) / 2
	}
	return (1 + EaseOutBounce(2*t-1)) / 2
}

// -----------------------------------------------------------------------------
// Back Curves (Overshoot)
// -----------------------------------------------------------------------------

func EaseInBack(t float64) float64 {
	t = clamp01(t)
	const s = 1.70158
	return t * t * ((s+1)*t - s)
}

func EaseOutBack(t float64) float64 {
	t = clamp01(t)
	const s = 1.70158
	t--
	return t*t*((s+1)*t+s) + 1
}

func EaseInOutBack(t float64) float64 {
	t = clamp01(t)
	const s = 1.70158 * 1.525
	t *= 2
	if t < 1 {
		return 0.5 * (t * t * ((s+1)*t - s))
	}
	t -= 2
	return 0.5 * (t*t*((s+1)*t+s) + 2)
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
