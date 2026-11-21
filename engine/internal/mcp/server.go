package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strings"

	"openmimic/engine"
	"openmimic/engine/internal/importer"
	"openmimic/engine/internal/profiler"
	pb "openmimic/engine/pkg/proto/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

// Server handles MCP protocol requests over stdio or JSON-RPC streams.
type Server struct {
	in     io.Reader
	out    io.Writer
	tools  map[string]Tool
	engine *serviceRegistry
}

type serviceRegistry struct {
	templates map[string]*pb.GeneratorConfig
}

// NewServer initializes a new OpenMimic MCP server instance.
func NewServer(in io.Reader, out io.Writer) *Server {
	s := &Server{
		in:    in,
		out:   out,
		tools: make(map[string]Tool),
		engine: &serviceRegistry{
			templates: loadPredefinedTemplates(),
		},
	}
	s.registerDefaultTools()
	return s
}

func loadPredefinedTemplates() map[string]*pb.GeneratorConfig {
	return map[string]*pb.GeneratorConfig{
		"ecommerce_order_flow": {
			Id:         "ecommerce-order-flow",
			Name:       "E-Commerce Order Flow",
			EventLimit: 1000,
			Sink: &pb.SinkConfig{
				SinkType: &pb.SinkConfig_Stdout{Stdout: &pb.StdoutSink{Pretty: true}},
			},
			Fields: []*pb.Field{
				{Name: "order_id", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_Uuid{Uuid: &pb.UuidFunction{}}}},
				{Name: "customer_id", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_EntityPool{EntityPool: &pb.EntityPoolFunction{PoolSize: 100, EntityPrefix: "usr_"}}}},
				{Name: "amount", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_Distribution{Distribution: &pb.DistributionFunction{DistributionType: &pb.DistributionFunction_Normal{Normal: &pb.NormalDistribution{Mean: 85.5, StdDev: 22.0}}}}}},
				{Name: "currency", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_WeightedChoice{WeightedChoice: &pb.WeightedChoiceFunction{Options: []*pb.WeightedChoiceOption{{Value: "USD", Weight: 70}, {Value: "EUR", Weight: 20}, {Value: "GBP", Weight: 10}}}}}},
				{Name: "status", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_WeightedChoice{WeightedChoice: &pb.WeightedChoiceFunction{Options: []*pb.WeightedChoiceOption{{Value: "COMPLETED", Weight: 92}, {Value: "PENDING", Weight: 5}, {Value: "FAILED", Weight: 3}}}}}},
				{Name: "created_at", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_Now{Now: &pb.NowFunction{Format: "RFC3339"}}}},
			},
		},
		"user_signups": {
			Id:         "user-signups",
			Name:       "User Signups Stream",
			EventLimit: 500,
			Sink: &pb.SinkConfig{
				SinkType: &pb.SinkConfig_Stdout{Stdout: &pb.StdoutSink{Pretty: true}},
			},
			Fields: []*pb.Field{
				{Name: "user_id", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_Uuid{Uuid: &pb.UuidFunction{}}}},
				{Name: "email", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_StringFaker{StringFaker: &pb.StringFakerFunction{Category: "email"}}}},
				{Name: "country", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_WeightedChoice{WeightedChoice: &pb.WeightedChoiceFunction{Options: []*pb.WeightedChoiceOption{{Value: "US", Weight: 45}, {Value: "DE", Weight: 15}, {Value: "IN", Weight: 25}, {Value: "JP", Weight: 15}}}}}},
				{Name: "signup_at", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_Now{Now: &pb.NowFunction{Format: "RFC3339"}}}},
			},
		},
		"iot_telemetry": {
			Id:         "iot-telemetry-sensors",
			Name:       "Industrial IoT Sensor Stream",
			EventLimit: 5000,
			Sink: &pb.SinkConfig{
				SinkType: &pb.SinkConfig_Stdout{Stdout: &pb.StdoutSink{Pretty: true}},
			},
			Fields: []*pb.Field{
				{Name: "sensor_id", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_EntityPool{EntityPool: &pb.EntityPoolFunction{PoolSize: 50, EntityPrefix: "sensor_edge_"}}}},
				{Name: "temperature_c", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_Distribution{Distribution: &pb.DistributionFunction{DistributionType: &pb.DistributionFunction_Normal{Normal: &pb.NormalDistribution{Mean: 68.2, StdDev: 4.5}}}}}},
				{Name: "vibration_hz", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_Distribution{Distribution: &pb.DistributionFunction{DistributionType: &pb.DistributionFunction_Normal{Normal: &pb.NormalDistribution{Mean: 120.0, StdDev: 15.0}}}}}},
				{Name: "battery_pct", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_Distribution{Distribution: &pb.DistributionFunction{DistributionType: &pb.DistributionFunction_Normal{Normal: &pb.NormalDistribution{Mean: 88.0, StdDev: 8.0}}}}}},
				{Name: "timestamp", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_Now{Now: &pb.NowFunction{Format: "RFC3339"}}}},
			},
		},
		"financial_fraud_stream": {
			Id:         "financial-fraud-stream",
			Name:       "Card Transactions with Fraud Jitter",
			EventLimit: 2000,
			Sink: &pb.SinkConfig{
				SinkType: &pb.SinkConfig_Stdout{Stdout: &pb.StdoutSink{Pretty: true}},
			},
			Fields: []*pb.Field{
				{Name: "txn_id", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_Uuid{Uuid: &pb.UuidFunction{}}}},
				{Name: "account_id", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_EntityPool{EntityPool: &pb.EntityPoolFunction{PoolSize: 200, EntityPrefix: "act_"}}}},
				{Name: "amount", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_Distribution{Distribution: &pb.DistributionFunction{DistributionType: &pb.DistributionFunction_Normal{Normal: &pb.NormalDistribution{Mean: 142.50, StdDev: 85.0}}}}}},
				{Name: "merchant_category", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_WeightedChoice{WeightedChoice: &pb.WeightedChoiceFunction{Options: []*pb.WeightedChoiceOption{{Value: "grocery", Weight: 40}, {Value: "electronics", Weight: 25}, {Value: "travel", Weight: 20}, {Value: "luxury", Weight: 15}}}}}},
				{Name: "is_fraud", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_WeightedChoice{WeightedChoice: &pb.WeightedChoiceFunction{Options: []*pb.WeightedChoiceOption{{Value: "false", Weight: 98.8}, {Value: "true", Weight: 1.2}}}}}},
				{Name: "timestamp", Generator: &pb.FunctionGenerator{Func: &pb.FunctionGenerator_Now{Now: &pb.NowFunction{Format: "RFC3339"}}}},
			},
		},
	}
}

