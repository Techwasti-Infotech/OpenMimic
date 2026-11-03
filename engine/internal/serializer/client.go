package serializer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	pb "openmimic/engine/pkg/proto/v1"
)

// RegistryHTTPClient abstracts network requests to Confluent / Karapace Schema Registry.
type RegistryHTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// SchemaRegistryClient communicates with Confluent/Karapace Schema Registry for auto-discovery and registration.
type SchemaRegistryClient struct {
	url    string
	client RegistryHTTPClient
}

// NewSchemaRegistryClient creates a Schema Registry API client.
func NewSchemaRegistryClient(rawURL string, client RegistryHTTPClient) *SchemaRegistryClient {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	cleanURL := strings.TrimRight(rawURL, "/")
	return &SchemaRegistryClient{
		url:    cleanURL,
		client: client,
	}
}

// RegisterSchema registers an Avro or JSON Schema under the given subject: POST /subjects/{subject}/versions
func (c *SchemaRegistryClient) RegisterSchema(ctx context.Context, subject, schemaType, schemaStr string) (int, error) {
	endpoint := fmt.Sprintf("%s/subjects/%s/versions", c.url, subject)
	body := map[string]string{
		"schema":     schemaStr,
		"schemaType": schemaType,
	}
	data, err := json.Marshal(body)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewBuffer(data))
	if err != nil {
		return 0, err
	}
	req.Header.Set("Content-Type", "application/vnd.schemaregistry.v1+json")

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("schema registry request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return 0, fmt.Errorf("schema registry returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	var res struct {
		ID int `json:"id"`
	}
	if err := json.Unmarshal(respBytes, &res); err != nil {
		return 0, fmt.Errorf("failed to decode schema registry response: %w", err)
	}

	return res.ID, nil
}

// AutoRegisterWireSerializer wraps WireSerializer with active auto-registration.
type AutoRegisterWireSerializer struct {
	*WireSerializer
	registryClient *SchemaRegistryClient
}

// NewAutoRegisterWireSerializer builds a wire serializer with Schema Registry integration.
func NewAutoRegisterWireSerializer(cfg *pb.SchemaRegistryConfig, client RegistryHTTPClient) *AutoRegisterWireSerializer {
	base := NewWireSerializer(cfg)
	var regClient *SchemaRegistryClient
	if cfg != nil && cfg.GetUrl() != "" {
		regClient = NewSchemaRegistryClient(cfg.GetUrl(), client)
	}
	return &AutoRegisterWireSerializer{
		WireSerializer: base,
		registryClient: regClient,
	}
}

// AutoRegisterIfNecessary queries or posts the schema if schema_id is 0.
func (s *AutoRegisterWireSerializer) AutoRegisterIfNecessary(ctx context.Context, schemaStr string) (uint32, error) {
	if s.cfg == nil || !s.cfg.GetEnabled() {
		return 0, nil
	}
	if s.cfg.GetSchemaId() > 0 {
		return uint32(s.cfg.GetSchemaId()), nil
	}
	if s.registryClient == nil {
		return 1, nil
	}

	subject := s.cfg.GetSubject()
	if subject == "" {
		subject = "openmimic-default-value"
	}
	format := s.cfg.GetFormat()
	if format == "" {
		format = "JSON"
	}

	id, err := s.registryClient.RegisterSchema(ctx, subject, strings.ToUpper(format), schemaStr)
	if err != nil {
		return 0, err
	}
	s.cfg.SchemaId = int32(id)
	return uint32(id), nil
}
