package sink

import (
	"context"
	"encoding/json"
	"fmt"

	pb "openmimic/engine/pkg/proto/v1"
)

// ClickHouseClient abstracts ClickHouse execution (allows in-memory mocking and driver integration).
type ClickHouseClient interface {
	Exec(ctx context.Context, query string, args ...any) error
	Close() error
}

// MockClickHouseClient stores executed SQL queries for verification.
type MockClickHouseClient struct {
	Queries []string
}

func NewMockClickHouseClient() *MockClickHouseClient {
	return &MockClickHouseClient{
		Queries: make([]string, 0),
	}
}

func (m *MockClickHouseClient) Exec(ctx context.Context, query string, args ...any) error {
	m.Queries = append(m.Queries, query)
	return nil
}

func (m *MockClickHouseClient) Close() error {
	return nil
}

// ClickHouseSink writes high-throughput events to ClickHouse tables.
type ClickHouseSink struct {
	cfg    *pb.ClickHouseSink
	client ClickHouseClient
}

func NewClickHouseSink(cfg *pb.ClickHouseSink, client ClickHouseClient) *ClickHouseSink {
	if client == nil {
		client = NewMockClickHouseClient()
	}
	return &ClickHouseSink{
		cfg:    cfg,
		client: client,
	}
}

func (s *ClickHouseSink) WriteEvent(ctx context.Context, key []byte, payload map[string]any) error {
	table := s.cfg.GetTable()
	if table == "" {
		table = "default.events"
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal clickhouse event: %w", err)
	}

	query := fmt.Sprintf("INSERT INTO %s FORMAT JSONEachRow %s", table, string(payloadJSON))
	return s.client.Exec(ctx, query)
}

func (s *ClickHouseSink) Close() error {
	return s.client.Close()
}

