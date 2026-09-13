package animation

import (
	"math"
)

// SpringConfig defines physical constants for harmonic oscillation.
type SpringConfig struct {
	Stiffness float64 // Tension k (default: 180.0)
	Damping   float64 // Friction c (default: 12.0)
	Mass      float64 // Mass m (default: 1.0)
	Precision float64 // Threshold to consider settled (default: 0.001)
}

// DefaultSpring creates a snappy, natural spring configuration.
func DefaultSpring() SpringConfig {
	return SpringConfig{
		Stiffness: 170.0,
		Damping:   26.0,
		Mass:      1.0,
		Precision: 0.001,
	}
}

// BouncySpring creates a spring with pronounced overshoot and bounce.
func BouncySpring() SpringConfig {
	return SpringConfig{
		Stiffness: 180.0,
		Damping:   12.0,
		Mass:      1.0,
		Precision: 0.001,
	}
}

// GentleSpring creates a smooth, dampened spring without overshoot.
func GentleSpring() SpringConfig {
	return SpringConfig{
		Stiffness: 120.0,
		Damping:   24.0,
		Mass:      1.0,
		Precision: 0.001,
	}
}

// SpringSimulation tracks a live physics simulation toward a target value.
type SpringSimulation struct {
	Config   SpringConfig
	Current  float64
	Target   float64
	Velocity float64
	Settled  bool
}

// NewSpringSimulation initializes a spring simulation.
func NewSpringSimulation(initial, target float64, config SpringConfig) *SpringSimulation {
	if config.Mass <= 0 {
		config.Mass = 1.0
	}
	if config.Precision <= 0 {
		config.Precision = 0.001
	}
	return &SpringSimulation{
		Config:   config,
		Current:  initial,
		Target:   target,
		Velocity: 0,
		Settled:  initial == target,
	}
}

// SetTarget updates the target value without resetting current position or momentum.
func (s *SpringSimulation) SetTarget(target float64) {
	s.Target = target
	s.Settled = false
}

// Update advances the simulation by dt seconds using sub-stepped Euler-Cromer integration.
func (s *SpringSimulation) Update(dt float64) (float64, bool) {
	if s.Settled {
		return s.Current, true
	}

	// Sub-step for numerical stability at large dt
	maxSubStep := 0.016 // ~60fps step
	steps := int(math.Ceil(dt / maxSubStep))
	if steps < 1 {
		steps = 1
	}
	subDt := dt / float64(steps)

	for i := 0; i < steps; i++ {
		// F = -k * (x - target) - c * v
		displacement := s.Current - s.Target
		springForce := -s.Config.Stiffness * displacement
		dampingForce := -s.Config.Damping * s.Velocity
		totalForce := springForce + dampingForce

		mass := s.Config.Mass
		if mass <= 0.0001 {
			mass = 1.0
		}
		accel := totalForce / mass
		s.Velocity += accel * subDt
		s.Current += s.Velocity * subDt
	}

	// Check if settled
	isNearTarget := math.Abs(s.Current-s.Target) < s.Config.Precision
	isNearRest := math.Abs(s.Velocity) < s.Config.Precision
	if isNearTarget && isNearRest {
		s.Current = s.Target
		s.Velocity = 0
		s.Settled = true
	}

	return s.Current, s.Settled
}
