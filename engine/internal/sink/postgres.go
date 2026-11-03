package sink

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	pb "openmimic/engine/pkg/proto/v1"
)

// PostgresDB is an interface abstraction over pgx.Conn to enable mock testing.
type PostgresDB interface {
	Exec(ctx context.Context, sql string, arguments ...any) error
	Close(ctx context.Context) error
}

type pgxConnWrapper struct {
	conn *pgx.Conn
}

func (w *pgxConnWrapper) Exec(ctx context.Context, sql string, arguments ...any) error {
	_, err := w.conn.Exec(ctx, sql, arguments...)
	return err
}

func (w *pgxConnWrapper) Close(ctx context.Context) error {
	return w.conn.Close(ctx)
}

// PostgresSink implements the Sink interface for PostgreSQL database tables.
type PostgresSink struct {
	db      PostgresDB
	table   string
	columns []string
}

// NewPostgresSink connects to PostgreSQL via pgx.
func NewPostgresSink(ctx context.Context, cfg *pb.PostgresSink) (*PostgresSink, error) {
	conn, err := pgx.Connect(ctx, cfg.GetConnectionString())
	if err != nil {
		return nil, fmt.Errorf("unable to connect to postgres: %w", err)
	}

	return &PostgresSink{
		db:      &pgxConnWrapper{conn: conn},
		table:   cfg.GetTable(),
		columns: cfg.GetColumns(),
	}, nil
}

// NewPostgresSinkWithDB constructs a PostgresSink with a custom or mock PostgresDB.
func NewPostgresSinkWithDB(db PostgresDB, table string, columns []string) *PostgresSink {
	return &PostgresSink{
		db:      db,
		table:   table,
		columns: columns,
	}
}

func (s *PostgresSink) WriteEvent(ctx context.Context, key []byte, payload map[string]any) error {
	if len(s.columns) == 0 {
		// Fallback: Store the entire event as a JSONB payload
		query := fmt.Sprintf("INSERT INTO %s (data) VALUES ($1)", s.table)
		rawJson, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		return s.db.Exec(ctx, query, string(rawJson))
	}

	colNames := strings.Join(s.columns, ", ")
	placeholders := make([]string, len(s.columns))
	values := make([]any, len(s.columns))

	for i, col := range s.columns {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		values[i] = payload[col]
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		s.table,
		colNames,
		strings.Join(placeholders, ", "),
	)

	return s.db.Exec(ctx, query, values...)
}

func (s *PostgresSink) Close() error {
	return s.db.Close(context.Background())
}

