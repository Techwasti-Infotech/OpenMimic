package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"openmimic/engine/internal/sink"
	pb "openmimic/engine/pkg/proto/v1"
)

func TestEngineGeneration(t *testing.T) {
	cfg := &pb.GeneratorConfig{
		Id:         "test-gen-1",
		Name:       "Test Generator",
		Seed:       42,
		EventLimit: 5,
		Fields: []*pb.Field{
			{
				Name: "id",
				Generator: &pb.FunctionGenerator{
					Func: &pb.FunctionGenerator_Uuid{
						Uuid: &pb.UuidFunction{Uppercase: false},
					},
				},
			},
			{
				Name: "value",
				Generator: &pb.FunctionGenerator{
					Func: &pb.FunctionGenerator_Distribution{
						Distribution: &pb.DistributionFunction{
							DistributionType: &pb.DistributionFunction_Normal{
								Normal: &pb.NormalDistribution{
									Mean:   100.0,
									StdDev: 15.0,
								},
							},
						},
					},
				},
			},
			{
				Name: "timestamp",
				Generator: &pb.FunctionGenerator{
					Func: &pb.FunctionGenerator_Now{
						Now: &pb.NowFunction{Format: "RFC3339"},
					},
				},
			},
			{
				Name: "seq",
				Generator: &pb.FunctionGenerator{
					Func: &pb.FunctionGenerator_SequentialInteger{
						SequentialInteger: &pb.SequentialIntegerFunction{Start: 1, Step: 1},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	eng := NewEngine(cfg, &buf)

	err := eng.Run(context.Background())
	if err != nil {
		t.Fatalf("unexpected engine error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 5 {
		t.Fatalf("expected 5 output lines, got %d", len(lines))
	}

	for i, line := range lines {
		var record map[string]any
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatalf("failed to unmarshal JSON line %d: %v", i, err)
		}

		if record["id"] == nil || record["id"] == "" {
			t.Errorf("line %d: missing 'id'", i)
		}
		if record["value"] == nil {
			t.Errorf("line %d: missing 'value'", i)
		}
		if record["timestamp"] == nil {
			t.Errorf("line %d: missing 'timestamp'", i)
		}
		if record["seq"] != float64(i+1) {
			t.Errorf("line %d: expected seq %d, got %v", i, i+1, record["seq"])
		}
	}
}

func TestReproducibleOutput(t *testing.T) {
	cfg := &pb.GeneratorConfig{
		Seed:       12345,
		EventLimit: 3,
		Fields: []*pb.Field{
			{
				Name: "uuid",
				Generator: &pb.FunctionGenerator{
					Func: &pb.FunctionGenerator_Uuid{Uuid: &pb.UuidFunction{}},
				},
			},
		},
	}

	var buf1, buf2 bytes.Buffer
	eng1 := NewEngine(cfg, &buf1)
	eng2 := NewEngine(cfg, &buf2)

	_ = eng1.Run(context.Background())
	_ = eng2.Run(context.Background())

	if buf1.String() != buf2.String() {
		t.Fatalf("expected deterministic output for identical seed\nRun 1: %s\nRun 2: %s", buf1.String(), buf2.String())
	}
}

func TestKafkaIntegrationWithMockBroker(t *testing.T) {
	cfg := &pb.GeneratorConfig{
		Id:         "kafka-test-gen",
		Seed:       999,
		EventLimit: 10,
		Sink: &pb.SinkConfig{
			SinkType: &pb.SinkConfig_Kafka{
				Kafka: &pb.KafkaSink{
					Brokers: []string{"localhost:9092"},
					Topic:   "user.signups",
				},
			},
		},
		Fields: []*pb.Field{
			{
				Name: "user_id",
				Generator: &pb.FunctionGenerator{
					Func: &pb.FunctionGenerator_Uuid{Uuid: &pb.UuidFunction{}},
				},
			},
			{
				Name: "status",
				Generator: &pb.FunctionGenerator{
					Func: &pb.FunctionGenerator_WeightedChoice{
						WeightedChoice: &pb.WeightedChoiceFunction{
							Options: []*pb.WeightedChoiceOption{
								{Value: "SUCCESS", Weight: 95.0},
								{Value: "FAILURE", Weight: 5.0},
							},
						},
					},
				},
			},
		},
	}

	mockProducer := &sink.MockKafkaProducer{}
	ks := sink.NewKafkaSinkWithProducer(mockProducer, "user.signups")

	eng := NewEngine(cfg, nil)
	eng.SetSink(ks)

	err := eng.Run(context.Background())
	if err != nil {
		t.Fatalf("engine run to Kafka sink failed: %v", err)
	}

	if len(mockProducer.Messages) != 10 {
		t.Fatalf("expected 10 messages produced, got %d", len(mockProducer.Messages))
	}

	for i, msg := range mockProducer.Messages {
		var payload map[string]any
		if err := json.Unmarshal(msg.Value, &payload); err != nil {
			t.Fatalf("msg %d unmarshal failed: %v", i, err)
		}
		if payload["user_id"] == nil || payload["user_id"] == "" {
			t.Errorf("msg %d missing user_id", i)
		}
		if payload["status"] != "SUCCESS" && payload["status"] != "FAILURE" {
			t.Errorf("msg %d invalid status value: %v", i, payload["status"])
		}
	}

	// Repeat run with identical seed into a second mock producer and verify determinism
	mockProducer2 := &sink.MockKafkaProducer{}
	ks2 := sink.NewKafkaSinkWithProducer(mockProducer2, "user.signups")

	eng2 := NewEngine(cfg, nil)
	eng2.SetSink(ks2)
	_ = eng2.Run(context.Background())

	for i := 0; i < len(mockProducer.Messages); i++ {
		if string(mockProducer.Messages[i].Value) != string(mockProducer2.Messages[i].Value) {
			t.Fatalf("message %d differed across identical seeded runs: %s vs %s",
				i, string(mockProducer.Messages[i].Value), string(mockProducer2.Messages[i].Value))
		}
	}
}
