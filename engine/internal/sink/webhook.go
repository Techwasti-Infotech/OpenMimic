package sink

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	pb "openmimic/engine/pkg/proto/v1"
)

// HTTPClient interface allows mocking HTTP calls during tests.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// WebhookSink sends generated payloads as HTTP POST/PUT JSON requests to an endpoint.
type WebhookSink struct {
	client        HTTPClient
	endpoint      string
	method        string
	headers       map[string]string
	retryAttempts int
}

// NewWebhookSink creates a new WebhookSink from configuration.
func NewWebhookSink(cfg *pb.WebhookSink, customClient HTTPClient) *WebhookSink {
	method := "POST"
	if cfg.GetMethod() != "" {
		method = cfg.GetMethod()
	}

	timeout := 5 * time.Second
	if cfg.GetTimeoutMs() > 0 {
		timeout = time.Duration(cfg.GetTimeoutMs()) * time.Millisecond
	}

	retries := int(cfg.GetRetryAttempts())
	if retries <= 0 {
		retries = 1
	}

	client := customClient
	if client == nil {
		client = &http.Client{Timeout: timeout}
	}

	return &WebhookSink{
		client:        client,
		endpoint:      cfg.GetEndpoint(),
		method:        method,
		headers:       cfg.GetHeaders(),
		retryAttempts: retries,
	}
}

// WriteEvent delivers the event payload to the configured webhook endpoint.
func (s *WebhookSink) WriteEvent(ctx context.Context, key []byte, payload map[string]any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt < s.retryAttempts; attempt++ {
		req, err := http.NewRequestWithContext(ctx, s.method, s.endpoint, bytes.NewBuffer(data))
		if err != nil {
			return fmt.Errorf("failed to create http request: %w", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "OpenMimic-Engine/1.0")
		if len(key) > 0 {
			req.Header.Set("X-OpenMimic-Key", string(key))
		}

		for k, v := range s.headers {
			req.Header.Set(k, v)
		}

		resp, err := s.client.Do(req)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				return nil
			}
			lastErr = fmt.Errorf("webhook responded with non-2xx status code: %d", resp.StatusCode)
		} else {
			lastErr = err
		}

		if attempt < s.retryAttempts-1 {
			time.Sleep(time.Duration(20*(attempt+1)) * time.Millisecond)
		}
	}

	return fmt.Errorf("webhook delivery failed after %d attempts: %w", s.retryAttempts, lastErr)
}

// Close is a no-op for WebhookSink.
func (s *WebhookSink) Close() error {
	return nil
}
