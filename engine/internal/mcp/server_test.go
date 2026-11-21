package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestMCPServerInitialize(t *testing.T) {
	in := strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}` + "\n")
	var out bytes.Buffer

	srv := NewServer(in, &out)
	if err := srv.Serve(context.Background()); err != nil {
		t.Fatalf("Server error: %v", err)
	}

	var resp JSONRPCResponse
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("Unmarshal error: %v, raw: %s", err, out.String())
	}

	if resp.Error != nil {
		t.Fatalf("Unexpected error: %v", resp.Error)
	}

	resMap, ok := resp.Result.(map[string]any)
	if !ok {
		t.Fatalf("Result is not map[string]any: %v", resp.Result)
	}

	if resMap["protocolVersion"] != "2024-11-05" {
		t.Errorf("Expected protocolVersion 2024-11-05, got %v", resMap["protocolVersion"])
	}
}

func TestMCPToolsList(t *testing.T) {
	in := strings.NewReader(`{"jsonrpc":"2.0","id":2,"method":"tools/list"}` + "\n")
	var out bytes.Buffer

	srv := NewServer(in, &out)
	_ = srv.Serve(context.Background())

	var resp JSONRPCResponse
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	resMap := resp.Result.(map[string]any)
	tools := resMap["tools"].([]any)
	if len(tools) < 6 {
		t.Errorf("Expected at least 6 tools, got %d", len(tools))
	}
}

func TestMCPToolGetTemplateAndGeneratePreview(t *testing.T) {
	req1 := `{"jsonrpc":"2.0","id":10,"method":"tools/call","params":{"name":"openmimic_get_template","arguments":{"template_name":"user_signups"}}}` + "\n"
	
	in := strings.NewReader(req1)
	var out bytes.Buffer

	srv := NewServer(in, &out)
	_ = srv.Serve(context.Background())

	var resp JSONRPCResponse
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	resMap := resp.Result.(map[string]any)
	content := resMap["content"].([]any)
	firstItem := content[0].(map[string]any)
	text := firstItem["text"].(string)

	if !strings.Contains(text, "user-signups") {
		t.Errorf("Expected template text to contain 'user-signups', got: %s", text)
	}

	// Now test openmimic_generate_preview using that schema
	// Extract JSON portion from text
	schemaJSON := text[strings.Index(text, "{"):]
	// Escape or build payload
	callPayload := map[string]any{
		"jsonrpc": "2.0",
		"id":      11,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "openmimic_generate_preview",
			"arguments": map[string]any{
				"config_json": schemaJSON,
				"count":       3,
			},
		},
	}
	req2Bytes, _ := json.Marshal(callPayload)

	in2 := bytes.NewReader(append(req2Bytes, '\n'))
	var out2 bytes.Buffer
	srv2 := NewServer(in2, &out2)
	_ = srv2.Serve(context.Background())

	var resp2 JSONRPCResponse
	if err := json.Unmarshal(out2.Bytes(), &resp2); err != nil {
		t.Fatalf("Unmarshal preview response error: %v", err)
	}

	resMap2 := resp2.Result.(map[string]any)
	content2 := resMap2["content"].([]any)
	previewText := content2[0].(map[string]any)["text"].(string)

	if !strings.Contains(previewText, "Successfully generated 3 synthetic records") {
		t.Errorf("Expected preview text to indicate 3 synthetic records, got: %s", previewText)
	}
	if !strings.Contains(previewText, "email") || !strings.Contains(previewText, "user_id") {
		t.Errorf("Expected preview records to have email and user_id fields, got: %s", previewText)
	}
}

func TestMCPToolSQLToSchema(t *testing.T) {
	ddl := "CREATE TABLE users (id UUID PRIMARY KEY, name VARCHAR(255), balance DECIMAL(10,2), created_at TIMESTAMP)"
	callPayload := map[string]any{
		"jsonrpc": "2.0",
		"id":      20,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "openmimic_sql_to_schema",
			"arguments": map[string]any{
				"ddl": ddl,
			},
		},
	}
	reqBytes, _ := json.Marshal(callPayload)
	in := bytes.NewReader(append(reqBytes, '\n'))
	var out bytes.Buffer

	srv := NewServer(in, &out)
	_ = srv.Serve(context.Background())

	var resp JSONRPCResponse
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	resMap := resp.Result.(map[string]any)
	content := resMap["content"].([]any)
	text := content[0].(map[string]any)["text"].(string)

	if !strings.Contains(text, "users") || !strings.Contains(text, "balance") {
		t.Errorf("Expected converted schema to have table 'users' and field 'balance', got: %s", text)
	}
}

func TestMCPToolProfileSampleData(t *testing.T) {
	samples := `[{"id": "usr_1", "age": 25, "active": true}, {"id": "usr_2", "age": 30, "active": false}]`
	callPayload := map[string]any{
		"jsonrpc": "2.0",
		"id":      30,
		"method":  "tools/call",
		"params": map[string]any{
			"name": "openmimic_profile_sample_data",
			"arguments": map[string]any{
				"table_name": "customer_profiles",
				"samples":    samples,
			},
		},
	}
	reqBytes, _ := json.Marshal(callPayload)
	in := bytes.NewReader(append(reqBytes, '\n'))
	var out bytes.Buffer

	srv := NewServer(in, &out)
	_ = srv.Serve(context.Background())

	var resp JSONRPCResponse
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	resMap := resp.Result.(map[string]any)
	content := resMap["content"].([]any)
	text := content[0].(map[string]any)["text"].(string)

	if !strings.Contains(text, "customer_profiles") {
		t.Errorf("Expected profiled schema to have customer_profiles, got: %s", text)
	}
}

func TestMCPResourcesAndPrompts(t *testing.T) {
	// Test resources/list
	inRes := strings.NewReader(`{"jsonrpc":"2.0","id":40,"method":"resources/list"}` + "\n")
	var outRes bytes.Buffer
	srvRes := NewServer(inRes, &outRes)
	_ = srvRes.Serve(context.Background())

	var respRes JSONRPCResponse
	_ = json.Unmarshal(outRes.Bytes(), &respRes)
	resList := respRes.Result.(map[string]any)["resources"].([]any)
	if len(resList) < 5 {
		t.Errorf("Expected at least 5 resources, got %d", len(resList))
	}

	// Test prompts/list
	inPrompt := strings.NewReader(`{"jsonrpc":"2.0","id":41,"method":"prompts/list"}` + "\n")
	var outPrompt bytes.Buffer
	srvPrompt := NewServer(inPrompt, &outPrompt)
	_ = srvPrompt.Serve(context.Background())

	var respPrompt JSONRPCResponse
	_ = json.Unmarshal(outPrompt.Bytes(), &respPrompt)
	promptList := respPrompt.Result.(map[string]any)["prompts"].([]any)
	if len(promptList) < 2 {
		t.Errorf("Expected at least 2 prompts, got %d", len(promptList))
	}
}
