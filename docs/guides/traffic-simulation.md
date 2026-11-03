# Traffic Simulation & Load Generation Guide

This comprehensive guide explains how to configure and execute production-grade traffic simulation, load testing, and streaming synthetic data generation with OpenMimic.

---

## 🎯 What is Production Traffic Simulation?

Generating synthetic data for local unit testing is simple, but testing production streaming systems (Kafka, Flink, Spark, PostgreSQL) requires replicating real-world operational conditions:
- **Temporal Realism:** Events should not arrive at flat rates; they follow day/night cycles, network bursts, and stochastic arrival patterns.
- **Referential Integrity:** Generated events across distinct topics or tables must reference the same entities (e.g. realistic repeat users).
- **Behavioral Funnels:** Users move through probabilistic state transitions rather than random actions.

OpenMimic provides built-in primitives to configure these behaviors declaratively.

---

## 🌊 1. Traffic Shaping Models

You can attach a `traffic_profile` to any generator configuration.

### A. Diurnal (Day/Night) Cycle
Simulates natural human diurnal patterns where traffic peaks during business hours and drops to a baseline at night using smooth harmonic sine curves:

```json
"traffic_profile": {
  "diurnal": {
    "base_rate": 5,
    "peak_rate": 100,
    "period_seconds": 86400
  }
}
```
- `base_rate`: Lowest events per second (night-time trough).
- `peak_rate`: Peak events per second (daytime crest).
- `period_seconds`: Cycle length (e.g. 86400 for 24 hours, or 60 for local rapid testing).

---

### B. Poisson Arrival Process (Stochastic Network Jitter)
Models authentic web requests and network packet arrivals where the time between events follows an exponential distribution:

```json
"traffic_profile": {
  "poisson": {
    "lambda_rate": 50.0
  }
}
```
- `lambda_rate`: The expected average throughput (events/second). Inter-arrival intervals vary naturally ($-\ln(U)/\lambda$).

---

### C. Flash Spikes & Burst Surges
Simulates periodic bursts such as flash sales, viral product drops, or DDOS surges:

```json
"traffic_profile": {
  "burst": {
    "normal_rate": 10,
    "burst_rate": 200,
    "interval_seconds": 60,
    "duration_seconds": 10
  }
}
```
Every 60 seconds, traffic will surge from 10 events/sec to 200 events/sec for 10 seconds before returning to baseline.

---

## 🔄 2. Relational Entity Pools (Returning Customers)

In standard random mock tools, every event generates a completely disconnected random UUID. In production, 80% of activity comes from repeat users.

Use `entity_pool` to enforce foreign-key consistency:

```json
{
  "name": "user_id",
  "generator": {
    "entity_pool": {
      "pool_name": "registered_users",
      "pool_size": 500,
      "entity_prefix": "usr_"
    }
  }
}
```
Output: All events across `orders`, `logins`, and `reviews` streams will consistently pick from `usr_00000` through `usr_00499`.

---

## 🧭 3. User Journeys via Markov Chains

To simulate a conversion funnel where actions depend on the prior state:

```json
{
  "name": "funnel_state",
  "generator": {
    "markov_chain": {
      "initial_state": "HOME",
      "states": [
        {
          "state_name": "HOME",
          "transitions": [
            { "to_state": "PRODUCT_SEARCH", "probability": 0.7 },
            { "to_state": "BOUNCE", "probability": 0.3 }
          ]
        },
        {
          "state_name": "PRODUCT_SEARCH",
          "transitions": [
            { "to_state": "ADD_TO_CART", "probability": 0.5 },
            { "to_state": "EXIT", "probability": 0.5 }
          ]
        },
        {
          "state_name": "ADD_TO_CART",
          "transitions": [
            { "to_state": "CHECKOUT_COMPLETE", "probability": 0.8 },
            { "to_state": "ABANDON", "probability": 0.2 }
          ]
        }
      ]
    }
  }
}
```

---

## 📥 4. Automatic SQL DDL Schema Ingestion

If you already have a relational database schema, you don't need to write JSON configurations:
1. Open **Mimic Studio** (`http://localhost:3000`).
2. Click **"Import DDL"** in the top navigation bar.
3. Paste your table definition:
   ```sql
   CREATE TABLE telemetry (
     device_id UUID PRIMARY KEY,
     sensor_type VARCHAR(50),
     temperature_celsius DECIMAL(5,2),
     voltage DECIMAL(4,2),
     battery_level INT,
     status VARCHAR(20),
     recorded_at TIMESTAMP
   );
   ```
4. OpenMimic will automatically configure:
   - `device_id` as UUID
   - `temperature_celsius` & `voltage` as normal Gaussian distributions
   - `status` as weighted choices
   - `recorded_at` as logical timestamps

---

## ☕ Support Open-Source

OpenMimic is 100% free software. If this helps your data engineering workflows, consider supporting the project:
👉 [Buy Me a Coffee](https://buymeacoffee.com/techwasti) | [GitHub Sponsors](https://github.com/sponsors/Techwasti)

