package sink

import (
	"context"
	"testing"

	"github.com/segmentio/kafka-go"
	pb "openmimic/engine/pkg/proto/v1"
)

type mockTransactionalKafkaProducer struct {
	messages []kafka.Message
	closed   bool
}

func (m *mockTransactionalKafkaProducer) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	m.messages = append(m.messages, msgs...)
	return nil
}

func (m *mockTransactionalKafkaProducer) Close() error {
	m.closed = true
	return nil
}

func TestKafkaTransactionalEOSCommitMarkers(t *testing.T) {
	mockWriter := &mockTransactionalKafkaProducer{}
	cfg := &pb.KafkaSink{
		Brokers: []string{"localhost:9092"},
		Topic:   "tx.orders",
		Transactional: &pb.KafkaTransactionalConfig{
			Enabled:          true,
			TransactionalId:  "tx-producer-01",
			BatchCommitSize:  3, // commit every 3 events
		},
	}

	sink := NewKafkaSink(cfg)
	sink.writer = mockWriter // override with mock writer

	// Write 5 events: should emit 3 events, 1 commit marker, then 2 events, and final commit marker on Close
	for i := 1; i <= 5; i++ {
		err := sink.WriteEvent(context.Background(), []byte("key"), map[string]any{"id": i})
		if err != nil {
			t.Fatalf("failed to write event %d: %v", i, err)
		}
	}

	// At 5 events: 5 data messages + 1 batch commit marker = 6 messages
	if len(mockWriter.messages) != 6 {
		t.Errorf("expected 6 messages before close (5 data + 1 marker), got %d", len(mockWriter.messages))
	}

	if err := sink.Close(); err != nil {
		t.Fatalf("failed to close sink: %v", err)
	}

	// After Close: 5 data messages + 1 batch commit marker + 1 final commit marker = 7 messages
	if len(mockWriter.messages) != 7 {
		t.Fatalf("expected 7 messages after close (5 data + 2 markers), got %d", len(mockWriter.messages))
	}

	lastMsg := mockWriter.messages[len(mockWriter.messages)-1]
	if string(lastMsg.Key) != "__mimic_tx_commit__" {
		t.Errorf("expected commit marker key '__mimic_tx_commit__', got %s", string(lastMsg.Key))
	}
}

