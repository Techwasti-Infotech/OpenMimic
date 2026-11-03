package engine

import (
	"context"
	"testing"

	"openmimic/engine/internal/state"
	pb "openmimic/engine/pkg/proto/v1"
)

func TestDynamicCrossTableReferentialIntegrity(t *testing.T) {
	// Clear registry pool
	state.GetDefaultRegistry().Clear("active_customers")

	// 1. Parent generator: Creates customers into pool "active_customers"
	parentCfg := &pb.GeneratorConfig{
		Id:         "users-stream",
		EventLimit: 10,
		Fields: []*pb.Field{
			{
				Name: "customer_id",
				Generator: &pb.FunctionGenerator{
					Func: &pb.FunctionGenerator_EntityPool{
						EntityPool: &pb.EntityPoolFunction{
							PoolName:     "active_customers",
							PoolSize:     50,
							EntityPrefix: "cust_",
						},
					},
				},
			},
		},
	}

	parentEngine := NewEngine(parentCfg, nil)
	if err := parentEngine.Run(context.Background()); err != nil {
		t.Fatalf("parent engine failed: %v", err)
	}

	registeredCount := state.GetDefaultRegistry().PoolSize("active_customers")
	if registeredCount != 10 {
		t.Fatalf("expected 10 registered parent entities, got %d", registeredCount)
	}

	registeredEntities := state.GetDefaultRegistry().GetEntities("active_customers")
	validParents := make(map[string]bool)
	for _, id := range registeredEntities {
		validParents[id] = true
	}

	// 2. Child generator: Generates orders referencing foreign key "active_customers"
	childCfg := &pb.GeneratorConfig{
		Id:         "orders-stream",
		EventLimit: 25,
		Fields: []*pb.Field{
			{
				Name: "order_id",
				Generator: &pb.FunctionGenerator{
					Func: &pb.FunctionGenerator_Uuid{
						Uuid: &pb.UuidFunction{Uppercase: false},
					},
				},
			},
			{
				Name: "customer_ref",
				Generator: &pb.FunctionGenerator{
					Func: &pb.FunctionGenerator_EntityPool{
						EntityPool: &pb.EntityPoolFunction{
							PoolName: "active_customers",
							ForeignKey: &pb.RelationalForeignKey{
								ParentStreamId: "users-stream",
								ParentPoolName: "active_customers",
								AllowOrphans:   false,
							},
						},
					},
				},
			},
		},
	}

	childEngine := NewEngine(childCfg, nil)
	for i := int64(0); i < 25; i++ {
		record, err := childEngine.GenerateEvent(i)
		if err != nil {
			t.Fatalf("child event generation failed at %d: %v", i, err)
		}

		ref, ok := record["customer_ref"].(string)
		if !ok {
			t.Fatalf("customer_ref is not string: %v", record["customer_ref"])
		}

		if !validParents[ref] {
			t.Fatalf("referential integrity violation! Child referenced non-existent parent: %s", ref)
		}
	}
}
