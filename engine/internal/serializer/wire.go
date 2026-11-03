package serializer

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strings"

	pb "openmimic/engine/pkg/proto/v1"
)

// WireSerializer formats generated records into Schema Registry wire format bytes.
// Confluent wire format spec:
// Byte 0: Magic byte (0x00)
// Bytes 1-4: 4-byte Schema ID (big-endian uint32)
// Bytes 5+: Serialized payload (Avro / JSON Schema / Protobuf)
type WireSerializer struct {
	cfg *pb.SchemaRegistryConfig
}

func NewWireSerializer(cfg *pb.SchemaRegistryConfig) *WireSerializer {
	return &WireSerializer{cfg: cfg}
}

// Serialize encodes payload according to Schema Registry specifications.
func (s *WireSerializer) Serialize(payload map[string]any) ([]byte, error) {
	if s.cfg == nil || !s.cfg.GetEnabled() {
		return json.Marshal(payload)
	}

	rawBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	format := strings.ToLower(s.cfg.GetFormat())
	if format == "" {
		format = "json_schema"
	}

	schemaID := uint32(s.cfg.GetSchemaId())
	if schemaID == 0 {
		schemaID = 1 // Default schema ID 1
	}

	// 5-byte Confluent framing header: [0x00, ID_B1, ID_B2, ID_B3, ID_B4]
	header := make([]byte, 5)
	header[0] = 0x00
	binary.BigEndian.PutUint32(header[1:5], schemaID)

	switch format {
	case "avro", "json_schema", "protobuf":
		// Prepend magic byte + schema id header to payload
		framed := append(header, rawBytes...)
		return framed, nil
	default:
		return rawBytes, nil
	}
}

// ExtractSchemaID parses the schema ID from a Confluent wire framed byte slice.
func ExtractSchemaID(data []byte) (uint32, bool) {
	if len(data) < 5 || data[0] != 0x00 {
		return 0, false
	}
	schemaID := binary.BigEndian.Uint32(data[1:5])
	return schemaID, true
}
