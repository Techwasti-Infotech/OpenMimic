package sink

import (
	"context"
	"strings"
	"testing"
)

type mockPostgresDB struct {
	executedSQL  string
	executedArgs []any
	closed       bool
}

func (m *mockPostgresDB) Exec(ctx context.Context, sql string, arguments ...any) error {
	m.executedSQL = sql
	m.executedArgs = arguments
	return nil
}

func (m *mockPostgresDB) Close(ctx context.Context) error {
	m.closed = true
	return nil
}

func TestPostgresSinkWriteEvent(t *testing.T) {
	mockDB := &mockPostgresDB{}
	columns := []string{"id", "email", "score"}
	pgSink := NewPostgresSinkWithDB(mockDB, "users", columns)

	event := map[string]any{
		"id":    "u-123",
		"email": "user@example.com",
		"score": 98.7,
	}

	err := pgSink.WriteEvent(context.Background(), []byte("key1"), event)
	if err != nil {
		t.Fatalf("unexpected WriteEvent error: %v", err)
	}

	if !strings.Contains(mockDB.executedSQL, "INSERT INTO users (id, email, score)") {
		t.Fatalf("unexpected SQL executed: %s", mockDB.executedSQL)
	}

	if len(mockDB.executedArgs) != 3 {
		t.Fatalf("expected 3 arguments, got %d", len(mockDB.executedArgs))
	}
	if mockDB.executedArgs[0] != "u-123" || mockDB.executedArgs[1] != "user@example.com" {
		t.Fatalf("unexpected arguments: %v", mockDB.executedArgs)
	}

	_ = pgSink.Close()
	if !mockDB.closed {
		t.Fatalf("expected mock db to be closed")
	}
}

