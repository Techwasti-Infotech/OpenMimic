package chaos

import (
	"context"
	"fmt"
	"time"

	"openmimic/engine/internal/state"
	pb "openmimic/engine/pkg/proto/v1"
)

// ChaosSimulator introduces controlled latency delays, network partition errors, and backpressure.
type ChaosSimulator struct {
	enabled            bool
	failureRate        float64
	minLatency         time.Duration
	maxLatency         time.Duration
	simulateBackpress  bool
}

// NewChaosSimulator builds a chaos simulator from configuration.
func NewChaosSimulator(cfg *pb.ChaosConfig) *ChaosSimulator {
	if cfg == nil || !cfg.GetEnabled() {
		return &ChaosSimulator{enabled: false}
	}

	minLat := time.Duration(cfg.GetMinLatencyMs()) * time.Millisecond
	maxLat := time.Duration(cfg.GetMaxLatencyMs()) * time.Millisecond
	if maxLat < minLat {
		maxLat = minLat
	}

	return &ChaosSimulator{
		enabled:           true,
		failureRate:       cfg.GetFailureRate(),
		minLatency:        minLat,
		maxLatency:        maxLat,
		simulateBackpress: cfg.GetSimulateBackpressure(),
	}
}

// Intercept injects latency jitter and network partition errors before sink dispatch.
func (c *ChaosSimulator) Intercept(ctx context.Context, rng *state.DeterministicRand) error {
	if !c.enabled {
		return nil
	}

	// 1. Check network partition / failure injection
	if c.failureRate > 0 && rng.Float64() < c.failureRate {
		return fmt.Errorf("chaos error: simulated downstream network partition / socket disconnect")
	}

	// 2. Inject latency delay / jitter
	if c.maxLatency > 0 {
		diff := c.maxLatency - c.minLatency
		delay := c.minLatency
		if diff > 0 {
			jitter := time.Duration(rng.Intn(int(diff)))
			delay += jitter
		}

		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}
