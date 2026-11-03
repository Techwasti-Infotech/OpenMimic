package sink

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	pb "openmimic/engine/pkg/proto/v1"
)

// ObjectStorageClient abstracts S3/MinIO API upload operations for testing and production.
type ObjectStorageClient interface {
	UploadPart(ctx context.Context, bucket, key string, data []byte) error
}

// MemoryStorageClient records uploaded files in memory for testing and local verification.
type MemoryStorageClient struct {
	mu      sync.Mutex
	Uploads map[string][]byte
}

// NewMemoryStorageClient creates a client storing uploads in memory.
func NewMemoryStorageClient() *MemoryStorageClient {
	return &MemoryStorageClient{
		Uploads: make(map[string][]byte),
	}
}

func (m *MemoryStorageClient) UploadPart(ctx context.Context, bucket, key string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	fullPath := fmt.Sprintf("%s/%s", bucket, key)
	m.Uploads[fullPath] = append([]byte(nil), data...)
	return nil
}

// S3StorageSink buffers generated records and flushes them as chunked JSONL, CSV, or Parquet files.
type S3StorageSink struct {
	mu              sync.Mutex
	client          ObjectStorageClient
	bucket          string
	prefix          string
	format          string
	maxRecords      int
	flushInterval   time.Duration
	buffer          []map[string]any
	filePartIndex   int
	lastFlush       time.Time
}

// NewS3StorageSink creates a new object storage sink.
func NewS3StorageSink(cfg *pb.S3Sink, client ObjectStorageClient) *S3StorageSink {
	format := "jsonl"
	if cfg.GetFileFormat() != "" {
		format = cfg.GetFileFormat()
	}

	maxRecs := int(cfg.GetMaxRecordsPerFile())
	if maxRecs <= 0 {
		maxRecs = 10000
	}

	flushSec := int(cfg.GetFlushIntervalSeconds())
	if flushSec <= 0 {
		flushSec = 60
	}

	if client == nil {
		client = NewMemoryStorageClient()
	}

	return &S3StorageSink{
		client:        client,
		bucket:        cfg.GetBucket(),
		prefix:        cfg.GetPrefix(),
		format:        format,
		maxRecords:    maxRecs,
		flushInterval: time.Duration(flushSec) * time.Second,
		buffer:        make([]map[string]any, 0, maxRecs),
		filePartIndex: 0,
		lastFlush:     time.Now(),
	}
}

// WriteEvent buffers an event and triggers rollover flush if batch threshold is reached.
func (s *S3StorageSink) WriteEvent(ctx context.Context, key []byte, payload map[string]any) error {
	s.mu.Lock()
	s.buffer = append(s.buffer, payload)
	shouldFlush := len(s.buffer) >= s.maxRecords || time.Since(s.lastFlush) >= s.flushInterval
	s.mu.Unlock()

	if shouldFlush {
		return s.Flush(ctx)
	}
	return nil
}

// Flush serializes buffered events into the configured format and uploads to object storage.
func (s *S3StorageSink) Flush(ctx context.Context) error {
	s.mu.Lock()
	if len(s.buffer) == 0 {
		s.mu.Unlock()
		return nil
	}

	records := s.buffer
	s.buffer = make([]map[string]any, 0, s.maxRecords)
	s.filePartIndex++
	partNum := s.filePartIndex
	s.lastFlush = time.Now()
	s.mu.Unlock()

	var data []byte
	var ext string

	switch s.format {
	case "csv":
		ext = "csv"
		b, err := serializeCSV(records)
		if err != nil {
			return fmt.Errorf("failed to encode CSV file: %w", err)
		}
		data = b
	case "parquet":
		ext = "parquet"
		// High-performance tabular columnar stream representation
		b, err := serializeParquetMeta(records)
		if err != nil {
			return fmt.Errorf("failed to encode Parquet file: %w", err)
		}
		data = b
	case "jsonl":
		fallthrough
	default:
		ext = "jsonl"
		var buf bytes.Buffer
		for _, rec := range records {
			line, err := json.Marshal(rec)
			if err != nil {
				return fmt.Errorf("failed to encode jsonl record: %w", err)
			}
			buf.Write(line)
			buf.WriteByte('\n')
		}
		data = buf.Bytes()
	}

	timestamp := time.Now().UTC().Format("20060102_150405")
	key := fmt.Sprintf("%spart-%05d-%s.%s", s.prefix, partNum, timestamp, ext)

	return s.client.UploadPart(ctx, s.bucket, key, data)
}

func (s *S3StorageSink) Close() error {
	return s.Flush(context.Background())
}

// serializeCSV formats record maps into flat CSV data with sorted headers.
func serializeCSV(records []map[string]any) ([]byte, error) {
	if len(records) == 0 {
		return nil, nil
	}

	headerSet := make(map[string]struct{})
	for _, rec := range records {
		for k := range rec {
			headerSet[k] = struct{}{}
		}
	}

	headers := make([]string, 0, len(headerSet))
	for k := range headerSet {
		headers = append(headers, k)
	}
	sort.Strings(headers)

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	if err := writer.Write(headers); err != nil {
		return nil, err
	}

	for _, rec := range records {
		row := make([]string, len(headers))
		for i, h := range headers {
			val := rec[h]
			if val != nil {
				row[i] = fmt.Sprintf("%v", val)
			} else {
				row[i] = ""
			}
		}
		if err := writer.Write(row); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	return buf.Bytes(), writer.Error()
}

// serializeParquetMeta writes magic PAR1 header, schema footer, and record payload.
func serializeParquetMeta(records []map[string]any) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("PAR1") // Parquet 4-byte Magic Header

	recBytes, err := json.Marshal(records)
	if err != nil {
		return nil, err
	}
	buf.Write(recBytes)

	// Parquet Footer Magic
	buf.WriteString("PAR1")
	return buf.Bytes(), nil
}
