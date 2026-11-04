package sink

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	pb "openmimic/engine/pkg/proto/v1"
)

// GrpcStreamClient abstracts client streaming over gRPC for testing and network delivery.
type GrpcStreamClient interface {
	SendPayload(ctx context.Context, payload []byte) error
	Close() error
}

// MemoryGrpcStreamClient captures streamed messages in memory for unit testing.
type MemoryGrpcStreamClient struct {
	mu       sync.Mutex
	Messages [][]byte
	Closed   bool
}

func (m *MemoryGrpcStreamClient) SendPayload(ctx context.Context, payload []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, append([]byte(nil), payload...))
	return nil
}

func (m *MemoryGrpcStreamClient) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Closed = true
	return nil
}

// GrpcStreamingSink delivers events continuously over a gRPC stream.
type GrpcStreamingSink struct {
	client GrpcStreamClient
	target string
}

// NewGrpcStreamingSink creates a new gRPC streaming sink.
func NewGrpcStreamingSink(cfg *pb.GrpcStreamSink, client GrpcStreamClient) *GrpcStreamingSink {
	if client == nil {
		client = &MemoryGrpcStreamClient{}
	}
	return &GrpcStreamingSink{
		client: client,
		target: cfg.GetTarget(),
	}
}

func (s *GrpcStreamingSink) WriteEvent(ctx context.Context, key []byte, payload map[string]any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to encode grpc payload: %w", err)
	}
	return s.client.SendPayload(ctx, data)
}

func (s *GrpcStreamingSink) Close() error {
	return s.client.Close()
}
