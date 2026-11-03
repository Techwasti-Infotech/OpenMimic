package conventions

import (
	"testing"

	pb "openmimic/engine/pkg/proto/v1"
)

func TestConventionDefaultsAndAdminCaps(t *testing.T) {
	global := &pb.GlobalSettings{
		MaxEventsPerGeneratorCap: 1000,
		MaxRateLimitPerSecondCap: 50,
	}

	tenant := &pb.TenantSettings{
		DefaultEventLimit:    200,
		DefaultRateLimit:     10,
		DefaultTopicPrefix:   "org.corp.",
		AutoInjectEventId:    true,
		AutoInjectTimestamps: true,
	}

	resolver := NewResolver(global, tenant)

	// User provides minimal config without limit, rate, id, or timestamp
	rawConfig := &pb.GeneratorConfig{
		Id:   "minimal_test",
		Name: "Minimal",
		Sink: &pb.SinkConfig{
			SinkType: &pb.SinkConfig_Kafka{
				Kafka: &pb.KafkaSink{Topic: "orders"},
			},
		},
		Fields: []*pb.Field{
			{Name: "amount", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_Constant{Constant: &pb.ConstantFunction{Value: &pb.ConstantFunction_DoubleVal{DoubleVal: 50.0}}}}},
		},
	}

	resolved := resolver.ApplyConventions(rawConfig)

	// 1. Convention: Tenant default event limit was applied
	if resolved.GetEventLimit() != 200 {
		t.Fatalf("expected event limit 200, got %d", resolved.GetEventLimit())
	}

	// 2. Convention: Tenant default rate limit was applied
	if resolved.GetRateLimitPerSecond() != 10 {
		t.Fatalf("expected rate limit 10, got %d", resolved.GetRateLimitPerSecond())
	}

	// 3. Convention: Topic prefix was prepended
	if resolved.GetSink().GetKafka().GetTopic() != "org.corp.orders" {
		t.Fatalf("expected topic org.corp.orders, got %s", resolved.GetSink().GetKafka().GetTopic())
	}

	// 4. Convention: Auto-injected id and created_at fields
	fieldNames := make([]string, len(resolved.GetFields()))
	for i, f := range resolved.GetFields() {
		fieldNames[i] = f.GetName()
	}

	hasID := false
	hasCreatedAt := false
	for _, n := range fieldNames {
		if n == "id" {
			hasID = true
		}
		if n == "created_at" {
			hasCreatedAt = true
		}
	}

	if !hasID || !hasCreatedAt {
		t.Fatalf("expected auto-injected fields id and created_at, got %v", fieldNames)
	}

	// 5. Super Admin hard ceiling enforcement
	greedyConfig := &pb.GeneratorConfig{
		EventLimit:         9999999,
		RateLimitPerSecond: 10000,
	}
	capped := resolver.ApplyConventions(greedyConfig)
	if capped.GetEventLimit() != 1000 {
		t.Fatalf("expected capped event limit 1000, got %d", capped.GetEventLimit())
	}
	if capped.GetRateLimitPerSecond() != 50 {
		t.Fatalf("expected capped rate limit 50, got %d", capped.GetRateLimitPerSecond())
	}
}

