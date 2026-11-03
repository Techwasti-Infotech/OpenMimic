package cluster

import (
	"context"
	"fmt"
	"sync"

	"openmimic/engine"
	pb "openmimic/engine/pkg/proto/v1"
)

// WorkerAgent executes partitioned sub-ranges of generator streams.
type WorkerAgent struct {
	id          string
	hostname    string
	cpuCores    int32
	coordinator *Coordinator
	mu          sync.Mutex
	running     bool
}

func NewWorkerAgent(id, hostname string, cpuCores int32, coord *Coordinator) *WorkerAgent {
	return &WorkerAgent{
		id:          id,
		hostname:    hostname,
		cpuCores:    cpuCores,
		coordinator: coord,
	}
}

// JoinCluster registers the worker with the cluster coordinator.
func (w *WorkerAgent) JoinCluster() error {
	req := &pb.RegisterWorkerRequest{
		WorkerId: w.id,
		Hostname: w.hostname,
		CpuCores: w.cpuCores,
	}
	_, err := w.coordinator.Register(req)
	return err
}

// ExecutePartition runs a specific event partition slice [startIndex, endIndex).
func (w *WorkerAgent) ExecutePartition(ctx context.Context, cfg *pb.GeneratorConfig, startIndex, endIndex int64) error {
	w.mu.Lock()
	w.running = true
	w.mu.Unlock()

	defer func() {
		w.mu.Lock()
		w.running = false
		w.mu.Unlock()
	}()

	// Clone config with scoped partition limit
	partitionCfg := &pb.GeneratorConfig{
		Id:                 fmt.Sprintf("%s_part_%d", cfg.GetId(), startIndex),
		Name:               cfg.GetName(),
		Seed:               cfg.GetSeed() + startIndex,
		EventLimit:         endIndex - startIndex,
		RateLimitPerSecond: cfg.GetRateLimitPerSecond(),
		Sink:               cfg.GetSink(),
		Fields:             cfg.GetFields(),
		Constraints:        cfg.GetConstraints(),
		TrafficProfile:     cfg.GetTrafficProfile(),
	}

	eng := engine.NewEngine(partitionCfg, nil)
	return eng.Run(ctx)
}
