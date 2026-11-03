package state

import (
	"math/rand"
	"sync"
	"time"
)

// Clock provides a deterministic logical time sequence or wall clock for generators.
type Clock struct {
	mu           sync.Mutex
	baseTime     time.Time
	stepDuration time.Duration
	ticks        int64
}

// NewClock initializes a Clock with a fixed base time and step duration per event.
// If baseTime is zero, the current time is used.
func NewClock(baseTime time.Time, stepDuration time.Duration) *Clock {
	if baseTime.IsZero() {
		baseTime = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	if stepDuration == 0 {
		stepDuration = 100 * time.Millisecond
	}
	return &Clock{
		baseTime:     baseTime,
		stepDuration: stepDuration,
		ticks:        0,
	}
}

// Now returns the deterministic timestamp for the current tick without advancing.
func (c *Clock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.baseTime.Add(time.Duration(c.ticks) * c.stepDuration)
}

// Tick advances the clock by one step and returns the advanced time.
func (c *Clock) Tick() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ticks++
	return c.baseTime.Add(time.Duration(c.ticks) * c.stepDuration)
}

// Reset resets the clock ticks back to 0.
func (c *Clock) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.ticks = 0
}

// DeterministicRand wraps rand.Rand with a thread-safe mutex and reproducible sequence.
type DeterministicRand struct {
	mu  sync.Mutex
	rng *rand.Rand
}

// NewDeterministicRand creates a new DeterministicRand seeded with the given value.
func NewDeterministicRand(seed int64) *DeterministicRand {
	return &DeterministicRand{
		rng: rand.New(rand.NewSource(seed)),
	}
}

func (d *DeterministicRand) Float64() float64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.rng.Float64()
}

func (d *DeterministicRand) NormFloat64() float64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.rng.NormFloat64()
}

func (d *DeterministicRand) Uint32() uint32 {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.rng.Uint32()
}

func (d *DeterministicRand) Uint64() uint64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.rng.Uint64()
}

func (d *DeterministicRand) Intn(n int) int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.rng.Intn(n)
}

// GeneratorState tracks runtime state across generation iterations.
type GeneratorState struct {
	Clock *Clock
	RNG   *DeterministicRand
	vars  map[string]any
	mu    sync.RWMutex
}

// NewGeneratorState initializes a complete deterministic state machine.
func NewGeneratorState(seed int64, baseTime time.Time, stepDuration time.Duration) *GeneratorState {
	return &GeneratorState{
		Clock: NewClock(baseTime, stepDuration),
		RNG:   NewDeterministicRand(seed),
		vars:  make(map[string]any),
	}
}

// SetVar sets a persistent state variable across events.
func (s *GeneratorState) SetVar(key string, val any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.vars[key] = val
}

// GetVar gets a persistent state variable.
func (s *GeneratorState) GetVar(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.vars[key]
	return val, ok
}

