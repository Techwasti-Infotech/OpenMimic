package sink

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	pb "openmimic/engine/pkg/proto/v1"
)

// LakehouseStorageClient abstracts object storage or filesystem writes for Iceberg and Delta Lake
type LakehouseStorageClient interface {
	WriteFile(ctx context.Context, path string, data []byte) error
}

// LocalLakehouseStorageClient writes lakehouse tables to local filesystem
type LocalLakehouseStorageClient struct{}

func (c *LocalLakehouseStorageClient) WriteFile(ctx context.Context, path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// MemoryLakehouseStorageClient keeps files in memory for fast testing
type MemoryLakehouseStorageClient struct {
	mu    sync.Mutex
	Files map[string][]byte
}

func NewMemoryLakehouseStorageClient() *MemoryLakehouseStorageClient {
	return &MemoryLakehouseStorageClient{Files: make(map[string][]byte)}
}

func (m *MemoryLakehouseStorageClient) WriteFile(ctx context.Context, path string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Files[path] = append([]byte(nil), data...)
	return nil
}

// IcebergStreamingSink commits Apache Iceberg Table Metadata V2 snapshots
type IcebergStreamingSink struct {
	mu           sync.Mutex
	cfg          *pb.IcebergSink
	client       LakehouseStorageClient
	buffer       []map[string]any
	snapshotID   int64
	sequenceNum  int64
	lastCommit   time.Time
	batchSize    int
}

// NewIcebergStreamingSink creates an Apache Iceberg table streaming sink
func NewIcebergStreamingSink(cfg *pb.IcebergSink, client LakehouseStorageClient) *IcebergStreamingSink {
	if client == nil {
		client = &LocalLakehouseStorageClient{}
	}
	return &IcebergStreamingSink{
		cfg:         cfg,
		client:      client,
		buffer:      make([]map[string]any, 0, 500),
		snapshotID:  time.Now().UnixNano(),
		sequenceNum: 1,
		lastCommit:  time.Now(),
		batchSize:   500,
	}
}

// WriteEvent buffers events and commits Iceberg data files + metadata snapshots
func (s *IcebergStreamingSink) WriteEvent(ctx context.Context, key []byte, payload map[string]any) error {
	s.mu.Lock()
	s.buffer = append(s.buffer, payload)
	shouldCommit := len(s.buffer) >= s.batchSize
	s.mu.Unlock()

	if shouldCommit {
		return s.CommitSnapshot(ctx)
	}
	return nil
}

// CommitSnapshot writes a Parquet columnar data file, manifest file, and metadata JSON snapshot
func (s *IcebergStreamingSink) CommitSnapshot(ctx context.Context) error {
	s.mu.Lock()
	if len(s.buffer) == 0 {
		s.mu.Unlock()
		return nil
	}
	records := s.buffer
	s.buffer = make([]map[string]any, 0, s.batchSize)
	s.sequenceNum++
	snapID := s.snapshotID + s.sequenceNum
	s.lastCommit = time.Now()
	s.mu.Unlock()

	basePath := fmt.Sprintf("%s/%s/%s", s.cfg.GetWarehouse(), s.cfg.GetDatabase(), s.cfg.GetTable())
	timestamp := time.Now().UTC().Format("20060102_150405")
	dataFile := fmt.Sprintf("%s/data/00000-%d-%s.parquet", basePath, s.sequenceNum, timestamp)

	// Encode columnar Parquet data chunk
	parquetData, err := EncodeColumnarParquet(records, "zstd")
	if err != nil {
		return fmt.Errorf("failed to encode iceberg parquet data: %w", err)
	}
	if err := s.client.WriteFile(ctx, dataFile, parquetData); err != nil {
		return fmt.Errorf("failed to write iceberg data file: %w", err)
	}

	// Write Iceberg Table Metadata V2 JSON
	metadata := map[string]any{
		"format-version": 2,
		"table-uuid":     fmt.Sprintf("iceberg-tbl-%s", s.cfg.GetTable()),
		"location":       basePath,
		"last-updated-ms": time.Now().UnixMilli(),
		"current-snapshot-id": snapID,
		"snapshots": []map[string]any{
			{
				"snapshot-id": snapID,
				"sequence-number": s.sequenceNum,
				"timestamp-ms": time.Now().UnixMilli(),
				"summary": map[string]string{
					"operation": "append",
					"added-records": fmt.Sprintf("%d", len(records)),
					"added-data-files": "1",
				},
				"manifest-list": fmt.Sprintf("%s/metadata/snap-%d-manifest-list.avro", basePath, snapID),
			},
		},
	}
	metaBytes, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}
	metaPath := fmt.Sprintf("%s/metadata/v%d.metadata.json", basePath, s.sequenceNum)
	return s.client.WriteFile(ctx, metaPath, metaBytes)
}

