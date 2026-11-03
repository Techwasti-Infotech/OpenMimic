package timeseries

import (
	"fmt"
	"math"
	"sync"
	"time"

	"openmimic/engine/internal/state"
	pb "openmimic/engine/pkg/proto/v1"
)

// DriftGenerator simulates continuous numerical time-series drift with sinusoidal seasonality and Gaussian noise.
type DriftGenerator struct {
	baseValue   float64
	trendSlope  float64
	amplitude   float64
	period      int64
	noiseStdDev float64
}

// NewDriftGenerator creates a generator for drifting time-series.
func NewDriftGenerator(cfg *pb.TimeSeriesDriftFunction) *DriftGenerator {
	period := cfg.GetSeasonalityPeriod()
	if period <= 0 {
		period = 100
	}
	return &DriftGenerator{
		baseValue:   cfg.GetBaseValue(),
		trendSlope:  cfg.GetTrendSlope(),
		amplitude:   cfg.GetSeasonalityAmplitude(),
		period:      period,
		noiseStdDev: cfg.GetNoiseStdDev(),
	}
}

// Generate calculates value at event index: Base + Trend*t + Amp*sin(2*pi*t/Period) + Noise
func (d *DriftGenerator) Generate(eventIndex int64, rng *state.DeterministicRand) float64 {
	t := float64(eventIndex)
	trend := d.trendSlope * t
	season := d.amplitude * math.Sin(2.0*math.Pi*t/float64(d.period))
	noise := 0.0
	if d.noiseStdDev > 0 {
		noise = rng.NormFloat64() * d.noiseStdDev
	}
	return d.baseValue + trend + season + noise
}

// SessionTracker tracks stateful user session lifetimes and inactivity timeouts.
type SessionTracker struct {
	mu             sync.Mutex
	prefix         string
	timeoutSec     int64
	activeSessions map[string]sessionState
	sessionCounter int64
}

type sessionState struct {
	sessionID string
	lastSeen  time.Time
}

// NewSessionTracker creates a stateful session window tracker.
func NewSessionTracker(cfg *pb.SessionWindowFunction) *SessionTracker {
	timeout := cfg.GetInactivityTimeoutSeconds()
	if timeout <= 0 {
		timeout = 1800 // default 30 minutes
	}
	prefix := cfg.GetSessionPrefix()
	if prefix == "" {
		prefix = "sess_"
	}
	return &SessionTracker{
		prefix:         prefix,
		timeoutSec:     timeout,
		activeSessions: make(map[string]sessionState),
		sessionCounter: 1000,
	}
}

// GetOrCreateSession returns the active session ID for user, or rotates if expired by inactivity.
func (s *SessionTracker) GetOrCreateSession(userID string, now time.Time) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, exists := s.activeSessions[userID]
	if exists && now.Sub(state.lastSeen).Seconds() < float64(s.timeoutSec) {
		// Existing valid session, refresh last seen
		state.lastSeen = now
		s.activeSessions[userID] = state
		return state.sessionID
	}

	// Create new session ID
	s.sessionCounter++
	newID := fmt.Sprintf("%s%08d", s.prefix, s.sessionCounter)
	s.activeSessions[userID] = sessionState{
		sessionID: newID,
		lastSeen:  now,
	}
	return newID
}
