package engine

import (
	"strings"
	"testing"

	pb "openmimic/engine/pkg/proto/v1"
)

func TestConstraintEnforcementInEngine(t *testing.T) {
	// Require fraud_flag == 1 AND amount > 50000
	cfg := &pb.GeneratorConfig{
		Id:         "fraud-stream",
		Seed:       42,
		EventLimit: 20,
		Fields: []*pb.Field{
			{
				Name: "fraud_flag",
				Generator: &pb.FunctionGenerator{
					Func: &pb.FunctionGenerator_WeightedChoice{
						WeightedChoice: &pb.WeightedChoiceFunction{
							Options: []*pb.WeightedChoiceOption{
								{Value: "0", Weight: 80.0},
								{Value: "1", Weight: 20.0},
							},
						},
					},
				},
			},
			{
				Name: "amount",
				Generator: &pb.FunctionGenerator{
					Func: &pb.FunctionGenerator_Distribution{
						Distribution: &pb.DistributionFunction{
							DistributionType: &pb.DistributionFunction_Normal{
								Normal: &pb.NormalDistribution{
									Mean:   40000.0,
									StdDev: 15000.0,
								},
							},
						},
					},
				},
			},
		},
		Constraints: []*pb.Constraint{
			{
				Id:                 "c_fraud",
				Name:               "Fraud predicate",
				Expression:         "fraud_flag == 1 && amount > 50000",
				MaxResampleRetries: 200,
			},
		},
	}

	eng := NewEngine(cfg, nil)

	for i := int64(0); i < 20; i++ {
		evt, err := eng.GenerateEvent(i)
		if err != nil {
			t.Fatalf("unexpected constraint generation error at event %d: %v", i, err)
		}

		if evt["fraud_flag"] != "1" {
			t.Fatalf("event %d violated constraint: fraud_flag expected '1', got '%v'", i, evt["fraud_flag"])
		}

		amt, ok := evt["amount"].(float64)
		if !ok || amt <= 50000.0 {
			t.Fatalf("event %d violated constraint: amount expected > 50000, got %v", i, evt["amount"])
		}
	}
}

func TestInfeasibleConstraintError(t *testing.T) {
	// Infeasible constraint: constant "A" == "B"
	cfg := &pb.GeneratorConfig{
		Id:   "infeasible-gen",
		Seed: 1,
		Fields: []*pb.Field{
			{
				Name: "status",
				Generator: &pb.FunctionGenerator{
					Func: &pb.FunctionGenerator_Constant{
						Constant: &pb.ConstantFunction{
							Value: &pb.ConstantFunction_StringVal{StringVal: "ACTIVE"},
						},
					},
				},
			},
		},
		Constraints: []*pb.Constraint{
			{
				Id:                 "c_impossible",
				Expression:         "status == DELETED",
				MaxResampleRetries: 5,
			},
		},
	}

	eng := NewEngine(cfg, nil)
	_, err := eng.GenerateEvent(0)
	if err == nil {
		t.Fatalf("expected error on infeasible constraint, got nil")
	}

	if !strings.Contains(err.Error(), "infeasible constraint set") {
		t.Fatalf("expected 'infeasible constraint set' error message, got: %v", err)
	}
}
