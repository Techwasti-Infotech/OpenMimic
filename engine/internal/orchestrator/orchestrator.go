package orchestrator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"openmimic/engine"
	pb "openmimic/engine/pkg/proto/v1"
)

// TopologyRunner coordinates multi-stream dependent generator execution (Point 4).
type TopologyRunner struct {
	topology *pb.MultiGeneratorTopology
}

func NewTopologyRunner(topology *pb.MultiGeneratorTopology) *TopologyRunner {
	return &TopologyRunner{topology: topology}
}

// Execute runs the generators honoring DAG dependencies.
func (r *TopologyRunner) Execute(ctx context.Context) error {
	completed := make(map[string]bool)
	var mu sync.Mutex

	nodes := r.topology.GetNodes()
	remaining := len(nodes)

	for remaining > 0 {
		progressMade := false

		for _, node := range nodes {
			mu.Lock()
			if completed[node.GetGeneratorId()] {
				mu.Unlock()
				continue
			}

			// Check all dependencies
			depsSatisfied := true
			for _, dep := range node.GetDependsOn() {
				if !completed[dep] {
					depsSatisfied = false
					break
				}
			}
			mu.Unlock()

			if depsSatisfied {
				if node.GetDelayAfterParentMs() > 0 {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case <-time.After(time.Duration(node.GetDelayAfterParentMs()) * time.Millisecond):
					}
				}

				// Execute generator
				eng := engine.NewEngine(node.GetConfig(), nil)
				if err := eng.Run(ctx); err != nil {
					return fmt.Errorf("node %s execution failed: %w", node.GetGeneratorId(), err)
				}

				mu.Lock()
				completed[node.GetGeneratorId()] = true
				mu.Unlock()

				remaining--
				progressMade = true
			}
		}

		if !progressMade && remaining > 0 {
			return fmt.Errorf("deadlock or cyclic dependency detected in multi-generator topology")
		}
	}

	return nil
}

