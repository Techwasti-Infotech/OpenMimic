package importer

import (
	"testing"
)

func TestSQLDDLToGeneratorConfig(t *testing.T) {
	ddl := `
	CREATE TABLE customers (
		customer_id UUID PRIMARY KEY,
		full_name VARCHAR(255) NOT NULL,
		email_address VARCHAR(255) UNIQUE,
		credit_limit DECIMAL(10, 2),
		status VARCHAR(50),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);
	`

	cfg, err := SQLDDLToGeneratorConfig(ddl)
	if err != nil {
		t.Fatalf("unexpected DDL import error: %v", err)
	}

	if cfg.Name != "Table: customers" {
		t.Fatalf("expected table name 'Table: customers', got '%s'", cfg.Name)
	}

	if len(cfg.Fields) != 6 {
		t.Fatalf("expected 6 fields extracted, got %d", len(cfg.Fields))
	}

	names := make(map[string]bool)
	for _, f := range cfg.Fields {
		names[f.Name] = true
	}

	expectedCols := []string{"customer_id", "full_name", "email_address", "credit_limit", "status", "created_at"}
	for _, expected := range expectedCols {
		if !names[expected] {
			t.Errorf("missing expected column: %s", expected)
		}
	}
}
