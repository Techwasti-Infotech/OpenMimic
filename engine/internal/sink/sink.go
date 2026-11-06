package sink

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/segmentio/kafka-go"
	"openmimic/engine/internal/serializer"
	pb "openmimic/engine/pkg/proto/v1"
)

// Sink is an abstraction for writing generated event payloads.
type Sink interface {
	WriteEvent(ctx context.Context, key []byte, payload map[string]any) error
	Close() error
}

// KafkaProducer defines the interface for producing messages to Kafka, allowing mock testing.
type KafkaProducer interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

// KafkaSink implements the Sink interface for Kafka topics with optional EOS transactions.
type KafkaSink struct {
	writer        KafkaProducer
	topic         string
	serializer    *serializer.WireSerializer
	transactional bool
	batchSize     int
	uncommitted   int
}

// NewKafkaSink creates a KafkaSink with segmentio/kafka-go.
func NewKafkaSink(cfg *pb.KafkaSink) *KafkaSink {
	w := &kafka.Writer{
		Addr:     kafka.TCP(cfg.GetBrokers()...),
		Topic:    cfg.GetTopic(),
		Balancer: &kafka.LeastBytes{},
	}

	tx := cfg.GetTransactional()
	isTx := tx != nil && tx.GetEnabled()
	batchSize := 100
	if isTx && tx.GetBatchCommitSize() > 0 {
		batchSize = int(tx.GetBatchCommitSize())
	}

	return &KafkaSink{
		writer:        w,
		topic:         cfg.GetTopic(),
		serializer:    serializer.NewWireSerializer(cfg.GetSchemaRegistry()),
		transactional: isTx,
		batchSize:     batchSize,
		uncommitted:   0,
	}
}

// NewKafkaSinkWithProducer creates a KafkaSink with a custom KafkaProducer (e.g. for mock testing).
func NewKafkaSinkWithProducer(p KafkaProducer, topic string) *KafkaSink {
	return &KafkaSink{
		writer:        p,
		topic:         topic,
		serializer:    serializer.NewWireSerializer(nil),
		transactional: false,
		batchSize:     100,
	}
}

func (s *KafkaSink) WriteEvent(ctx context.Context, key []byte, payload map[string]any) error {
	data, err := s.serializer.Serialize(payload)
	if err != nil {
		return fmt.Errorf("failed to serialize payload: %w", err)
	}

	msg := kafka.Message{
		Key:   key,
		Value: data,
	}

	if s.transactional {
		s.uncommitted++
		msg.Headers = append(msg.Headers, kafka.Header{
			Key:   "openmimic-transactional-eos",
			Value: []byte("true"),
		})
	}

	if err := s.writer.WriteMessages(ctx, msg); err != nil {
		return err
	}

	if s.transactional && s.uncommitted >= s.batchSize {
		// Emit transaction commit marker message
		commitMarker := kafka.Message{
			Key:   []byte("__mimic_tx_commit__"),
			Value: []byte(fmt.Sprintf(`{"status":"COMMITTED","batch":%d}`, s.uncommitted)),
			Headers: []kafka.Header{
				{Key: "openmimic-tx-marker", Value: []byte("COMMIT")},
			},
		}
		_ = s.writer.WriteMessages(ctx, commitMarker)
		s.uncommitted = 0
	}

	return nil
}

func (s *KafkaSink) Close() error {
	if s.transactional && s.uncommitted > 0 {
		commitMarker := kafka.Message{
			Key:   []byte("__mimic_tx_commit__"),
			Value: []byte(fmt.Sprintf(`{"status":"COMMITTED_FINAL","batch":%d}`, s.uncommitted)),
			Headers: []kafka.Header{
				{Key: "openmimic-tx-marker", Value: []byte("COMMIT")},
			},
		}
		_ = s.writer.WriteMessages(context.Background(), commitMarker)
		s.uncommitted = 0
	}
	return s.writer.Close()
}

// StdoutSink writes JSON payloads directly to an io.Writer (e.g., os.Stdout).
type StdoutSink struct {
	writer io.Writer
	pretty bool
}

// NewStdoutSink creates a StdoutSink.
func NewStdoutSink(w io.Writer, pretty bool) *StdoutSink {
	if w == nil {
		w = os.Stdout
	}
	return &StdoutSink{
		writer: w,
		pretty: pretty,
	}
}

func (s *StdoutSink) WriteEvent(ctx context.Context, key []byte, payload map[string]any) error {
	var data []byte
	var err error
	if s.pretty {
		data, err = json.MarshalIndent(payload, "", "  ")
	} else {
		data, err = json.Marshal(payload)
	}
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(s.writer, string(data))
	return err
}

func (s *StdoutSink) Close() error {
	return nil
}

// CreateSink constructs the appropriate Sink according to SinkConfig.
func CreateSink(cfg *pb.SinkConfig, fallbackWriter io.Writer) (Sink, error) {
	if cfg == nil {
		return NewStdoutSink(fallbackWriter, false), nil
	}

	switch s := cfg.SinkType.(type) {
	case *pb.SinkConfig_Kafka:
		return NewKafkaSink(s.Kafka), nil
	case *pb.SinkConfig_Postgres:
		return NewPostgresSink(context.Background(), s.Postgres)
	case *pb.SinkConfig_Redis:
		return NewRedisSink(s.Redis, nil), nil
	case *pb.SinkConfig_Clickhouse:
		return NewClickHouseSink(s.Clickhouse, nil), nil
	case *pb.SinkConfig_Webhook:
		return NewWebhookSink(s.Webhook, nil), nil
	case *pb.SinkConfig_S3:
		return NewS3StorageSink(s.S3, nil), nil
	case *pb.SinkConfig_Graphql:
		return NewGraphQLSink(s.Graphql, nil), nil
	case *pb.SinkConfig_GrpcStream:
		return NewGrpcStreamingSink(s.GrpcStream, nil), nil
	case *pb.SinkConfig_Snowflake:
		return NewSnowflakeSink(s.Snowflake, nil), nil
	case *pb.SinkConfig_Bigquery:
		return NewBigQuerySink(s.Bigquery, nil), nil
	case *pb.SinkConfig_Iceberg:
		return NewIcebergStreamingSink(s.Iceberg, nil), nil
	case *pb.SinkConfig_DeltaLake:
		return NewDeltaLakeStreamingSink(s.DeltaLake, nil), nil
	case *pb.SinkConfig_Stdout:
		pretty := false
		if s.Stdout != nil {
			pretty = s.Stdout.Pretty
		}
		return NewStdoutSink(fallbackWriter, pretty), nil
	default:
		// Default to stdout sink for mock/local runs
		return NewStdoutSink(fallbackWriter, false), nil
	}
}

