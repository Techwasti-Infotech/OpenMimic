package sink

import (
	"context"
	"sync"

	"github.com/segmentio/kafka-go"
)

// MockKafkaProducer records messages in memory for integration testing.
type MockKafkaProducer struct {
	Mu       sync.Mutex
	Messages []kafka.Message
	Closed   bool
}

// NewMockKafkaProducer creates an in-memory MockKafkaProducer.
func NewMockKafkaProducer() *MockKafkaProducer {
	return &MockKafkaProducer{}
}

func (m *MockKafkaProducer) WriteMessages(ctx context.Context, msgs ...kafka.Message) error {
	m.Mu.Lock()
	defer m.Mu.Unlock()
	m.Messages = append(m.Messages, msgs...)
	return nil
}

func (m *MockKafkaProducer) Close() error {
	m.Mu.Lock()
	defer m.Mu.Unlock()
	m.Closed = true
	return nil
}

