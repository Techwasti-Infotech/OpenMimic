package cluster

import (
	"context"
	"testing"

	pb "openmimic/engine/pkg/proto/v1"
)

func TestClusterCoordinatorPartitionPlanning(t *testing.T) {
	coord := NewCoordinator("test-cluster-1")

	// Register 4 worker nodes
	for i := 1; i <= 4; i++ {
		_, err := coord.Register(&pb.RegisterWorkerRequest{
			WorkerId: "worker-" + string(rune('0'+i)),
			Hostname: "node" + string(rune('0'+i)) + ".corp",
			CpuCores: 8,
		})
		if err != nil {
			t.Fatalf("failed to register worker %d: %v", i, err)
		}
	}

	if coord.WorkerCount() != 4 {
		t.Fatalf("expected 4 workers, got %d", coord.WorkerCount())
	}

	// Split 1,000,000 events across 4 workers
	plan, err := coord.PlanPartitions(1000000)
	if err != nil {
		t.Fatalf("failed to plan partitions: %v", err)
	}

	if len(plan) != 4 {
		t.Fatalf("expected 4 assignments, got %d", len(plan))
	}

	var totalCount int64 = 0
	for _, assignment := range plan {
		totalCount += assignment.Count
		if assignment.Count != 250000 {
			t.Fatalf("expected 250,000 events per worker, got %d", assignment.Count)
		}
	}

	if totalCount != 1000000 {
		t.Fatalf("expected total 1,000,000 events planned, got %d", totalCount)
	}
}

func TestWorkerAgentExecution(t *testing.T) {
	coord := NewCoordinator("agent-cluster")
	agent := NewWorkerAgent("worker-alpha", "localhost", 4, coord)

	if err := agent.JoinCluster(); err != nil {
		t.Fatalf("agent join failed: %v", err)
	}

	cfg := &pb.GeneratorConfig{
		Id:   "stream-orders",
		Seed: 100,
		Fields: []*pb.Field{
			{
				Name: "order_id",
				Generator: &pb.FunctionGenerator{
					Func: &pb.FunctionGenerator_Uuid{Uuid: &pb.UuidFunction{}},
				},
			},
		},
	}

	ctx := context.Background()
	// Execute partition: events 0 to 10
	err := agent.ExecutePartition(ctx, cfg, 0, 10)
	if err != nil {
		t.Fatalf("worker partition execution failed: %v", err)
	}
}
