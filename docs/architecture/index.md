---
layout: default
title: "Architecture & Design Concepts"
---

# Architecture & Design Concepts

OpenMimic is designed around a **schema-first, deterministic generation engine** communicating over gRPC with modern visual and AI interfaces.

---

## 🏛️ High-Level System Architecture

```
                    ┌───────────────────────────┐
                    │       Mimic Studio        │
                    │  (React 18 + Mantine UI)  │
                    └─────────────┬─────────────┘
                                  │
                  ┌───────────────┴───────────────┐
                  │                               │
       HTTP /api/*│                               │ HTTP /generate-config
                  ▼                               ▼
       ┌─────────────────────┐         ┌─────────────────────┐
       │   OpenMimic Engine  │         │    Mimic Copilot    │
       │   (Go 1.23+ Core)   │         │ (FastAPI/LangChain) │
       │  gRPC :50051 + HTTP │         │      Port: 8000     │
       └──────────┬──────────┘         └──────────┬──────────┘
                  │                               │
                  │ Validated by                  │ Validated by
                  ▼                               ▼
       ┌─────────────────────────────────────────────────────┐
       │              proto/mimic_config.proto               │
       │           (Single Source of Truth)                  │
       └─────────────────────────────────────────────────────┘
                  │
        Streaming Sinks
                  ├── Kafka (:9092)
                  ├── PostgreSQL (pgx)
                  ├── Amazon S3
                  └── Stdout (JSON stream)
```

---

## 🔑 Core Architecture Pillars

### 1. Single Source of Truth (`proto3`)
All configuration models, sink destinations, function generators, and service RPCs originate in `proto/mimic_config.proto`.
- Go bindings are generated via `protoc-gen-go` and `protoc-gen-go-grpc`.
- Python bindings are generated via `grpc_tools.protoc`.
- TypeScript definitions mirror the proto structures for frontend type safety.

### 2. Deterministic State Engine
Replayability and test verification require that the exact same stream can be reproduced across multiple runs or distributed workers:
- **Seeded PRNG:** Every engine instance initializes a dedicated `DeterministicRand` seeded with `config.seed`.
- **Logical Clock:** The clock increments deterministically on each event tick (`Clock.Tick()`), ensuring that timestamps advance uniformly regardless of system wall clock variations or CPU throttling.
- **State Store:** Persistent variables can be stored and referenced across event boundaries (e.g. tracking cumulative balances, customer IDs, or error states).

### 3. High-Throughput Sinks
The engine supports hot-swappable streaming sinks implementing the `Sink` interface:
- **`KafkaSink`:** High-performance batched producing via `segmentio/kafka-go`.
- **`PostgresSink`:** Connection-pooled multi-row insert execution with `jackc/pgx/v5`.
- **`StdoutSink`:** High-speed NDJSON output for pipe processing (`mimic | jq`).

### 4. Zero-Overhead Observability
The internal metrics collector records:
- Atomic cumulative event counts.
- Error counts and failure distributions.
- Sliding-window throughput calculations (events/sec).
- P99 and mean latency sampling for every write operation.
- Telemetry is streamed to clients via gRPC `StreamMetrics`.

