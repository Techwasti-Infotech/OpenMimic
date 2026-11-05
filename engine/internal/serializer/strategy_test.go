package serializer

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	pb "openmimic/engine/pkg/proto/v1"
)

func TestSubjectNamingStrategies(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *pb.SchemaRegistryConfig
		topic    string
		expected string
	}{
		{
			name:     "Nil config defaults to topic-value",
			cfg:      nil,
			topic:    "sensor-readings",
			expected: "sensor-readings-value",
		},
		{
			name: "Explicit subject overrides everything",
			cfg: &pb.SchemaRegistryConfig{
				Subject: "custom-override-subject",
			},
			topic:    "orders",
			expected: "custom-override-subject",
		},
		{
			name: "TopicNameStrategy for value (default)",
			cfg: &pb.SchemaRegistryConfig{
				SubjectNamingStrategy: "TopicNameStrategy",
				IsKey:                 false,
			},
			topic:    "orders",
			expected: "orders-value",
		},
		{
			name: "TopicNameStrategy for key",
			cfg: &pb.SchemaRegistryConfig{
				SubjectNamingStrategy: "TopicNameStrategy",
				IsKey:                 true,
			},
			topic:    "orders",
			expected: "orders-key",
		},
		{
			name: "RecordNameStrategy with custom record name",
			cfg: &pb.SchemaRegistryConfig{
				SubjectNamingStrategy: "RecordNameStrategy",
				RecordName:            "com.example.ecommerce.OrderEvent",
			},
			topic:    "orders",
			expected: "com.example.ecommerce.OrderEvent",
		},
		{
			name: "TopicRecordNameStrategy for value",
			cfg: &pb.SchemaRegistryConfig{
				SubjectNamingStrategy: "TopicRecordNameStrategy",
				RecordName:            "OrderEvent",
				IsKey:                 false,
			},
			topic:    "ecommerce.orders",
			expected: "ecommerce.orders-OrderEvent",
		},
		{
			name: "TopicRecordNameStrategy for key",
			cfg: &pb.SchemaRegistryConfig{
				SubjectNamingStrategy: "TopicRecordNameStrategy",
				RecordName:            "OrderKey",
				IsKey:                 true,
			},
			topic:    "ecommerce.orders",
			expected: "ecommerce.orders-OrderKey-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeSubject(tt.cfg, tt.topic)
			if got != tt.expected {
				t.Errorf("ComputeSubject() = %q, want %q", got, tt.expected)
			}
		})
	}
}

type mockCustomHTTPClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockCustomHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.doFunc(req)
}

func TestAutoRegisterWithTopicStrategy(t *testing.T) {
	mockClient := &mockCustomHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			if !strings.Contains(req.URL.Path, "/subjects/payments-PaymentEvent/versions") {
				t.Fatalf("unexpected endpoint: %s", req.URL.Path)
			}
			jsonResp := `{"id": 42}`
			return &http.Response{
				StatusCode: 200,
				Body:       io.NopCloser(strings.NewReader(jsonResp)),
			}, nil
		},
	}

	cfg := &pb.SchemaRegistryConfig{
		Enabled:               true,
		Url:                   "http://mock-registry:8081",
		SubjectNamingStrategy: "TopicRecordNameStrategy",
		RecordName:            "PaymentEvent",
		Format:                "avro",
	}

	serializer := NewAutoRegisterWireSerializer(cfg, mockClient)
	id, err := serializer.AutoRegisterWithTopicIfNecessary(context.Background(), "payments", `{"type":"record","name":"PaymentEvent"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 42 {
		t.Fatalf("expected id 42, got %d", id)
	}
}
