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
    "client_id": "mimic-client"
  }
}
```

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

### 3. Amazon S3 Sink
```json
"sink": {
  "s3": {
    "bucket": "data-lake-raw",
    "region": "us-east-1",
    "prefix": "incoming/events/",
    "file_format": "json"
  }
}
```

### 4. Redis Sink
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

### 5. ClickHouse Sink
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

### 6. Stdout Sink
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


