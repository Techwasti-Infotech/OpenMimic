package service

import (
	"context"
	"testing"

	pb "openmimic/engine/pkg/proto/v1"
	"google.golang.org/grpc"
)

type mockPreviewStream struct {
	grpc.ServerStream
	ctx       context.Context
	responses []*pb.StreamPreviewResponse
}

func (m *mockPreviewStream) Context() context.Context {
	return m.ctx
}

func (m *mockPreviewStream) Send(resp *pb.StreamPreviewResponse) error {
	m.responses = append(m.responses, resp)
	return nil
}

func TestGeneratorServerStreamPreview(t *testing.T) {
	server := NewGeneratorServer()

	cfg := &pb.GeneratorConfig{
		Id:   "preview-test",
		Seed: 123,
		Fields: []*pb.Field{
			{
				Name: "uuid",
				Generator: &pb.FunctionGenerator{
					Func: &pb.FunctionGenerator_Uuid{Uuid: &pb.UuidFunction{}},
				},
			},
			{
				Name: "value",
				Generator: &pb.FunctionGenerator{
					Func: &pb.FunctionGenerator_Distribution{
						Distribution: &pb.DistributionFunction{
							DistributionType: &pb.DistributionFunction_Normal{
								Normal: &pb.NormalDistribution{Mean: 100, StdDev: 10},
							},
						},
					},
				},
			},
			{
				Name: "created_at",
				Generator: &pb.FunctionGenerator{
					Func: &pb.FunctionGenerator_Now{Now: &pb.NowFunction{Format: "RFC3339"}},
				},
			},
		},
	}

	stream := &mockPreviewStream{
		ctx: context.Background(),
	}

	err := server.StreamPreview(&pb.StreamPreviewRequest{
		Config:       cfg,
		PreviewCount: 10,
	}, stream)

	if err != nil {
		t.Fatalf("StreamPreview error: %v", err)
	}

	if len(stream.responses) != 10 {
		t.Fatalf("expected 10 streamed responses, got %d", len(stream.responses))
	}

	for i, r := range stream.responses {
		if r.Index != int64(i) {
			t.Errorf("resp %d has index %d", i, r.Index)
		}
		if r.JsonPayload == "" {
			t.Errorf("resp %d has empty json payload", i)
		}
	}
}

func TestGeneratorStartAndStop(t *testing.T) {
	server := NewGeneratorServer()

	cfg := &pb.GeneratorConfig{
		Id:         "gen-test-lifecycle",
		Seed:       42,
		EventLimit: 100,
	}

	ctx := context.Background()
	startResp, err := server.StartGenerator(ctx, &pb.StartGeneratorRequest{Config: cfg})
	if err != nil {
		t.Fatalf("StartGenerator failed: %v", err)
	}
	if startResp.Status != "RUNNING" {
		t.Fatalf("expected status RUNNING, got %s", startResp.Status)
	}

	stopResp, err := server.StopGenerator(ctx, &pb.StopGeneratorRequest{GeneratorId: cfg.Id})
	if err != nil {
		t.Fatalf("StopGenerator failed: %v", err)
	}
	if stopResp.Status != "STOPPED" {
		t.Fatalf("expected status STOPPED, got %s", stopResp.Status)
	}
}

