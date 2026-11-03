package metrics

import (
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// Snapshot captures the performance telemetry at a given point in time.
type Snapshot struct {
	TotalEvents     int64
	TotalErrors     int64
	ThroughputEPS   float64
	LatencyP99Ms    float64
	LatencyMeanMs   float64
	TimestampMillis int64
}

// Collector records real-time event generation metrics and computes windowed percentiles.
type Collector struct {
	mu           sync.RWMutex
	totalEvents  int64
	totalErrors  int64
	lastEvents   int64
	lastTickTime time.Time
	latencies    []float64
}

// NewCollector creates an initialized Collector.
func NewCollector() *Collector {
	return &Collector{
		lastTickTime: time.Now(),
		latencies:    make([]float64, 0, 1000),
	}
}

// RecordEvent records a successful event and its production duration.
func (c *Collector) RecordEvent(duration time.Duration) {
	atomic.AddInt64(&c.totalEvents, 1)

	ms := float64(duration.Microseconds()) / 1000.0

	c.mu.Lock()
	if len(c.latencies) < 5000 {
		c.latencies = append(c.latencies, ms)
	}
	c.mu.Unlock()
}

// RecordError records a generation or sink error.
func (c *Collector) RecordError() {
	atomic.AddInt64(&c.totalErrors, 1)
}

// Snapshot computes rates and latency percentiles, resetting windowed buffers.
func (c *Collector) Snapshot() Snapshot {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(c.lastTickTime).Seconds()
	if elapsed <= 0 {
		elapsed = 1.0
	}

	currEvents := atomic.LoadInt64(&c.totalEvents)
	currErrors := atomic.LoadInt64(&c.totalErrors)
	windowEvents := currEvents - c.lastEvents
	throughput := float64(windowEvents) / elapsed

	c.lastEvents = currEvents
	c.lastTickTime = now

	// Calculate P99 and mean latency
	var p99, mean float64
	n := len(c.latencies)
	if n > 0 {
		sort.Float64s(c.latencies)
		p99Idx := int(float64(n) * 0.99)
		if p99Idx >= n {
			p99Idx = n - 1
		}
		p99 = c.latencies[p99Idx]

		var sum float64
		for _, v := range c.latencies {
			sum += v
		}
		mean = sum / float64(n)

		// Reset latency window
		c.latencies = c.latencies[:0]
	}

	return Snapshot{
		TotalEvents:     currEvents,
		TotalErrors:     currErrors,
		ThroughputEPS:   throughput,
		LatencyP99Ms:    p99,
		LatencyMeanMs:   mean,
		TimestampMillis: now.UnixMilli(),
	}
}

