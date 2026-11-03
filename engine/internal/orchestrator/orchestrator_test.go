package orchestrator

import (
	"context"
	"testing"

	pb "openmimic/engine/pkg/proto/v1"
)

func TestTopologyRunnerOrchestration(t *testing.T) {
	// Topology: User Stream runs first -> Order Stream runs second referencing user IDs
	userConfig := &pb.GeneratorConfig{
		Id:         "users-stream",
		EventLimit: 5,
		Fields: []*pb.Field{
			{Name: "user_id", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_Uuid{Uuid: &pb.UuidFunction{}}}},
		},
	}

	orderConfig := &pb.GeneratorConfig{
		Id:         "orders-stream",
		EventLimit: 5,
		Fields: []*pb.Field{
			{Name: "order_id", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_Uuid{Uuid: &pb.UuidFunction{}}}},
			{Name: "amount", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_Constant{Constant: &pb.ConstantFunction{Value: &pb.ConstantFunction_DoubleVal{DoubleVal: 99.5}}}}},
		},
	}

	topology := &pb.MultiGeneratorTopology{
		Id:   "ecommerce-flow",
		Name: "Users followed by Orders",
		Nodes: []*pb.TopologyNode{
			{
				GeneratorId: "users-stream",
				Config:      userConfig,
			},
			{
				GeneratorId:        "orders-stream",
				Config:             orderConfig,
				DependsOn:          []string{"users-stream"},
				DelayAfterParentMs: 10,
			},
		},
	}

	runner := NewTopologyRunner(topology)
	ctx := context.Background()

	if err := runner.Execute(ctx); err != nil {
		t.Fatalf("topology execution failed: %v", err)
	}
}

func TestTopologyCycleDetection(t *testing.T) {
	// Cyclic dependency: A depends on B, B depends on A
	topology := &pb.MultiGeneratorTopology{
		Id: "cyclic-flow",
		Nodes: []*pb.TopologyNode{
			{
				GeneratorId: "stream-a",
				DependsOn:   []string{"stream-b"},
				Config:      &pb.GeneratorConfig{Id: "stream-a", EventLimit: 1},
			},
			{
				GeneratorId: "stream-b",
				DependsOn:   []string{"stream-a"},
				Config:      &pb.GeneratorConfig{Id: "stream-b", EventLimit: 1},
			},
		},
	}

	runner := NewTopologyRunner(topology)
	ctx := context.Background()

	err := runner.Execute(ctx)
	if err == nil {
		t.Fatalf("expected cycle error, got nil")
	}
}

