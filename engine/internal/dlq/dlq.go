package dlq

import (
	"context"
	"fmt"
	"sync"
	"time"

	pb "openmimic/engine/pkg/proto/v1"
)

// QuarantinedRecord encapsulates a malformed or rejected event and its error stack.
type QuarantinedRecord struct {
	TimestampMs int64          `json:"timestamp_ms"`
	GeneratorID string         `json:"generator_id"`
	Reason      string         `json:"reason"`
	Attempt     int            `json:"attempt_count"`
	Payload     map[string]any `json:"payload"`
}

// Router routes events failing constraints or validation to a Dead Letter Queue destination.
type Router struct {
	cfg         *pb.DLQConfig
	mu          sync.RWMutex
	quarantined []*QuarantinedRecord
}

func NewRouter(cfg *pb.DLQConfig) *Router {
	return &Router{
		cfg:         cfg,
		quarantined: make([]*QuarantinedRecord, 0),
	}
}

// Quarantine stores or routes an invalid record.
func (r *Router) Quarantine(ctx context.Context, generatorID, reason string, attempt int, payload map[string]any) error {
	if r.cfg == nil || !r.cfg.GetEnabled() {
		return nil // DLQ disabled, record dropped
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	item := &QuarantinedRecord{
		TimestampMs: time.Now().UnixMilli(),
		GeneratorID: generatorID,
		Reason:      reason,
		Attempt:     attempt,
		Payload:     payload,
	}

	// In-memory quarantine buffer cap (up to 1,000 recent quarantined events)
	if len(r.quarantined) >= 1000 {
		if r.cfg.GetDropOnOverflow() {
			r.quarantined = r.quarantined[1:]
		} else {
			return fmt.Errorf("DLQ quarantine buffer is full")
		}
	}

	r.quarantined = append(r.quarantined, item)
	return nil
}

// GetQuarantined returns all records quarantined by the router.
func (r *Router) GetQuarantined() []*QuarantinedRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()

	res := make([]*QuarantinedRecord, len(r.quarantined))
	copy(res, r.quarantined)
	return res
}

// Count returns the number of quarantined events.
func (r *Router) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.quarantined)
}
