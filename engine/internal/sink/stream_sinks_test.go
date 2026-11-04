package sink

import (
	"context"
	"testing"

	pb "openmimic/engine/pkg/proto/v1"
)

func TestGraphQLSinkWriteEvent(t *testing.T) {
	mockClient := &mockHTTPClient{statusCode: 200}
	cfg := &pb.GraphQLSink{
		Endpoint:          "https://example.com/graphql",
		QueryOrMutation:   "mutation AddEvent($input: EventInput!) { addEvent(input: $input) { id } }",
		OperationName:     "AddEvent",
		Headers:           map[string]string{"Authorization": "Bearer gql_token"},
	}

	sink := NewGraphQLSink(cfg, mockClient)
	payload := map[string]any{"user_id": "u100", "score": 98.6}

	err := sink.WriteEvent(context.Background(), nil, payload)
	if err != nil {
		t.Fatalf("unexpected error in graphql sink: %v", err)
	}

	if mockClient.calls != 1 {
		t.Errorf("expected 1 http call, got %d", mockClient.calls)
	}
}

func TestGrpcStreamingSinkWriteEvent(t *testing.T) {
	mockClient := &MemoryGrpcStreamClient{}
	cfg := &pb.GrpcStreamSink{
		Target: "localhost:50051",
	}

	sink := NewGrpcStreamingSink(cfg, mockClient)
	payload := map[string]any{"sensor_id": "temp_01", "reading": 23.4}

	err := sink.WriteEvent(context.Background(), nil, payload)
	if err != nil {
		t.Fatalf("unexpected error in grpc streaming sink: %v", err)
	}

	if len(mockClient.Messages) != 1 {
		t.Fatalf("expected 1 message in stream client, got %d", len(mockClient.Messages))
	}

	if err := sink.Close(); err != nil {
		t.Fatalf("failed to close grpc stream sink: %v", err)
	}

	if !mockClient.Closed {
		t.Errorf("expected client to be closed")
	}
}
