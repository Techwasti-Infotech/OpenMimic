package engine

import (
	"testing"

	pb "openmimic/engine/pkg/proto/v1"
)

func TestMultimodalAndNestedPayloadGeneration(t *testing.T) {
	cfg := &pb.GeneratorConfig{
		Id:   "multimodal-gen",
		Seed: 42,
		Fields: []*pb.Field{
			{
				Name: "customer",
				Generator: &pb.FunctionGenerator{
					Func: &pb.FunctionGenerator_NestedObject{
						NestedObject: &pb.NestedObjectFunction{
							Fields: []*pb.Field{
								{
									Name: "id",
									Generator: &pb.FunctionGenerator{
										Func: &pb.FunctionGenerator_Uuid{Uuid: &pb.UuidFunction{}},
									},
								},
								{
									Name: "email",
									Generator: &pb.FunctionGenerator{
										Func: &pb.FunctionGenerator_StringFaker{
											StringFaker: &pb.StringFakerFunction{Category: "email"},
										},
									},
								},
							},
						},
					},
				},
			},
			{
				Name: "tags",
				Generator: &pb.FunctionGenerator{
					Func: &pb.FunctionGenerator_Array{
						Array: &pb.ArrayFunction{
							MinItems: 2,
							MaxItems: 4,
							ItemGenerator: &pb.FunctionGenerator{
								Func: &pb.FunctionGenerator_StringFaker{
									StringFaker: &pb.StringFakerFunction{Category: "city"},
								},
							},
						},
					},
				},
			},
			{
				Name: "avatar_meta",
				Generator: &pb.FunctionGenerator{
					Func: &pb.FunctionGenerator_MediaMeta{
						MediaMeta: &pb.MediaMetaFunction{
							MediaType: "image",
							MimeType:  "image/png",
							Width:     512,
							Height:    512,
						},
					},
				},
			},
		},
	}

	eng := NewEngine(cfg, nil)
	evt, err := eng.GenerateEvent(0)
	if err != nil {
		t.Fatalf("failed to generate event: %v", err)
	}

	// 1. Verify Nested Object
	cust, ok := evt["customer"].(map[string]any)
	if !ok {
		t.Fatalf("expected nested map for 'customer', got %T", evt["customer"])
	}
	if cust["id"] == nil || cust["email"] == nil {
		t.Fatalf("expected nested customer fields, got %v", cust)
	}

	// 2. Verify Repeating Array
	tags, ok := evt["tags"].([]any)
	if !ok {
		t.Fatalf("expected slice for 'tags', got %T", evt["tags"])
	}
	if len(tags) < 2 || len(tags) > 4 {
		t.Fatalf("expected 2-4 items in tags array, got %d", len(tags))
	}

	// 3. Verify Media Metadata
	media, ok := evt["avatar_meta"].(map[string]any)
	if !ok {
		t.Fatalf("expected media metadata map, got %T", evt["avatar_meta"])
	}
	if media["mime_type"] != "image/png" || media["width"] != int32(512) {
		t.Fatalf("unexpected media metadata: %v", media)
	}
}