func (s *Server) registerDefaultTools() {
	s.tools["openmimic_generate_preview"] = Tool{
		Name:        "openmimic_generate_preview",
		Description: "Generates realistic synthetic data events directly from an OpenMimic schema configuration (JSON or Protocol Buffer). Returns generated records as JSON.",
		InputSchema: ToolSchema{
			Type: "object",
			Properties: map[string]PropertyDef{
				"config_json": {
					Type:        "string",
					Description: "The JSON string representation of an OpenMimic GeneratorConfig (fields, distributions, sinks, constraints).",
				},
				"count": {
					Type:        "integer",
					Description: "Number of synthetic records to generate (default: 5, max: 100).",
					Default:     5,
				},
			},
			Required: []string{"config_json"},
		},
	}

	s.tools["openmimic_validate_schema"] = Tool{
		Name:        "openmimic_validate_schema",
		Description: "Validates an OpenMimic schema configuration for syntactic correctness, valid field generators, and constraint satisfiability.",
		InputSchema: ToolSchema{
			Type: "object",
			Properties: map[string]PropertyDef{
				"config_json": {
					Type:        "string",
					Description: "The JSON string representation of the OpenMimic GeneratorConfig to validate.",
				},
				"check_constraints": {
					Type:        "boolean",
					Description: "Whether to execute a test event generation to verify predicate and constraint feasibility (default: true).",
					Default:     true,
				},
			},
			Required: []string{"config_json"},
		},
	}

	s.tools["openmimic_sql_to_schema"] = Tool{
		Name:        "openmimic_sql_to_schema",
		Description: "Parses standard SQL DDL ('CREATE TABLE ...') statements into a complete, ready-to-run OpenMimic GeneratorConfig with inferred types, UUIDs, and distributions.",
		InputSchema: ToolSchema{
			Type: "object",
			Properties: map[string]PropertyDef{
				"ddl": {
					Type:        "string",
					Description: "The SQL CREATE TABLE DDL statement.",
				},
			},
			Required: []string{"ddl"},
		},
	}

	s.tools["openmimic_profile_sample_data"] = Tool{
		Name:        "openmimic_profile_sample_data",
		Description: "Reverse-engineers an OpenMimic GeneratorConfig from sample JSON records, automatically inferring data types, numeric means/deviations, categories, and null ratios.",
		InputSchema: ToolSchema{
			Type: "object",
			Properties: map[string]PropertyDef{
				"table_name": {
					Type:        "string",
					Description: "Target name for the generated table schema.",
					Default:     "inferred_table",
				},
				"samples": {
					Type:        "string",
					Description: "A JSON array of sample objects (e.g. '[{\"id\": 1, \"name\": \"Alice\", \"amount\": 42.5}]').",
				},
			},
			Required: []string{"samples"},
		},
	}

	s.tools["openmimic_get_template"] = Tool{
		Name:        "openmimic_get_template",
		Description: "Fetches a pre-built enterprise synthetic data generator template. Available templates: 'ecommerce_order_flow', 'user_signups', 'iot_telemetry', 'financial_fraud_stream'.",
		InputSchema: ToolSchema{
			Type: "object",
			Properties: map[string]PropertyDef{
				"template_name": {
					Type:        "string",
					Description: "Name of the starter template.",
					Enum:        []string{"ecommerce_order_flow", "user_signups", "iot_telemetry", "financial_fraud_stream"},
				},
			},
			Required: []string{"template_name"},
		},
	}

	s.tools["openmimic_list_templates"] = Tool{
		Name:        "openmimic_list_templates",
		Description: "Lists all available pre-built starter templates and their schema configurations.",
		InputSchema: ToolSchema{
			Type:       "object",
			Properties: map[string]PropertyDef{},
		},
	}
}

