package serializer

import (
	"encoding/json"
	"testing"

	pb "openmimic/engine/pkg/proto/v1"
)

func TestWireSerializerConfluentHeader(t *testing.T) {
	cfg := &pb.SchemaRegistryConfig{
		Enabled:  true,
		Format:   "avro",
		SchemaId: 42,
		Subject:  "orders-value",
	}

	ser := NewWireSerializer(cfg)
	payload := map[string]any{
		"order_id": "ord_999",
		"amount":   125.5,
	}

	encoded, err := ser.Serialize(payload)
	if err != nil {
		t.Fatalf("unexpected serialization error: %v", err)
	}

	// 1. Verify 5-byte header
	if len(encoded) < 5 {
		t.Fatalf("expected at least 5 bytes for wire format, got %d", len(encoded))
	}
	if encoded[0] != 0x00 {
		t.Fatalf("expected magic byte 0x00, got 0x%02x", encoded[0])
	}

	// 2. Verify schema ID extraction
	schemaID, ok := ExtractSchemaID(encoded)
	if !ok || schemaID != 42 {
		t.Fatalf("expected schema ID 42, got %d (ok=%v)", schemaID, ok)
	}

	// 3. Verify payload payload decodability
	var decoded map[string]any
	if err := json.Unmarshal(encoded[5:], &decoded); err != nil {
		t.Fatalf("failed to decode framed payload: %v", err)
	}
	if decoded["order_id"] != "ord_999" {
		t.Fatalf("expected order_id ord_999, got %v", decoded["order_id"])
	}
}
