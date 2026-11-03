package sink

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	pb "openmimic/engine/pkg/proto/v1"
)

type mockHTTPClient struct {
	statusCode int
	calls      int
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	m.calls++
	return &http.Response{
		StatusCode: m.statusCode,
		Body:       io.NopCloser(bytes.NewBufferString("ok")),
	}, nil
}

func TestWebhookSinkWriteEvent(t *testing.T) {
	mockClient := &mockHTTPClient{statusCode: 200}
	cfg := &pb.WebhookSink{
		Endpoint:      "https://example.com/api/events",
		Method:        "POST",
		Headers:       map[string]string{"Authorization": "Bearer secret_token"},
		RetryAttempts: 2,
	}

	sink := NewWebhookSink(cfg, mockClient)
	defer sink.Close()

	payload := map[string]any{
		"event_id": "evt-12345",
		"status":   "CONFIRMED",
		"amount":   49.99,
	}

	err := sink.WriteEvent(context.Background(), []byte("evt-12345"), payload)
	if err != nil {
		t.Fatalf("unexpected error writing to webhook: %v", err)
	}

	if mockClient.calls != 1 {
		t.Errorf("expected 1 http call, got %d", mockClient.calls)
	}
}

func TestWebhookSinkRetryFailure(t *testing.T) {
	mockClient := &mockHTTPClient{statusCode: 500}
	cfg := &pb.WebhookSink{
		Endpoint:      "https://example.com/api/events",
		RetryAttempts: 3,
	}

	sink := NewWebhookSink(cfg, mockClient)
	defer sink.Close()

	payload := map[string]any{"data": "test"}
	err := sink.WriteEvent(context.Background(), nil, payload)
	if err == nil {
		t.Fatalf("expected error from 500 responses, got nil")
	}

	if mockClient.calls != 3 {
		t.Errorf("expected 3 retry attempts, got %d", mockClient.calls)
	}
}