// Serve runs the JSON-RPC loop over in/out until EOF.
func (s *Server) Serve(ctx context.Context) error {
	scanner := bufio.NewScanner(s.in)
	// Support up to 10MB payloads for large schemas/sample profiling
	buf := make([]byte, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := scanner.Bytes()
		if len(strings.TrimSpace(string(line))) == 0 {
			continue
		}

		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			s.writeError(nil, -32700, "Parse error: "+err.Error())
			continue
		}

		s.handleRequest(&req)
	}

	return scanner.Err()
}

func (s *Server) handleRequest(req *JSONRPCRequest) {
	switch req.Method {
	case "initialize":
		s.handleInitialize(req)
	case "notifications/initialized":
		// No-op for initialized notification
	case "ping":
		s.writeResult(req.ID, map[string]any{})
	case "tools/list":
		s.handleToolsList(req)
	case "tools/call":
		s.handleToolsCall(req)
	case "resources/list":
		s.handleResourcesList(req)
	case "resources/read":
		s.handleResourcesRead(req)
	case "prompts/list":
		s.handlePromptsList(req)
	case "prompts/get":
		s.handlePromptsGet(req)
	default:
		s.writeError(req.ID, -32601, fmt.Sprintf("Method not found: %s", req.Method))
	}
}

func (s *Server) handleInitialize(req *JSONRPCRequest) {
	result := map[string]any{
		"protocolVersion": "2024-11-05",
		"serverInfo": map[string]any{
			"name":    "openmimic-mcp",
			"version": "1.0.0",
		},
		"capabilities": map[string]any{
			"tools":     map[string]any{"listChanged": false},
			"resources": map[string]any{"subscribe": false, "listChanged": false},
			"prompts":   map[string]any{"listChanged": false},
		},
	}
	s.writeResult(req.ID, result)
}

func (s *Server) handleToolsList(req *JSONRPCRequest) {
	list := make([]Tool, 0, len(s.tools))
	for _, t := range s.tools {
		list = append(list, t)
	}
	s.writeResult(req.ID, map[string]any{"tools": list})
}

func (s *Server) handleToolsCall(req *JSONRPCRequest) {
	rawParams, err := json.Marshal(req.Params)
	if err != nil {
		s.writeError(req.ID, -32602, "Invalid params: "+err.Error())
		return
	}

	var callParams struct {
		Name      string         `json:"name"`
		Arguments map[string]any `json:"arguments"`
	}
	if err := json.Unmarshal(rawParams, &callParams); err != nil {
		s.writeError(req.ID, -32602, "Invalid params structure: "+err.Error())
		return
	}

	switch callParams.Name {
	case "openmimic_generate_preview":
		s.callGeneratePreview(req.ID, callParams.Arguments)
	case "openmimic_validate_schema":
		s.callValidateSchema(req.ID, callParams.Arguments)
	case "openmimic_sql_to_schema":
		s.callSQLToSchema(req.ID, callParams.Arguments)
	case "openmimic_profile_sample_data":
		s.callProfileSampleData(req.ID, callParams.Arguments)
	case "openmimic_get_template":
		s.callGetTemplate(req.ID, callParams.Arguments)
	case "openmimic_list_templates":
		s.callListTemplates(req.ID)
	default:
		s.writeToolError(req.ID, fmt.Sprintf("Unknown tool: %s", callParams.Name))
	}
}

