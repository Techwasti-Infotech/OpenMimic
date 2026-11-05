package cep

import (
	"fmt"
	"math"
	"sync"

	pb "openmimic/engine/pkg/proto/v1"
)

// RollingAggregator maintains sliding and tumbling event windows across streams.
type RollingAggregator struct {
	mu         sync.Mutex
	streamID   string
	windowType string
	windowSize int
	op         string
	targetCol  string
	buffer     []float64
}

// GlobalAggregatorStore holds shared stateful aggregation windows across correlated streams.
var (
	globalStoreMu sync.Mutex
	globalStore   = make(map[string]*RollingAggregator)
)

// GetOrCreateAggregator retrieves or creates an aggregator for stream cep rollups.
func GetOrCreateAggregator(cfg *pb.StreamAggregationFunction) *RollingAggregator {
	globalStoreMu.Lock()
	defer globalStoreMu.Unlock()

	key := fmt.Sprintf("%s:%s:%s", cfg.GetSourceStreamId(), cfg.GetTargetField(), cfg.GetAggregationOp())
	agg, exists := globalStore[key]
	if !exists {
		winSize := int(cfg.GetWindowSizeEvents())
		if winSize <= 0 {
			winSize = 10
		}
		agg = &RollingAggregator{
			streamID:   cfg.GetSourceStreamId(),
			windowType: cfg.GetWindowType(),
			windowSize: winSize,
			op:         cfg.GetAggregationOp(),
			targetCol:  cfg.GetTargetField(),
			buffer:     make([]float64, 0, winSize),
		}
		globalStore[key] = agg
	}
	return agg
}

// Ingest records an incoming numeric observation into the window buffer.
func (a *RollingAggregator) Ingest(val float64) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.buffer = append(a.buffer, val)
	if len(a.buffer) > a.windowSize {
		// Sliding window: discard oldest
		a.buffer = a.buffer[1:]
	}
}

// ComputeRollup evaluates the configured aggregation over the active window.
func (a *RollingAggregator) ComputeRollup() float64 {
	a.mu.Lock()
	defer a.mu.Unlock()

	if len(a.buffer) == 0 {
		return 0.0
	}

	switch a.op {
	case "count":
		return float64(len(a.buffer))
	case "max":
		maxVal := -math.MaxFloat64
		for _, v := range a.buffer {
			if v > maxVal {
				maxVal = v
			}
		}
		return maxVal
	case "min":
		minVal := math.MaxFloat64
		for _, v := range a.buffer {
			if v < minVal {
				minVal = v
			}
		}
		return minVal
	case "avg":
		sum := 0.0
		for _, v := range a.buffer {
			sum += v
		}
		return sum / float64(len(a.buffer))
	case "sum":
		fallthrough
	default:
		sum := 0.0
		for _, v := range a.buffer {
			sum += v
		}
		return sum
	}
}

