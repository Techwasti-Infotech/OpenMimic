# OpenMimic âš¡

> **The 100% Free & Open-Source, Visual-First Synthetic Data & Traffic Generation Platform.**  
> Effortlessly generate high-throughput streaming datasets, simulate production traffic patterns, and test distributed architectures across Kafka, PostgreSQL, S3, and standard output.

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Report Card](https://img.shields.io/badge/Go-1.23%2B-00ADD8?logo=go)](https://golang.org)
[![React](https://img.shields.io/badge/React-18-61DAFB?logo=react)](https://react.dev)
[![Buy Me A Coffee](https://img.shields.io/badge/Buy%20Me%20a%20Coffee-Donate-orange?logo=buymeacoffee)](https://buymeacoffee.com/techwasti)
[![GitHub Sponsors](https://img.shields.io/badge/Sponsor-GitHub-ea4aaa?logo=github)](https://github.com/sponsors/Techwasti)

---

## ðŸŒŸ Support Open-Source & Buy Us a Coffee

OpenMimic is **100% free, open-source software** released under the **Apache 2.0 license**. Anyone can freely use, modify, distribute, and run OpenMimic for personal, educational, or commercial production workloads with zero licensing fees.

If OpenMimic saves your team time, speeds up your testing, or helps validate your data infrastructure, please consider supporting continued development:

- â˜• **Buy Me a Coffee:** [buymeacoffee.com/techwasti](https://buymeacoffee.com/techwasti)
- ðŸ’– **GitHub Sponsors:** [github.com/sponsors/Techwasti](https://github.com/sponsors/Techwasti)
- â­ **Star this repo:** A simple GitHub star helps spread the word!

---

## âš¡ Key Features

- **ðŸŽ¯ High-Performance Engine (Go 1.23+):** Multithreaded streaming engine generating tens of thousands of records per second directly to sinks.
- **ðŸŽ¨ Mimic Studio IDE:** Modern React 18 + Mantine visual interface for designing schemas, viewing live 10-event streaming previews, inspecting statistical fidelity, and tracking real-time metrics.
- **ðŸ¤– Mimic Copilot:** FastAPI + LangChain AI assistant that converts plain-English prompts into validated Protocol Buffers configurations.
- **âš–ï¸ Predicate Constraints Solver:** Define cross-field business logic (`amount > 50000 && fraud_flag == 1`). Engine samples and evaluates candidate events with bounded retries.
- **ðŸ”¬ Statistical Fidelity & Quality Gates:** Built-in two-sample Kolmogorov-Smirnov test (`KolmogorovSmirnovTest`) and Pearson cross-column correlation (`PearsonCorrelation`) to compare synthetic streams against baseline data.
- **ðŸ›¡ï¸ Privacy & PII Scanner:** Proactive detector highlighting unmasked sensitive data (`email`, `ssn`, `phone`, `card`) with recommended masking transforms.
- **ðŸ’¥ Anomaly Injection Playbooks:** Inject real-world edge cases (null field injection, corrupt payload bytes, burst spikes) at configurable rates.
- **ðŸŒŠ Realistic Traffic Shaper:**
  - **Diurnal (Day/Night) Cycles:** Sine/cosine waves simulating daily traffic curves.
  - **Poisson Arrivals:** Stochastic inter-arrival times ($-\ln(U)/\lambda$).
  - **Burst Spikes:** Periodic surges simulating flash sales or load testing.
- **ðŸ”„ Relational Entity Pools & Markov Chains:**
  - Share customer IDs across multiple topics with bounded entity pools (`EntityPool`).
  - Model multi-step user funnels through probabilistic transition matrices (`MarkovChain`).
- **ðŸ“¥ SQL DDL Schema Auto-Import:** Paste any `CREATE TABLE` statement to auto-generate a matching generator configuration.
- **ðŸ”Œ Universal Sinks:** Apache Kafka (with Confluent Schema Registry `TopicRecordNameStrategy` & EOS Transactions), Apache Iceberg V2 ACID Table Sink, Delta Lake ACID Log Sink, PostgreSQL (`pgx`), Amazon S3 / MinIO (with Snappy/GZIP/ZSTD Columnar Parquet), Snowflake Streaming, Google BigQuery, GraphQL mutations, gRPC streams, Redis, ClickHouse, and NDJSON Stdout.
- **ðŸ“¦ Enterprise Security & Identity:** Single Sign-On (SSO) with Okta, Microsoft Entra ID (Azure AD), Keycloak, and SAML 2.0; automatic user & group provisioning via RFC 7643/7644 SCIM 2.0.
- **🧠 Model Context Protocol (MCP) Server:** Native Go MCP server (`mimic mcp` / `mimic-mcp`) enabling LLMs (Claude Desktop, Cursor, Antigravity, Copilot) to generate preview events, validate schemas, convert SQL DDL, and profile sample records autonomously via standard JSON-RPC 2.0.
- **âš¡ Advanced Streaming:** Complex Event Processing (CEP) window rollups (`sum`, `avg`, `min`, `max`, `count`), synthetic camera video broadcaster (real-time MJPEG frames with motion tracking and OCR timecodes), PCM WAV audio synthesizers, visual DAG canvas, and schema version control with visual diff.

---

## ðŸ–¥ï¸ One-Click Local Setup Guide

We provide one-click scripts for Windows, macOS, and Linux that launch the Engine, Copilot, and Studio UI simultaneously with a single command!

### Option 1: One-Click Launcher

**Windows (PowerShell):**
```powershell
.\run-local.ps1
```

**Windows (Command Prompt / Batch):**
```cmd
run-local.bat
```

**macOS / Linux:**
```bash
chmod +x run-local.sh
./run-local.sh
```

The launcher will verify prerequisites, build the Go binaries, and start:
- **Mimic Studio:** `http://localhost:3000`
- **Mimic Engine REST/Bridge:** `http://localhost:8080`
- **Engine gRPC:** `localhost:50051`
- **Mimic Copilot Service:** `http://localhost:8000`

---

### Option 2: Docker Compose

1. Clone the repository:
   ```bash
   git clone https://github.com/Techwasti/OpenMimic.git
   cd OpenMimic
   ```

2. Start the Engine, Copilot, and Redpanda Kafka broker:
   ```bash
   docker compose up -d
   ```

3. Launch Mimic Studio:
   ```bash
   cd studio
   npm install
   npm run dev
   ```
   Open **http://localhost:3000** in your browser.

---

## ðŸ› ï¸ Mimic GitOps CLI (`mimic-cli`)

OpenMimic includes a full-featured GitOps command-line tool for local testing, automation, and CI/CD pipelines.

### Build CLI Binary
```bash
cd engine
go build -o bin/mimic-cli.exe ./cmd/mimic-cli
```

### CLI Subcommands
- **Validate Schema & Constraints:**
  ```bash
  .\engine\bin\mimic-cli.exe validate --config sample_order_flow.json --constraints
  ```
- **Stream Event Preview to Terminal:**
  ```bash
  .\engine\bin\mimic-cli.exe preview --config sample_order_flow.json --count 5
  ```
- **Dry-Run Event Generation (Null Sink):**
  ```bash
  .\engine\bin\mimic-cli.exe run --config sample_order_flow.json --dry-run
  ```
- **Export Config as Starter Template:**
  ```bash
  .\engine\bin\mimic-cli.exe export --config sample_order_flow.json --to-template
  ```
- **Auto-Discover from SQL DDL:**
  ```bash
  .\engine\bin\mimic-cli.exe discover --sql "CREATE TABLE users (id UUID, email VARCHAR, age INT);"
  ```

---

## ðŸ“‹ Sample Configuration (`mimic.json`)

```json
{
  "id": "gen-signups-traffic",
  "name": "User Signups with Bursts & Constraints",
  "seed": 42,
  "event_limit": 1000,
  "rate_limit_per_second": 20,
  "sink": {
    "kafka": {
      "brokers": ["localhost:9092"],
      "topic": "user.signups",
      "client_id": "openmimic-engine"
    }
  },
  "constraints": [
    {
      "id": "c_amount",
      "name": "Positive Amount",
      "expression": "amount > 0",
      "max_resample_retries": 100
    }
  ],
  "playbooks": [
    {
      "name": "Corrupt Payload Edge Case",
      "anomaly_rate": 0.05,
      "injection_type": "corrupt_payload"
    }
  ],
  "traffic_profile": {
    "burst": {
      "normal_rate": 10,
      "burst_rate": 100,
      "interval_seconds": 20,
      "duration_seconds": 5
    }
  },
  "fields": [
    { "name": "user_id", "generator": { "entity_pool": { "pool_size": 100, "entity_prefix": "usr_" } } },
    { "name": "email", "generator": { "string_faker": { "category": "email" } } },
    { "name": "amount", "generator": { "distribution": { "normal": { "mean": 120.0, "std_dev": 25.0 } } } },
    { "name": "signup_timestamp", "generator": { "now": { "format": "RFC3339" } } }
  ]
}
```

---

## ðŸ¤ Contributing

We welcome community contributions! Please check out [CONTRIBUTING.md](CONTRIBUTING.md) to get started with our development guidelines, issue templates, and pull request workflows.

---

## ðŸ“„ License

OpenMimic is 100% open-source software licensed under the **Apache License 2.0**. See the [LICENSE](LICENSE) file for details.