func (s *Server) callGeneratePreview(reqID any, args map[string]any) {
	configJSON, _ := args["config_json"].(string)
	if configJSON == "" {
		s.writeToolError(reqID, "Missing required argument 'config_json'")
		return
	}

	count := 5
	if c, ok := args["count"].(float64); ok && c > 0 {
		count = int(math.Min(c, 100))
	}

	var cfg pb.GeneratorConfig
	if err := protojson.Unmarshal([]byte(configJSON), &cfg); err != nil {
		s.writeToolError(reqID, fmt.Sprintf("Failed to parse config_json: %v", err))
		return
	}

	eng := engine.NewEngine(&cfg, nil)
	records := make([]map[string]any, 0, count)
	for i := 0; i < count; i++ {
		evt, err := eng.GenerateEvent(int64(i))
		if err != nil {
			s.writeToolError(reqID, fmt.Sprintf("Generation error at event #%d: %v", i, err))
			return
		}
		records = append(records, evt)
	}

	previewBytes, _ := json.MarshalIndent(records, "", "  ")
	s.writeToolSuccess(reqID, fmt.Sprintf("Successfully generated %d synthetic records:\n\n%s", count, string(previewBytes)))
}

func (s *Server) callValidateSchema(reqID any, args map[string]any) {
	configJSON, _ := args["config_json"].(string)
	if configJSON == "" {
		s.writeToolError(reqID, "Missing required argument 'config_json'")
		return
	}

	var cfg pb.GeneratorConfig
	if err := protojson.Unmarshal([]byte(configJSON), &cfg); err != nil {
		s.writeToolError(reqID, fmt.Sprintf("INVALID SCHEMA: Protobuf/JSON deserialization error: %v", err))
		return
	}

	checkConstraints := true
	if cc, ok := args["check_constraints"].(bool); ok {
		checkConstraints = cc
	}

	if checkConstraints {
		eng := engine.NewEngine(&cfg, nil)
		_, err := eng.GenerateEvent(0)
		if err != nil {
			s.writeToolError(reqID, fmt.Sprintf("INVALID CONFIGURATION: Runtime generation test failed: %v", err))
			return
		}
	}

	s.writeToolSuccess(reqID, fmt.Sprintf("VALID SCHEMA: Verified successfully. Config ID: '%s', Fields: %d, Constraints: %d.",
		cfg.GetId(), len(cfg.GetFields()), len(cfg.GetConstraints())))
}

func (s *Server) callSQLToSchema(reqID any, args map[string]any) {
	ddl, _ := args["ddl"].(string)
	if ddl == "" {
		s.writeToolError(reqID, "Missing required argument 'ddl'")
		return
	}

	cfg, err := importer.SQLDDLToGeneratorConfig(ddl)
	if err != nil {
		s.writeToolError(reqID, fmt.Sprintf("Failed to convert SQL DDL to OpenMimic schema: %v", err))
		return
	}

	outJSON, err := protojson.MarshalOptions{Indent: "  "}.Marshal(cfg)
	if err != nil {
		s.writeToolError(reqID, fmt.Sprintf("Serialization error: %v", err))
		return
	}

	s.writeToolSuccess(reqID, fmt.Sprintf("Successfully converted SQL DDL to OpenMimic GeneratorConfig:\n\n%s", string(outJSON)))
}

