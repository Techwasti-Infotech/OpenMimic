package cluster

import (
	"fmt"
	"sync"
	"time"

	pb "openmimic/engine/pkg/proto/v1"
)

// WorkerNode represents an active cluster worker.
type WorkerNode struct {
	ID             string
	Hostname       string
	CPUCores       int32
	LastHeartbeat  time.Time
	EventsProduced int64
	Throughput     float64
	Status         string // "IDLE", "RUNNING", "FAILED"
}

// Coordinator manages scale-out partitioned generation across distributed workers.
type Coordinator struct {
	mu        sync.RWMutex
	clusterID string
	workers   map[string]*WorkerNode
}

func NewCoordinator(clusterID string) *Coordinator {
	if clusterID == "" {
		clusterID = fmt.Sprintf("mimic-cluster-%d", time.Now().Unix())
	}
	return &Coordinator{
		clusterID: clusterID,
		workers:   make(map[string]*WorkerNode),
	}
}

// Register registers a new worker node into the cluster.
func (c *Coordinator) Register(req *pb.RegisterWorkerRequest) (*pb.RegisterWorkerResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	idx := int32(len(c.workers))
	c.workers[req.GetWorkerId()] = &WorkerNode{
		ID:            req.GetWorkerId(),
		Hostname:      req.GetHostname(),
		CPUCores:      req.GetCpuCores(),
		LastHeartbeat: time.Now(),
		Status:        "IDLE",
	}

	return &pb.RegisterWorkerResponse{
		Accepted:             true,
		ClusterId:            c.clusterID,
		AssignedWorkerIndex: idx,
	}, nil
}

// Heartbeat updates telemetry received from worker nodes.
func (c *Coordinator) Heartbeat(hb *pb.WorkerHeartbeat) (*pb.WorkerHeartbeatAck, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	node, exists := c.workers[hb.GetWorkerId()]
	if !exists {
		return &pb.WorkerHeartbeatAck{Acknowledged: false, ShouldTerminate: true}, nil
	}

	node.LastHeartbeat = time.Now()
	node.EventsProduced = hb.GetEventsProduced()
	node.Throughput = hb.GetCurrentThroughput()

	return &pb.WorkerHeartbeatAck{Acknowledged: true, ShouldTerminate: false}, nil
}

// PartitionPlan splits a large total event goal across available workers.
type PartitionAssignment struct {
	WorkerID   string
	StartIndex int64
	EndIndex   int64
	Count      int64
}

func (c *Coordinator) PlanPartitions(totalEvents int64) ([]PartitionAssignment, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	n := len(c.workers)
	if n == 0 {
		return nil, fmt.Errorf("no workers available in cluster")
	}

	workerIDs := make([]string, 0, n)
	for id := range c.workers {
		workerIDs = append(workerIDs, id)
	}

	chunk := totalEvents / int64(n)
	remainder := totalEvents % int64(n)

	assignments := make([]PartitionAssignment, 0, n)
	var currentStart int64 = 0

	for i, id := range workerIDs {
		count := chunk
		if int64(i) < remainder {
			count++
		}
		end := currentStart + count

		assignments = append(assignments, PartitionAssignment{
			WorkerID:   id,
			StartIndex: currentStart,
			EndIndex:   end,
			Count:      count,
		})
		currentStart = end
	}

	return assignments, nil
}

func (c *Coordinator) WorkerCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.workers)
}

// TotalClusterThroughput returns aggregated throughput across all active workers.
func (c *Coordinator) TotalClusterThroughput() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var total float64
	for _, w := range c.workers {
		total += w.Throughput
	}
	return total
}
