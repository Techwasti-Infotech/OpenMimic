package cep

import (
	"math"
	"testing"

	pb "openmimic/engine/pkg/proto/v1"
)

func TestRollingAggregatorSlidingWindow(t *testing.T) {
	cfg := &pb.StreamAggregationFunction{
		SourceStreamId:    "orders_stream",
		TargetField:       "amount",
		AggregationOp:     "avg",
		WindowType:        "sliding",
		WindowSizeEvents:  3, // Last 3 events
	}

	agg := GetOrCreateAggregator(cfg)
	agg.Ingest(10.0)
	agg.Ingest(20.0)
	agg.Ingest(30.0)

	// Window: [10, 20, 30] -> avg = 20.0
	res := agg.ComputeRollup()
	if math.Abs(res-20.0) > 0.001 {
		t.Errorf("expected avg 20.0, got %f", res)
	}

	// Ingest 40.0 -> Window slides to [20, 30, 40] -> avg = 30.0
	agg.Ingest(40.0)
	res2 := agg.ComputeRollup()
	if math.Abs(res2-30.0) > 0.001 {
		t.Errorf("expected sliding avg 30.0, got %f", res2)
	}
}

func TestRollingAggregatorSum(t *testing.T) {
	cfg := &pb.StreamAggregationFunction{
		SourceStreamId:   "clicks_stream",
		TargetField:      "count",
		AggregationOp:    "sum",
		WindowSizeEvents: 5,
	}

	agg := GetOrCreateAggregator(cfg)
	agg.Ingest(5.0)
	agg.Ingest(15.0)

	if agg.ComputeRollup() != 20.0 {
		t.Errorf("expected sum 20.0, got %f", agg.ComputeRollup())
	}
}
