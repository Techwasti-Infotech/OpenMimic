package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"openmimic/engine"
	"openmimic/engine/internal/metrics"
	pb "openmimic/engine/pkg/proto/v1"
)

type runningGen struct {
	cancel    context.CancelFunc
	config    *pb.GeneratorConfig
	status    string
	collector *metrics.Collector
}

// GeneratorServer implements pb.GeneratorServiceServer.
type GeneratorServer struct {
	pb.UnimplementedGeneratorServiceServer

	mu         sync.RWMutex
	generators map[string]*runningGen
}

func NewGeneratorServer() *GeneratorServer {
	return &GeneratorServer{
		generators: make(map[string]*runningGen),
	}
}

func (s *GeneratorServer) StartGenerator(ctx context.Context, req *pb.StartGeneratorRequest) (*pb.StartGeneratorResponse, error) {
	if req.GetConfig() == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	cfg := req.GetConfig()
	genID := cfg.GetId()
	if genID == "" {
		genID = fmt.Sprintf("gen_%d", time.Now().UnixNano())
		cfg.Id = genID
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if prev, exists := s.generators[genID]; exists && prev.status == "RUNNING" {
		prev.cancel()
	}

	genCtx, cancel := context.WithCancel(context.Background())
	collector := metrics.NewCollector()
	s.generators[genID] = &runningGen{
		cancel:    cancel,
		config:    cfg,
		status:    "RUNNING",
		collector: collector,
	}

	// Start engine in background goroutine
	go func() {
		eng := engine.NewEngine(cfg, nil)
		eng.SetMetricsCollector(collector)
		_ = eng.Run(genCtx)

		s.mu.Lock()
		if cur, ok := s.generators[genID]; ok {
			cur.status = "STOPPED"
		}
		s.mu.Unlock()
	}()

	return &pb.StartGeneratorResponse{
		GeneratorId: genID,
		Status:      "RUNNING",
	}, nil
}

func (s *GeneratorServer) StopGenerator(ctx context.Context, req *pb.StopGeneratorRequest) (*pb.StopGeneratorResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	genID := req.GetGeneratorId()
	gen, exists := s.generators[genID]
	if !exists {
		return &pb.StopGeneratorResponse{
			GeneratorId: genID,
			Status:      "NOT_FOUND",
		}, nil
	}

	gen.cancel()
	gen.status = "STOPPED"

	return &pb.StopGeneratorResponse{
		GeneratorId: genID,
		Status:      "STOPPED",
	}, nil
}

func (s *GeneratorServer) StreamPreview(req *pb.StreamPreviewRequest, stream pb.GeneratorService_StreamPreviewServer) error {
	cfg := req.GetConfig()
	if cfg == nil {
		return fmt.Errorf("config is required for preview")
	}

	previewCount := req.GetPreviewCount()
	if previewCount <= 0 {
		previewCount = 10
	}

	// Create engine for isolated in-memory preview
	previewEngine := engine.NewEngine(cfg, nil)

	for i := int64(0); i < int64(previewCount); i++ {
		record, err := previewEngine.GenerateEvent(i)
		if err != nil {
			return err
		}

		payloadBytes, err := json.Marshal(record)
		if err != nil {
			return err
		}

		resp := &pb.StreamPreviewResponse{
			Index:       i,
			JsonPayload: string(payloadBytes),
			TimestampMs: time.Now().UnixMilli(),
		}

		if err := stream.Send(resp); err != nil {
			return err
		}
	}

	return nil
}

func (s *GeneratorServer) StreamMetrics(req *pb.StreamMetricsRequest, stream pb.GeneratorService_StreamMetricsServer) error {
	genID := req.GetGeneratorId()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case <-ticker.C:
			s.mu.RLock()
			gen, exists := s.generators[genID]
			status := "UNKNOWN"
			var snap metrics.Snapshot
			if exists {
				status = gen.status
				if gen.collector != nil {
					snap = gen.collector.Snapshot()
				}
			}
			s.mu.RUnlock()

			metricsResp := &pb.StreamMetricsResponse{
				GeneratorId:            genID,
				Status:                 status,
				TotalEvents:            snap.TotalEvents,
				ThroughputEventsPerSec: snap.ThroughputEPS,
				TotalErrors:            snap.TotalErrors,
				LatencyP99Ms:           snap.LatencyP99Ms,
				TimestampMs:            snap.TimestampMillis,
			}

			if err := stream.Send(metricsResp); err != nil {
				return err
			}
		}
	}
}

