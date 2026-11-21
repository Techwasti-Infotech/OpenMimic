---
layout: default
title: "Configuration Schema Reference (`mimic.json`)"
---

# Configuration Schema Reference (`mimic.json`)

`mimic.json` is the configuration file for the OpenMimic generation engine. It is strictly typed and governed by `proto/mimic_config.proto`.

---

## 📄 Complete Example

```json
{
  "id": "gen-signups-prod",
  "name": "User Signups Stream",
  "seed": 42,
  "event_limit": 1000,
  "rate_limit_per_second": 25,
  "sink": {
    "kafka": {
      "brokers": ["localhost:9092"],
      "topic": "user.signups",
      "client_id": "openmimic-engine"
    }
  },
  "conventions": {
    "tenant_id": "tenant-corp-a",
    "topic_namespace": "prod.us-east",
    "inject_audit_columns": true
  },
  "telemetry": {
    "enable_w3c_traceparent": true,
    "metrics_port": 8080
  },
  "dlq": {
    "sink_type": "DLQ_SINK_BUFFER",
    "dlq_topic": "dlq.invalid_events",
    "max_buffer_size": 1000
  },
  "fields": [
    {
      "name": "user_id",
      "generator": {
        "uuid": { "uppercase": false }
      }
    },
    {
      "name": "email",
      "generator": {
        "string_faker": { "category": "email" }
      }
    },
    {
      "name": "status",
      "generator": {
        "weighted_choice": {
          "options": [
            { "value": "SUCCESS", "weight": 95.0 },
            { "value": "FAILURE", "weight": 5.0 }
          ]
        }
      }
    },
    {
      "name": "created_at",
      "generator": {
        "now": { "format": "RFC3339" }
      }
    }
  ]
}
```

---

## ⚙️ Top-Level Properties

| Property | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `id` | `string` | Optional | Unique identifier for the generator instance. Auto-generated if omitted. |
| `name` | `string` | Required | Human-readable label displayed in Studio. |
| `seed` | `int64` | Optional | PRNG seed for deterministic output. Defaults to current epoch nanoseconds. |
| `event_limit` | `int64` | Optional | Total events to generate. `0` indicates an unbounded, infinite stream. |
| `rate_limit_per_second` | `int32` | Optional | Max events generated per second. `0` indicates maximum throughput without artificial sleep. |
| `sink` | `SinkConfig` | Required | Destination where events will be delivered. |
| `fields` | `repeated Field` | Required | List of field definitions to synthesize. |

---

## 🎯 Sink Configuration

The `sink` object accepts one of the following definitions:

### 1. Kafka Sink
```json
"sink": {
  "kafka": {
    "brokers": ["localhost:9092"],
    "topic": "events.raw",
    "client_id": "mimic-client",
    "schema_registry": {
      "enabled": true,
      "url": "http://localhost:8081",
      "format": "avro",
      "subject_naming_strategy": "TopicRecordNameStrategy",
      "record_name": "io.confluent.examples.OrderEvent"
    },
    "transactional": {
      "enabled": true,
      "transactional_id": "mimic-tx-producer-01",
      "transaction_timeout_ms": 60000,
      "batch_commit_size": 100
    }
  }
}
```
- **`subject_naming_strategy`**: Supports `TopicNameStrategy` (default: `<topic>-value`), `RecordNameStrategy` (`<record_name>`), and `TopicRecordNameStrategy` (`<topic>-<record_name>`).
- **`transactional`**: Emits transactional Kafka Exactly-Once Semantics (EOS) commit markers with atomic two-phase commit batches.

### 2. PostgreSQL Sink
```json
"sink": {
  "postgres": {
    "connection_string": "postgres://user:pass@localhost:5432/db",
    "table": "events",
    "columns": ["id", "email", "status"]
  }
}
```
*Note: If `columns` is omitted, the payload is written directly to a single `data` JSONB column.*

### 3. Amazon S3 / Cloudflare R2 / MinIO Sink
```json
"sink": {
  "s3": {
    "bucket": "data-lake-raw",
    "region": "us-east-1",
    "prefix": "incoming/events/",
    "file_format": "parquet",
    "compression_codec": "zstd",
    "max_records_per_file": 10000,
    "flush_interval_seconds": 60
  }
}
```
- **`file_format`**: `"jsonl"`, `"csv"`, or `"parquet"`.
- **`compression_codec`**: `"none"`, `"snappy"`, `"gzip"`, or `"zstd"`. High-performance columnar pivoting with dictionary encoding.

