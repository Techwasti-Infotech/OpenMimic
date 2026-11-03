package service

import (
	"sync"
	"time"

	pb "openmimic/engine/pkg/proto/v1"
)

// PlatformStore manages in-memory workspaces, starter templates, audit logs, and conventions over settings (Phase 9).
type PlatformStore struct {
	mu             sync.RWMutex
	workspaces     map[string]*pb.Workspace
	templates      map[string]*pb.Template
	auditLogs      []*pb.AuditEvent
	globalSettings *pb.GlobalSettings
	tenantSettings map[string]*pb.TenantSettings
}

func NewPlatformStore() *PlatformStore {
	ps := &PlatformStore{
		workspaces:     make(map[string]*pb.Workspace),
		templates:      make(map[string]*pb.Template),
		auditLogs:      make([]*pb.AuditEvent, 0),
		tenantSettings: make(map[string]*pb.TenantSettings),
		globalSettings: &pb.GlobalSettings{
			Environment:              "development",
			MaxEventsPerGeneratorCap: 1000000,
			MaxRateLimitPerSecondCap: 5000,
			EnforcePiiMasking:        false,
			AllowedSinks:             []string{"kafka", "postgres", "redis", "clickhouse", "stdout"},
			DefaultKafkaCluster:      "localhost:9092",
		},
	}
	ps.tenantSettings["tenant_default"] = &pb.TenantSettings{
		TenantId:             "tenant_default",
		TenantName:           "Default Enterprise Tenant",
		DefaultEventLimit:    1000,
		DefaultRateLimit:     25,
		DefaultTopicPrefix:   "mimic.prod.",
		AutoInjectTimestamps: true,
		AutoInjectEventId:    true,
		DefaultCurrency:      "USD",
		DefaultTimezone:      "UTC",
	}
	ps.loadDefaultTemplates()
	return ps
}

func (s *PlatformStore) GetGlobalSettings() *pb.GlobalSettings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.globalSettings
}

func (s *PlatformStore) UpdateGlobalSettings(g *pb.GlobalSettings) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.globalSettings = g
}

func (s *PlatformStore) GetTenantSettings(tenantID string) *pb.TenantSettings {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if ts, ok := s.tenantSettings[tenantID]; ok {
		return ts
	}
	return s.tenantSettings["tenant_default"]
}

func (s *PlatformStore) UpdateTenantSettings(ts *pb.TenantSettings) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tenantSettings[ts.GetTenantId()] = ts
}

func (s *PlatformStore) loadDefaultTemplates() {
	// Template 1: E-commerce Order Flow
	s.templates["tpl_ecommerce"] = &pb.Template{
		Id:       "tpl_ecommerce",
		Title:    "E-commerce Order Stream",
		Category: "ecommerce",
		Config: &pb.GeneratorConfig{
			Id:                 "ecommerce-stream",
			Name:               "E-commerce Orders & Payments",
			Seed:               42,
			EventLimit:         500,
			RateLimitPerSecond: 25,
			Fields: []*pb.Field{
				{Name: "order_id", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_Uuid{Uuid: &pb.UuidFunction{}}}},
				{Name: "customer_id", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_EntityPool{EntityPool: &pb.EntityPoolFunction{PoolName: "customers", PoolSize: 200, EntityPrefix: "cust_"}}}},
				{Name: "amount", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_Distribution{Distribution: &pb.DistributionFunction{DistributionType: &pb.DistributionFunction_Normal{Normal: &pb.NormalDistribution{Mean: 85.0, StdDev: 30.0}}}}}},
				{Name: "currency", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_Constant{Constant: &pb.ConstantFunction{Value: &pb.ConstantFunction_StringVal{StringVal: "USD"}}}}},
				{Name: "created_at", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_Now{Now: &pb.NowFunction{Format: "RFC3339"}}}},
			},
		},
	}

	// Template 2: Financial Fraud & High Value Transactions
	s.templates["tpl_fraud"] = &pb.Template{
		Id:       "tpl_fraud",
		Title:    "Financial Fraud & High Value Transactions",
		Category: "finance",
		Config: &pb.GeneratorConfig{
			Id:                 "fraud-stream",
			Name:               "Fraud Detection Stream",
			Seed:               101,
			EventLimit:         250,
			RateLimitPerSecond: 15,
			Fields: []*pb.Field{
				{Name: "tx_id", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_Uuid{Uuid: &pb.UuidFunction{}}}},
				{Name: "account_id", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_EntityPool{EntityPool: &pb.EntityPoolFunction{PoolName: "accounts", PoolSize: 50, EntityPrefix: "acc_"}}}},
				{Name: "amount", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_Distribution{Distribution: &pb.DistributionFunction{DistributionType: &pb.DistributionFunction_Normal{Normal: &pb.NormalDistribution{Mean: 60000.0, StdDev: 10000.0}}}}}},
				{Name: "fraud_flag", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_Constant{Constant: &pb.ConstantFunction{Value: &pb.ConstantFunction_IntVal{IntVal: 1}}}}},
			},
			Constraints: []*pb.Constraint{
				{
					Id:                 "c_high_value",
					Name:               "High Value Fraud Predicate",
					Expression:         "fraud_flag == 1 && amount > 50000",
					MaxResampleRetries: 100,
				},
			},
		},
	}
}

// LogAudit records an audit event.
func (s *PlatformStore) LogAudit(action, generatorID, userID, details string) *pb.AuditEvent {
	s.mu.Lock()
	defer s.mu.Unlock()

	evt := &pb.AuditEvent{
		Id:          time.Now().Format("20060102150405.000000"),
		UserId:      userID,
		Action:      action,
		GeneratorId: generatorID,
		TimestampMs: time.Now().UnixMilli(),
		Details:     details,
	}
	s.auditLogs = append(s.auditLogs, evt)
	return evt
}

// GetAuditLogs returns the last N audit events.
func (s *PlatformStore) GetAuditLogs(limit int) []*pb.AuditEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	n := len(s.auditLogs)
	if limit <= 0 || limit > n {
		limit = n
	}
	res := make([]*pb.AuditEvent, limit)
	copy(res, s.auditLogs[n-limit:])
	return res
}

// ListTemplates returns all available starter templates.
func (s *PlatformStore) ListTemplates() []*pb.Template {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := make([]*pb.Template, 0, len(s.templates))
	for _, tpl := range s.templates {
		res = append(res, tpl)
	}
	return res
}