func (s *Server) callProfileSampleData(reqID any, args map[string]any) {
	samplesStr, _ := args["samples"].(string)
	if samplesStr == "" {
		s.writeToolError(reqID, "Missing required argument 'samples'")
		return
	}

	tableName := "inferred_table"
	if tn, ok := args["table_name"].(string); ok && tn != "" {
		tableName = tn
	}

	var rows []map[string]any
	if err := json.Unmarshal([]byte(samplesStr), &rows); err != nil {
		s.writeToolError(reqID, fmt.Sprintf("Failed to parse samples JSON array: %v", err))
		return
	}

	if len(rows) == 0 {
		s.writeToolError(reqID, "Samples array cannot be empty")
		return
	}

	p := profiler.NewDatabaseProfiler(tableName, rows)
	cfg, err := p.Profile()
	if err != nil {
		s.writeToolError(reqID, fmt.Sprintf("Profiling failed: %v", err))
		return
	}

	outJSON, err := protojson.MarshalOptions{Indent: "  "}.Marshal(cfg)
	if err != nil {
		s.writeToolError(reqID, fmt.Sprintf("Serialization error: %v", err))
		return
	}

	s.writeToolSuccess(reqID, fmt.Sprintf("Successfully profiled %d rows and generated OpenMimic schema:\n\n%s", len(rows), string(outJSON)))
}

func (s *Server) callGetTemplate(reqID any, args map[string]any) {
	templateName, _ := args["template_name"].(string)
	cfg, exists := s.engine.templates[templateName]
	if !exists {
		s.writeToolError(reqID, fmt.Sprintf("Template '%s' not found. Available: ecommerce_order_flow, user_signups, iot_telemetry, financial_fraud_stream", templateName))
		return
	}

	outJSON, err := protojson.MarshalOptions{Indent: "  "}.Marshal(cfg)
	if err != nil {
		s.writeToolError(reqID, fmt.Sprintf("Serialization error: %v", err))
		return
	}

	s.writeToolSuccess(reqID, fmt.Sprintf("Template '%s':\n\n%s", templateName, string(outJSON)))
}

func (s *Server) callListTemplates(reqID any) {
	summaries := make([]map[string]any, 0, len(s.engine.templates))
	for name, cfg := range s.engine.templates {
		summaries = append(summaries, map[string]any{
			"name":        name,
			"description": cfg.GetName(),
			"field_count": len(cfg.GetFields()),
		})
	}

	outJSON, _ := json.MarshalIndent(summaries, "", "  ")
	s.writeToolSuccess(reqID, fmt.Sprintf("Available OpenMimic Starter Templates:\n\n%s", string(outJSON)))
}

func (s *Server) handleResourcesList(req *JSONRPCRequest) {
	resources := []Resource{
		{
			URI:         "mimic://templates/ecommerce_order_flow",
			Name:        "E-Commerce Order Flow Template",
			Description: "Transactional order events with amounts, currencies, user IDs, and statuses.",
			MimeType:    "application/json",
		},
		{
			URI:         "mimic://templates/user_signups",
			Name:        "User Signups Template",
			Description: "User registrations with email fakers and geographical distributions.",
			MimeType:    "application/json",
		},
		{
			URI:         "mimic://templates/iot_telemetry",
			Name:        "Industrial IoT Telemetry Template",
			Description: "Sensor metrics including temperature, vibration, and battery distributions.",
			MimeType:    "application/json",
		},
		{
			URI:         "mimic://templates/financial_fraud_stream",
			Name:        "Financial Fraud Stream Template",
			Description: "Card transaction streams with stochastic fraud injected.",
			MimeType:    "application/json",
		},
		{
			URI:         "mimic://engine/status",
			Name:        "OpenMimic Engine Status",
			Description: "Status and capabilities of the in-process synthetic engine.",
			MimeType:    "application/json",
		},
	}

	s.writeResult(req.ID, map[string]any{"resources": resources})
}

func (s *Server) handleResourcesRead(req *JSONRPCRequest) {
	rawParams, _ := json.Marshal(req.Params)
	var params struct {
		URI string `json:"uri"`
	}
	_ = json.Unmarshal(rawParams, &params)

	if strings.HasPrefix(params.URI, "mimic://templates/") {
		name := strings.TrimPrefix(params.URI, "mimic://templates/")
		if cfg, ok := s.engine.templates[name]; ok {
			bytes, _ := protojson.MarshalOptions{Indent: "  "}.Marshal(cfg)
			s.writeResult(req.ID, map[string]any{
				"contents": []map[string]any{
					{
						"uri":      params.URI,
						"mimeType": "application/json",
						"text":     string(bytes),
					},
				},
			})
			return
		}
	} else if params.URI == "mimic://engine/status" {
		status := map[string]any{
			"status":            "ONLINE",
			"engine":            "OpenMimic Go Engine",
			"version":           "1.0.0",
			"license":           "Apache 2.0 (100% Free & Open Source)",
			"supported_sinks":   []string{"kafka", "postgres", "s3", "iceberg", "delta", "redis", "clickhouse", "snowflake", "bigquery", "mjpeg_video", "stdout"},
			"traffic_shapers":   []string{"constant", "sine_diurnal", "poisson", "flash_crowd", "burst"},
			"privacy_mechanisms": []string{"laplace", "gaussian", "fpe_numeric", "sha256_salt", "mask_characters"},
		}
		bytes, _ := json.MarshalIndent(status, "", "  ")
		s.writeResult(req.ID, map[string]any{
			"contents": []map[string]any{
				{
					"uri":      params.URI,
					"mimeType": "application/json",
					"text":     string(bytes),
				},
			},
		})
		return
	}

	s.writeError(req.ID, -32602, fmt.Sprintf("Resource not found: %s", params.URI))
}

