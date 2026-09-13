package animation

import (
	"math/rand"

	"github.com/baibeicha/goatui/pkg/core/buffer"
	"github.com/baibeicha/goatui/pkg/core/cell"
)

// Particle represents a single moving visual element.
type Particle struct {
	X, Y          float64
	Vx, Vy        float64
	Life, MaxLife float64
	Rune          rune
	Color         cell.Color
}

// ParticleSystem manages a collection of particles.
type ParticleSystem struct {
	particles []Particle
	gravity   float64
	bounds    buffer.Rect
}

// NewParticleSystem creates a new ParticleSystem with given bounds and gravity.
func NewParticleSystem(bounds buffer.Rect, gravity float64) *ParticleSystem {
	return &ParticleSystem{
		particles: make([]Particle, 0),
		gravity:   gravity,
		bounds:    bounds,
	}
}

// SetBounds updates the bounding box for particle rendering and bounds checks.
func (ps *ParticleSystem) SetBounds(bounds buffer.Rect) {
	ps.bounds = bounds
}

// Emit spawns new particles.
func (ps *ParticleSystem) Emit(count int, x, y float64, vxMin, vxMax, vyMin, vyMax, lifeMin, lifeMax float64, runes []rune, colors []cell.Color) {
	if len(runes) == 0 || len(colors) == 0 {
		return
	}
	for i := 0; i < count; i++ {
		p := Particle{
			X:       x,
			Y:       y,
			Vx:      vxMin + rand.Float64()*(vxMax-vxMin),
			Vy:      vyMin + rand.Float64()*(vyMax-vyMin),
			Life:    lifeMin + rand.Float64()*(lifeMax-lifeMin),
			Rune:    runes[rand.Intn(len(runes))],
			Color:   colors[rand.Intn(len(colors))],
		}
		p.MaxLife = p.Life
		ps.particles = append(ps.particles, p)
	}
}

// Update advances the physics simulation.
func (ps *ParticleSystem) Update(dt float64) {
	alive := ps.particles[:0]
	for _, p := range ps.particles {
		p.Life -= dt
		if p.Life > 0 {
			p.Vy += ps.gravity * dt
			p.X += p.Vx * dt
			p.Y += p.Vy * dt
			alive = append(alive, p)
		}
	}
	ps.particles = alive
}

// Draw renders all active particles into the buffer.
func (ps *ParticleSystem) Draw(buf *buffer.Buffer, area buffer.Rect) {
	for _, p := range ps.particles {
		if p.X < 0 || p.Y < 0 {
			continue
		}
		ix, iy := int(p.X), int(p.Y)
		if ix < ps.bounds.Width && iy < ps.bounds.Height {
			buf.SetRune(ps.bounds.X+ix, ps.bounds.Y+iy, p.Rune, p.Color, cell.DefaultColor(), cell.AttrNone)
		}
	}
}

// EmitConfetti is a preset to emit confetti-style particles.
func (ps *ParticleSystem) EmitConfetti(count int, x, y float64) {
	runes := []rune{'*', '~', 'o', '+', 'x'}
	colors := []cell.Color{
		cell.ColorHex("#FF5555"),
		cell.ColorHex("#55FF55"),
		cell.ColorHex("#5555FF"),
		cell.ColorHex("#FFFF55"),
		cell.ColorHex("#FF55FF"),
		cell.ColorHex("#55FFFF"),
	}
	ps.Emit(count, x, y, -20.0, 20.0, -10.0, 5.0, 1.0, 3.0, runes, colors)
}

// EmitSparks is a preset to emit spark-style particles.
func (ps *ParticleSystem) EmitSparks(count int, x, y float64) {
	runes := []rune{'.', ',', '*', '`', '\''}
	colors := []cell.Color{
		cell.ColorHex("#FFAA00"),
		cell.ColorHex("#FF5500"),
		cell.ColorHex("#FFFF00"),
		cell.ColorHex("#FFFFFF"),
	}
	ps.Emit(count, x, y, -15.0, 15.0, -15.0, -5.0, 0.5, 1.5, runes, colors)
}
