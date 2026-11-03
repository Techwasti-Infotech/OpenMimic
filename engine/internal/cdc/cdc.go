package cdc

import (
	"fmt"

	"openmimic/engine/internal/state"
	pb "openmimic/engine/pkg/proto/v1"
)

// DebeziumEnvelope wraps standard synthetic records into Debezium CDC message formats.
type DebeziumEnvelope struct {
	cfg *pb.CDCEnvelope
}

func NewDebeziumEnvelope(cfg *pb.CDCEnvelope) *DebeziumEnvelope {
	return &DebeziumEnvelope{cfg: cfg}
}

// Wrap produces a Debezium-compatible envelope with 'before', 'after', 'source', 'op', and 'ts_ms'.
func (d *DebeziumEnvelope) Wrap(record map[string]any, index int64, rng *state.DeterministicRand, clock *state.Clock) map[string]any {
	if d.cfg == nil || !d.cfg.GetEnabled() {
		return record
	}

	connector := d.cfg.GetConnectorType()
	if connector == "" {
		connector = "postgresql"
	}
	db := d.cfg.GetDatabaseName()
	if db == "" {
		db = "production"
	}
	schema := d.cfg.GetSchemaName()
	if schema == "" {
		schema = "public"
	}
	table := d.cfg.GetTableName()
	if table == "" {
		table = "events"
	}

	// Determine operation (c = create, u = update, d = delete, r = read)
	op := "c"
	r := rng.Float64()
	insRatio := d.cfg.GetInsertRatio()
	updRatio := d.cfg.GetUpdateRatio()
	delRatio := d.cfg.GetDeleteRatio()

	if insRatio == 0 && updRatio == 0 && delRatio == 0 {
		insRatio = 0.70
		updRatio = 0.20
		delRatio = 0.10
	}

	var before any = nil
	var after any = nil

	if r < insRatio {
		op = "c"
		after = record
	} else if r < (insRatio + updRatio) {
		op = "u"
		// Simulate modified field for update
		beforeRecord := make(map[string]any)
		for k, v := range record {
			beforeRecord[k] = v
		}
		before = beforeRecord
		after = record
	} else {
		op = "d"
		before = record
		after = nil
	}

	tsMs := clock.Now().UnixMilli()

	source := map[string]any{
		"version":   "2.5.0.Final",
		"connector": connector,
		"name":      fmt.Sprintf("%s_cdc", db),
		"ts_ms":     tsMs,
		"snapshot":  "false",
		"db":        db,
		"schema":    schema,
		"table":     table,
		"txId":      1000 + index,
		"lsn":       20000 + index*100,
	}

	return map[string]any{
		"before": before,
		"after":  after,
		"source": source,
		"op":     op,
		"ts_ms":  tsMs,
	}
}