func (s *Server) handlePromptsList(req *JSONRPCRequest) {
	prompts := []Prompt{
		{
			Name:        "design_synthetic_dataset",
			Description: "Guides an LLM through designing a realistic synthetic data generator for stress-testing databases or streaming message queues.",
			Arguments: []PromptArgument{
				{Name: "domain", Description: "The target business domain (e.g. Fintech, Healthcare, E-Commerce, IoT)", Required: true},
				{Name: "volume_per_second", Description: "Desired generation rate (e.g. 500/s)", Required: false},
			},
		},
		{
			Name:        "convert_sql_to_generator",
			Description: "Guides the conversion of a relational database schema into a high-throughput realistic streaming generator.",
			Arguments: []PromptArgument{
				{Name: "sql_ddl", Description: "The SQL DDL schema", Required: true},
			},
		},
	}
	s.writeResult(req.ID, map[string]any{"prompts": prompts})
}

func (s *Server) handlePromptsGet(req *JSONRPCRequest) {
	rawParams, _ := json.Marshal(req.Params)
	var params struct {
		Name      string            `json:"name"`
		Arguments map[string]string `json:"arguments"`
	}
	_ = json.Unmarshal(rawParams, &params)

	switch params.Name {
	case "design_synthetic_dataset":
		domain := params.Arguments["domain"]
		if domain == "" {
			domain = "General E-Commerce"
		}
		rate := params.Arguments["volume_per_second"]
		if rate == "" {
			rate = "100"
		}

		s.writeResult(req.ID, map[string]any{
			"description": "Design an OpenMimic synthetic dataset configuration",
			"messages": []map[string]any{
				{
					"role": "user",
					"content": map[string]any{
						"type": "text",
						"text": fmt.Sprintf("Please design an OpenMimic GeneratorConfig JSON for domain '%s' running at '%s' events/sec. Use appropriate fields (UUIDs, faker names/emails, normal/log-normal distributions, entity pools for referential integrity) and return valid JSON that can be verified with openmimic_validate_schema.", domain, rate),
					},
				},
			},
		})
	case "convert_sql_to_generator":
		ddl := params.Arguments["sql_ddl"]
		s.writeResult(req.ID, map[string]any{
			"description": "Convert SQL DDL to OpenMimic schema",
			"messages": []map[string]any{
				{
					"role": "user",
					"content": map[string]any{
						"type": "text",
						"text": fmt.Sprintf("Use the tool 'openmimic_sql_to_schema' to convert the following SQL DDL into an OpenMimic generator config, and then preview 5 sample records using 'openmimic_generate_preview':\n\n```sql\n%s\n```", ddl),
					},
				},
			},
		})
	default:
		s.writeError(req.ID, -32602, fmt.Sprintf("Prompt not found: %s", params.Name))
	}
}

func (s *Server) writeToolSuccess(reqID any, text string) {
	res := CallToolResult{
		Content: []ContentItem{
			{Type: "text", Text: text},
		},
	}
	s.writeResult(reqID, res)
}

func (s *Server) writeToolError(reqID any, text string) {
	res := CallToolResult{
		Content: []ContentItem{
			{Type: "text", Text: text},
		},
		IsError: true,
	}
	s.writeResult(reqID, res)
}

func (s *Server) writeResult(reqID any, result any) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      reqID,
		Result:  result,
	}
	bytes, _ := json.Marshal(resp)
	fmt.Fprintf(s.out, "%s\n", string(bytes))
}

func (s *Server) writeError(reqID any, code int, message string) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      reqID,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
		},
	}
	bytes, _ := json.Marshal(resp)
	fmt.Fprintf(s.out, "%s\n", string(bytes))
}