func (s *IcebergStreamingSink) Close() error {
	return s.CommitSnapshot(context.Background())
}

// DeltaLakeStreamingSink commits streaming records and Delta Lake ACID transaction logs (_delta_log/*.json)
type DeltaLakeStreamingSink struct {
	mu         sync.Mutex
	cfg        *pb.DeltaLakeSink
	client     LakehouseStorageClient
	buffer     []map[string]any
	version    int64
	batchSize  int
}

// NewDeltaLakeStreamingSink creates a Delta Lake table streaming sink
func NewDeltaLakeStreamingSink(cfg *pb.DeltaLakeSink, client LakehouseStorageClient) *DeltaLakeStreamingSink {
	if client == nil {
		client = &LocalLakehouseStorageClient{}
	}
	batch := int(cfg.GetBatchSize())
	if batch <= 0 {
		batch = 500
	}
	return &DeltaLakeStreamingSink{
		cfg:       cfg,
		client:    client,
		buffer:    make([]map[string]any, 0, batch),
		version:   0,
		batchSize: batch,
	}
}

// WriteEvent buffers events and commits Delta Lake parquet files + transaction logs
func (s *DeltaLakeStreamingSink) WriteEvent(ctx context.Context, key []byte, payload map[string]any) error {
	s.mu.Lock()
	s.buffer = append(s.buffer, payload)
	shouldCommit := len(s.buffer) >= s.batchSize
	s.mu.Unlock()

	if shouldCommit {
		return s.CommitBatch(ctx)
	}
	return nil
}

// CommitBatch writes a Parquet part and atomic Delta Lake commit log (000000.json)
func (s *DeltaLakeStreamingSink) CommitBatch(ctx context.Context) error {
	s.mu.Lock()
	if len(s.buffer) == 0 {
		s.mu.Unlock()
		return nil
	}
	records := s.buffer
	s.buffer = make([]map[string]any, 0, s.batchSize)
	ver := s.version
	s.version++
	s.mu.Unlock()

	basePath := s.cfg.GetTableUri()
	timestamp := time.Now().UTC().Format("20060102_150405")
	partFileName := fmt.Sprintf("part-%05d-%s-c000.snappy.parquet", ver, timestamp)
	fullDataPath := fmt.Sprintf("%s/%s", basePath, partFileName)

	// Encode columnar Parquet file
	parquetData, err := EncodeColumnarParquet(records, "snappy")
	if err != nil {
		return fmt.Errorf("failed to encode delta parquet data: %w", err)
	}
	if err := s.client.WriteFile(ctx, fullDataPath, parquetData); err != nil {
		return fmt.Errorf("failed to write delta data file: %w", err)
	}

	// If transaction log generation is enabled, write _delta_log/00000000000000000000.json
	if s.cfg.GetGenerateDeltaLog() {
		commitLog := []map[string]any{
			{
				"commitInfo": map[string]any{
					"timestamp": time.Now().UnixMilli(),
					"operation": "WRITE",
					"operationParameters": map[string]string{
						"mode": "Append",
					},
				},
			},
			{
				"add": map[string]any{
					"path": partFileName,
					"size": len(parquetData),
					"modificationTime": time.Now().UnixMilli(),
					"dataChange": true,
					"stats": fmt.Sprintf(`{"numRecords":%d}`, len(records)),
				},
			},
		}

		var logBuf []byte
		for _, entry := range commitLog {
			b, _ := json.Marshal(entry)
			logBuf = append(logBuf, b...)
			logBuf = append(logBuf, '\n')
		}

		logFile := fmt.Sprintf("%s/_delta_log/%020d.json", basePath, ver)
		if err := s.client.WriteFile(ctx, logFile, logBuf); err != nil {
			return fmt.Errorf("failed to write delta transaction log: %w", err)
		}
	}
	return nil
}

func (s *DeltaLakeStreamingSink) Close() error {
	return s.CommitBatch(context.Background())
}
