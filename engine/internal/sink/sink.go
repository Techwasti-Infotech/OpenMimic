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

// KafkaSink implements the Sink interface for Kafka topics.
type KafkaSink struct {
	writer     KafkaProducer
	topic      string
	serializer *serializer.WireSerializer
}

// NewKafkaSink creates a KafkaSink with segmentio/kafka-go.
func NewKafkaSink(cfg *pb.KafkaSink) *KafkaSink {
	w := &kafka.Writer{
		Addr:     kafka.TCP(cfg.GetBrokers()...),
		Topic:    cfg.GetTopic(),
		Balancer: &kafka.LeastBytes{},
	}
	return &KafkaSink{
		writer:     w,
		topic:      cfg.GetTopic(),
		serializer: serializer.NewWireSerializer(cfg.GetSchemaRegistry()),
	}
}

// NewKafkaSinkWithProducer creates a KafkaSink with a custom KafkaProducer (e.g. for mock testing).
func NewKafkaSinkWithProducer(p KafkaProducer, topic string) *KafkaSink {
	return &KafkaSink{
		writer:     p,
		topic:      topic,
		serializer: serializer.NewWireSerializer(nil),
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
	return s.writer.WriteMessages(ctx, msg)
}

func (s *KafkaSink) Close() error {
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

