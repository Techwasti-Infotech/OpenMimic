package conventions

import (
	"strings"

	pb "openmimic/engine/pkg/proto/v1"
)

// Resolver resolves conventions over configurations:
// Applies tenant conventions for missing fields/rates, and enforces Super Admin global guardrails.
type Resolver struct {
	global *pb.GlobalSettings
	tenant *pb.TenantSettings
}

func NewResolver(global *pb.GlobalSettings, tenant *pb.TenantSettings) *Resolver {
	if global == nil {
		global = &pb.GlobalSettings{
			Environment:               "development",
			MaxEventsPerGeneratorCap:  1000000,
			MaxRateLimitPerSecondCap: 5000,
			EnforcePiiMasking:         false,
			AllowedSinks:              []string{"kafka", "postgres", "redis", "clickhouse", "stdout"},
		}
	}
	if tenant == nil {
		tenant = &pb.TenantSettings{
			TenantId:             "tenant_default",
			TenantName:           "Default Tenant",
			DefaultEventLimit:    500,
			DefaultRateLimit:     20,
			DefaultTopicPrefix:   "stream.",
			AutoInjectTimestamps: true,
			AutoInjectEventId:    true,
			DefaultCurrency:      "USD",
		}
	}
	return &Resolver{global: global, tenant: tenant}
}

// ApplyConventions fills in default configuration where omitted by the user, and enforces admin limits.
func (r *Resolver) ApplyConventions(cfg *pb.GeneratorConfig) *pb.GeneratorConfig {
	if cfg == nil {
		cfg = &pb.GeneratorConfig{Id: "gen_auto", Name: "Auto Convention Generator"}
	}

	// 1. Convention: Default event limits
	if cfg.GetEventLimit() == 0 && r.tenant.GetDefaultEventLimit() > 0 {
		cfg.EventLimit = int64(r.tenant.GetDefaultEventLimit())
	}
	// Super Admin hard ceiling cap
	if r.global.GetMaxEventsPerGeneratorCap() > 0 && cfg.GetEventLimit() > int64(r.global.GetMaxEventsPerGeneratorCap()) {
		cfg.EventLimit = int64(r.global.GetMaxEventsPerGeneratorCap())
	}

	// 2. Convention: Default rate limits
	if cfg.GetRateLimitPerSecond() == 0 && r.tenant.GetDefaultRateLimit() > 0 {
		cfg.RateLimitPerSecond = r.tenant.GetDefaultRateLimit()
	}
	// Super Admin rate ceiling cap
	if r.global.GetMaxRateLimitPerSecondCap() > 0 && cfg.GetRateLimitPerSecond() > r.global.GetMaxRateLimitPerSecondCap() {
		cfg.RateLimitPerSecond = r.global.GetMaxRateLimitPerSecondCap()
	}

	// 3. Convention: Automatic Kafka topic prefixing if not provided with prefix
	if cfg.GetSink() != nil {
		if kafkaSink := cfg.GetSink().GetKafka(); kafkaSink != nil {
			prefix := r.tenant.GetDefaultTopicPrefix()
			if prefix != "" && !strings.HasPrefix(kafkaSink.GetTopic(), prefix) {
				kafkaSink.Topic = prefix + kafkaSink.GetTopic()
			}
			if len(kafkaSink.GetBrokers()) == 0 && r.global.GetDefaultKafkaCluster() != "" {
				kafkaSink.Brokers = []string{r.global.GetDefaultKafkaCluster()}
			}
		}
	}

	// 4. Convention: Auto-inject standard metadata fields (event_id, created_at) if enabled
	hasID := false
	hasTimestamp := false
	for _, f := range cfg.GetFields() {
		name := strings.ToLower(f.GetName())
		if name == "id" || name == "event_id" || name == "uuid" {
			hasID = true
		}
		if name == "created_at" || name == "timestamp" || name == "ts" {
			hasTimestamp = true
		}
	}

	if r.tenant.GetAutoInjectEventId() && !hasID {
		idField := &pb.Field{
			Name: "id",
			Generator: &pb.FunctionGenerator{
				Func: &pb.FunctionGenerator_Uuid{
					Uuid: &pb.UuidFunction{Uppercase: false},
				},
			},
		}
		cfg.Fields = append([]*pb.Field{idField}, cfg.Fields...)
	}

	if r.tenant.GetAutoInjectTimestamps() && !hasTimestamp {
		tsField := &pb.Field{
			Name: "created_at",
			Generator: &pb.FunctionGenerator{
				Func: &pb.FunctionGenerator_Now{
					Now: &pb.NowFunction{Format: "RFC3339"},
				},
			},
		}
		cfg.Fields = append(cfg.Fields, tsField)
	}

	return cfg
}