### 4. Snowflake Streaming Sink
```json
"sink": {
  "snowflake": {
    "account_identifier": "xy12345.us-east-1",
    "warehouse": "COMPUTE_WH",
    "database": "ANALYTICS_DB",
    "schema": "PUBLIC",
    "table": "RAW_EVENTS",
    "batch_size": 500,
    "flush_interval_ms": 5000
  }
}
```

### 5. Google BigQuery Streaming Sink
```json
"sink": {
  "bigquery": {
    "project_id": "my-gcp-project",
    "dataset_id": "telemetry",
    "table_id": "streaming_events",
    "batch_size": 500,
    "write_mode": "STORAGE_WRITE_API"
  }
}
```

### 6. Apache Iceberg Streaming Lakehouse Sink
```json
"sink": {
  "iceberg": {
    "catalog_uri": "http://localhost:8181",
    "warehouse": "s3://mimic-lakehouse",
    "database": "analytics",
    "table": "user_orders",
    "format_version": "2",
    "commit_interval_ms": 5000
  }
}
```
- **`format_version`**: `"2"` for Iceberg V2 ACID tables supporting row-level commits and snapshots.
- Emits real-time columnar Parquet data files alongside snapshot metadata (`v1.metadata.json`, `v2.metadata.json`).

### 7. Delta Lake Streaming ACID Sink
```json
"sink": {
  "delta_lake": {
    "table_uri": "s3://lakehouse/delta/events",
    "mode": "append",
    "batch_size": 1000,
    "generate_delta_log": true
  }
}
```
- **`generate_delta_log`**: Writes atomic JSON commit manifests directly to `_delta_log/00000000000000000000.json`.

### 8. GraphQL & gRPC Sinks
- **`graphql`**: Streaming mutations over HTTP/HTTPS with custom operation templates.
- **`grpc`**: Native protocol buffer streaming directly to gRPC ingest services.

### 9. Enterprise SSO & Identity Provider (`sso_config`)
```json
"sso_config": {
  "enabled": true,
  "provider_type": "OKTA",
  "issuer_url": "https://techwasti.okta.com/oauth2/default",
  "client_id": "0oae9k2z45OpenMimicClient",
  "group_role_mappings": {
    "DataPlatform-Admins": "super_admin",
    "Data-Engineers": "tenant_admin",
    "Security-Auditors": "viewer"
  },
  "default_role": "viewer",
  "enforce_pkce": true
}
```
- **`provider_type`**: `"OIDC"`, `"OKTA"`, `"AZURE_AD"`, `"KEYCLOAK"`, or `"SAML"`.
- **`group_role_mappings`**: Dynamic RBAC mapping from corporate IdP groups to OpenMimic roles (`super_admin`, `tenant_admin`, `viewer`).

### 8. Redis Sink
```json
"sink": {
  "redis": {
    "address": "localhost:6379",
    "key_prefix": "events:order:",
    "key_field": "order_id",
    "ttl_seconds": 3600
  }
}
```

### 9. ClickHouse Sink
```json
"sink": {
  "clickhouse": {
    "address": "localhost:9000",
    "database": "default",
    "table": "events_stream",
    "batch_size": 1000
  }
}
```

### 10. Stdout Sink
```json
"sink": {
  "stdout": {
    "pretty": false
  }
}
```
Outputs NDJSON (newline-delimited JSON) to standard output.

---

## ⚡ Scale-Out Cluster & CLI Capabilities

### 1. Distributed Partitioned Generation
Run across $N$ parallel worker agents partitioned deterministically:
```bash
mimic-cli cluster --config mimic.json --workers 4 --total 1000000 --rate 5000
```

### 2. Reverse-Engineering from Sample DB Records
Profile existing JSON database records to automatically produce generator definitions:
```bash
mimic-cli profile --table orders --sample sample_db_rows.json
```

### 3. Prometheus & OpenTelemetry Telemetry
Scrape standard Prometheus metrics and propagate W3C distributed trace context:
```bash
curl http://localhost:8080/metrics
```
Headers injected into outgoing messages:
- `traceparent`: `00-<trace_id>-<span_id>-01`
- `tracestate`: vendor tracing metadata


