package sink

import (
	"context"
	"fmt"
	"sync"

	pb "openmimic/engine/pkg/proto/v1"
)

// DataWarehouseBatchReceiver abstracts Snowflake/BigQuery micro-batch ingestion.
type DataWarehouseBatchReceiver interface {
	FlushBatch(ctx context.Context, target string, records []map[string]any) error
}

// MemoryWarehouseClient tracks micro-batches in memory for unit testing and local execution.
type MemoryWarehouseClient struct {
	mu      sync.Mutex
	Batches map[string][][]map[string]any
}

func NewMemoryWarehouseClient() *MemoryWarehouseClient {
	return &MemoryWarehouseClient{
		Batches: make(map[string][][]map[string]any),
	}
}

func (m *MemoryWarehouseClient) FlushBatch(ctx context.Context, target string, records []map[string]any) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	copied := make([]map[string]any, len(records))
	copy(copied, records)
	m.Batches[target] = append(m.Batches[target], copied)
	return nil
}

// SnowflakeSink buffers events and flushes them as micro-batches into Snowflake tables/Snowpipe.
type SnowflakeSink struct {
	mu        sync.Mutex
	receiver  DataWarehouseBatchReceiver
	target    string
	batchSize int
	buffer    []map[string]any
}

// NewSnowflakeSink creates a new Snowflake sink.
func NewSnowflakeSink(cfg *pb.SnowflakeSink, receiver DataWarehouseBatchReceiver) *SnowflakeSink {
	batchSize := int(cfg.GetBatchSize())
	if batchSize <= 0 {
		batchSize = 500
	}
	if receiver == nil {
		receiver = NewMemoryWarehouseClient()
	}
	target := fmt.Sprintf("%s.%s.%s", cfg.GetDatabase(), cfg.GetSchema(), cfg.GetTable())
	return &SnowflakeSink{
		receiver:  receiver,
		target:    target,
		batchSize: batchSize,
		buffer:    make([]map[string]any, 0, batchSize),
	}
}

func (s *SnowflakeSink) WriteEvent(ctx context.Context, key []byte, payload map[string]any) error {
	s.mu.Lock()
	s.buffer = append(s.buffer, payload)
	shouldFlush := len(s.buffer) >= s.batchSize
	s.mu.Unlock()

	if shouldFlush {
		return s.Flush(ctx)
	}
	return nil
}

func (s *SnowflakeSink) Flush(ctx context.Context) error {
	s.mu.Lock()
	if len(s.buffer) == 0 {
		s.mu.Unlock()
		return nil
	}
	records := s.buffer
	s.buffer = make([]map[string]any, 0, s.batchSize)
	s.mu.Unlock()

	return s.receiver.FlushBatch(ctx, s.target, records)
}

func (s *SnowflakeSink) Close() error {
	return s.Flush(context.Background())
}

// BigQuerySink buffers events and flushes them into Google Cloud BigQuery streaming ingestion.
type BigQuerySink struct {
	mu        sync.Mutex
	receiver  DataWarehouseBatchReceiver
	target    string
	batchSize int
	buffer    []map[string]any
}

// NewBigQuerySink creates a new BigQuery streaming sink.
func NewBigQuerySink(cfg *pb.BigQuerySink, receiver DataWarehouseBatchReceiver) *BigQuerySink {
	batchSize := int(cfg.GetBatchSize())
	if batchSize <= 0 {
		batchSize = 500
	}
	if receiver == nil {
		receiver = NewMemoryWarehouseClient()
	}
	target := fmt.Sprintf("%s:%s.%s", cfg.GetProjectId(), cfg.GetDatasetId(), cfg.GetTableId())
	return &BigQuerySink{
		receiver:  receiver,
		target:    target,
		batchSize: batchSize,
		buffer:    make([]map[string]any, 0, batchSize),
	}
}

func (b *BigQuerySink) WriteEvent(ctx context.Context, key []byte, payload map[string]any) error {
	b.mu.Lock()
	b.buffer = append(b.buffer, payload)
	shouldFlush := len(b.buffer) >= b.batchSize
	b.mu.Unlock()

	if shouldFlush {
		return b.Flush(ctx)
	}
	return nil
}

func (b *BigQuerySink) Flush(ctx context.Context) error {
	b.mu.Lock()
	if len(b.buffer) == 0 {
		b.mu.Unlock()
		return nil
	}
	records := b.buffer
	b.buffer = make([]map[string]any, 0, b.batchSize)
	b.mu.Unlock()

	return b.receiver.FlushBatch(ctx, b.target, records)
}

func (b *BigQuerySink) Close() error {
	return b.Flush(context.Background())
}
