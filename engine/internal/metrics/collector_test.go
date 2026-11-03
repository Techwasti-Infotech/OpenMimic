package metrics

import (
	"testing"
	"time"
)

func TestCollectorRecordingAndSnapshot(t *testing.T) {
	c := NewCollector()

	// Record 100 events with 2ms latency
	for i := 0; i < 100; i++ {
		c.RecordEvent(2 * time.Millisecond)
	}
	c.RecordError()

	snap := c.Snapshot()
	if snap.TotalEvents != 100 {
		t.Fatalf("expected 100 events, got %d", snap.TotalEvents)
	}
	if snap.TotalErrors != 1 {
		t.Fatalf("expected 1 error, got %d", snap.TotalErrors)
	}
	if snap.LatencyP99Ms <= 0 {
		t.Fatalf("expected positive latency p99, got %f", snap.LatencyP99Ms)
	}
}

