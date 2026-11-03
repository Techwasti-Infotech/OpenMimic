package engine

import (
	"math"
	"testing"

	pb "openmimic/engine/pkg/proto/v1"
)

func TestEngineCalculatedExpressionField(t *testing.T) {
	cfg := &pb.GeneratorConfig{
		Id:                 "test-expr-engine",
		Seed:               42,
		EventLimit:         5,
		RateLimitPerSecond: 100,
		Fields: []*pb.Field{
			{
				Name: "quantity",
				Generator: &pb.FunctionGenerator{
					Func: &pb.FunctionGenerator_SequentialInteger{
						SequentialInteger: &pb.SequentialIntegerFunction{Start: 2, Step: 1},
					},
				},
			},
			{
				Name: "unit_price",
				Generator: &pb.FunctionGenerator{
					Func: &pb.FunctionGenerator_Constant{
						Constant: &pb.ConstantFunction{
							Value: &pb.ConstantFunction_DoubleVal{DoubleVal: 50.0},
						},
					},
				},
			},
			{
				Name: "total_amount",
				Generator: &pb.FunctionGenerator{
					Func: &pb.FunctionGenerator_Expression{
						Expression: &pb.CalculatedExpressionFunction{
							Expression: "quantity * unit_price",
							ReturnType: "double",
						},
					},
				},
			},
		},
	}

	eng := NewEngine(cfg, nil)

	event, err := eng.GenerateEvent(0)
	if err != nil {
		t.Fatalf("failed to generate event: %v", err)
	}

	total, ok := event["total_amount"].(float64)
	if !ok {
		t.Fatalf("expected float64 total_amount, got %T (%v)", event["total_amount"], event["total_amount"])
	}

	// First event: quantity=2, unit_price=50.0 -> total_amount=100.0
	if math.Abs(total-100.0) > 0.001 {
		t.Errorf("expected total_amount 100.0, got %f", total)
	}
}
