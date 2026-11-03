# Mimic Copilot Manual

**Mimic Copilot** is an AI assistant and translation service designed to convert natural language descriptions of data streams into validated, executable `mimic.json` Protocol Buffers configurations.

---

## 🤖 How It Works

```
User Prompt:
"Generate a stream of 100 user signup events with a 5% failure rate"
                         │
                         ▼
        ┌──────────────────────────────────┐
        │       POST /generate-config      │
        │      (Python 3.12 + FastAPI)     │
        └────────────────┬─────────────────┘
                         │
                         ▼
        ┌──────────────────────────────────┐
        │   Natural Language Parser & LLM  │
        │ Extracts: limits, fields, weights│
        └────────────────┬─────────────────┘
                         │
                         ▼
        ┌──────────────────────────────────┐
        │    Protocol Buffers Validation   │
        │  google.protobuf.json_format     │
        └────────────────┬─────────────────┘
                         │ Validated GeneratorConfig
                         ▼
        ┌──────────────────────────────────┐
        │   Populated in Mimic Studio      │
        │        Ready to Execute          │
        └──────────────────────────────────┘
```

---

## 💡 Example Prompts

### User Signups with Failure Rates
```text
Generate a stream of 100 user signup events with a 5% failure rate
```
**Synthesized Fields:**
- `user_id`: `uuid` v4
- `email`: `string_faker` (email category)
- `status`: `weighted_choice` (`SUCCESS`: 95.0, `FAILURE`: 5.0)
- `signup_timestamp`: `now` (RFC3339)

---

### High-Volume Financial Payments
```text
Generate 10000 payment events with a normal distribution amount mean 150 std dev 20 to Kafka topic transactions.raw
```
**Synthesized Fields:**
- `transaction_id`: `uuid`
- `amount`: `normalDistribution` (mean: 150.0, std_dev: 20.0)
- `timestamp`: `now`

---

## 🛡️ Schema Self-Correction & Validation Guarantee

Every configuration synthesized by Copilot is checked against `copilot/proto/mimic_config_pb2.py`:
1. If the generated JSON payload violates types, missing required fields, or unrecognized sink targets, `json_format.ParseDict` raises an explicit validation exception.
2. The endpoint returns an HTTP 422 if invalid, prompting an automatic retry before any invalid configuration can reach the UI or the engine.

