package sink

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	pb "openmimic/engine/pkg/proto/v1"
)

// GraphQLSink sends generated payloads as variables in GraphQL mutations/queries.
type GraphQLSink struct {
	client        HTTPClient
	endpoint      string
	query         string
	operationName string
	headers       map[string]string
}

// NewGraphQLSink creates a new GraphQL sink.
func NewGraphQLSink(cfg *pb.GraphQLSink, client HTTPClient) *GraphQLSink {
	timeout := 5 * time.Second
	if cfg.GetTimeoutMs() > 0 {
		timeout = time.Duration(cfg.GetTimeoutMs()) * time.Millisecond
	}
	if client == nil {
		client = &http.Client{Timeout: timeout}
	}
	return &GraphQLSink{
		client:        client,
		endpoint:      cfg.GetEndpoint(),
		query:         cfg.GetQueryOrMutation(),
		operationName: cfg.GetOperationName(),
		headers:       cfg.GetHeaders(),
	}
}

// WriteEvent executes the GraphQL mutation with the generated event payload as $input.
func (s *GraphQLSink) WriteEvent(ctx context.Context, key []byte, payload map[string]any) error {
	reqBody := map[string]any{
		"query":     s.query,
		"variables": map[string]any{"input": payload},
	}
	if s.operationName != "" {
		reqBody["operationName"] = s.operationName
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal graphql body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.endpoint, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range s.headers {
		req.Header.Set(k, v)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("graphql request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("graphql returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}
	return nil
}

func (s *GraphQLSink) Close() error {
	return nil
}
