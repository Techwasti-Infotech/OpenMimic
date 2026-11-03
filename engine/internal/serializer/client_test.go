package serializer

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	pb "openmimic/engine/pkg/proto/v1"
)

type mockRegistryClient struct {
	status int
	body   string
}

func (m *mockRegistryClient) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: m.status,
		Body:       io.NopCloser(bytes.NewBufferString(m.body)),
	}, nil
}

func TestSchemaRegistryAutoRegistration(t *testing.T) {
	mockClient := &mockRegistryClient{
		status: 200,
		body:   `{"id": 42}`,
	}

	cfg := &pb.SchemaRegistryConfig{
		Enabled:  true,
		Url:      "http://localhost:8081",
		Subject:  "orders-value",
		Format:   "avro",
		SchemaId: 0, // Trigger auto-registration
	}

	serializer := NewAutoRegisterWireSerializer(cfg, mockClient)
	id, err := serializer.AutoRegisterIfNecessary(context.Background(), `{"type":"record","name":"Order"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id != 42 {
		t.Errorf("expected registered schema ID 42, got %d", id)
	}

	// Verify serialization uses the newly acquired schema ID
	payload := map[string]any{"order_id": "ord-1"}
	data, err := serializer.Serialize(payload)
	if err != nil {
		t.Fatalf("failed to serialize: %v", err)
	}

	if data[0] != 0x00 {
		t.Errorf("expected magic byte 0x00, got 0x%02x", data[0])
	}
	if data[4] != 42 {
		t.Errorf("expected schema ID byte 42, got %d", data[4])
	}
}
