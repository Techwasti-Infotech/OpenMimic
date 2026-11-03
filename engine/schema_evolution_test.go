package engine

import (
	"testing"

	pb "openmimic/engine/pkg/proto/v1"
)

func TestEngineSchemaEvolutionMidStream(t *testing.T) {
	cfg := &pb.GeneratorConfig{
		Id:                 "evo-gen",
		Seed:               123,
		EventLimit:         10,
		RateLimitPerSecond: 100,
		Fields: []*pb.Field{
			{
				Name: "event_id",
				Generator: &pb.FunctionGenerator{
					Func: &pb.FunctionGenerator_SequentialInteger{
						SequentialInteger: &pb.SequentialIntegerFunction{Start: 1, Step: 1},
					},
				},
			},
			{
				Name: "deprecated_v1_field",
				Generator: &pb.FunctionGenerator{
					Func: &pb.FunctionGenerator_Constant{
						Constant: &pb.ConstantFunction{
							Value: &pb.ConstantFunction_StringVal{StringVal: "old_payload"},
						},
					},
				},
			},
		},
		SchemaEvolution: &pb.SchemaEvolutionConfig{
			Enabled:        true,
			EvolveAtEvent:  3, // At event index >= 3, add new field and remove deprecated field
			RemovedFields:  []string{"deprecated_v1_field"},
			AddedFields: []*pb.Field{
				{
					Name: "new_v2_schema_field",
					Generator: &pb.FunctionGenerator{
						Func: &pb.FunctionGenerator_Constant{
							Constant: &pb.ConstantFunction{
								Value: &pb.ConstantFunction_StringVal{StringVal: "upgraded_v2"},
							},
						},
					},
				},
			},
		},
	}

	eng := NewEngine(cfg, nil)

	// Event index 0: Before evolution threshold
	rec0, err := eng.GenerateEvent(0)
	if err != nil {
		t.Fatalf("failed to generate rec0: %v", err)
	}
	if rec0["deprecated_v1_field"] != "old_payload" {
		t.Errorf("expected deprecated_v1_field before evolution")
	}
	if _, exists := rec0["new_v2_schema_field"]; exists {
		t.Errorf("new_v2_schema_field should not exist before index 3")
	}

	// Event index 3: Evolution kicks in
	rec3, err := eng.GenerateEvent(3)
	if err != nil {
		t.Fatalf("failed to generate rec3: %v", err)
	}
	if _, exists := rec3["deprecated_v1_field"]; exists {
		t.Errorf("deprecated_v1_field should have been removed after evolution at index 3")
	}
	if rec3["new_v2_schema_field"] != "upgraded_v2" {
		t.Errorf("expected new_v2_schema_field to be present after evolution")
	}
}
