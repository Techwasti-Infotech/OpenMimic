# OpenMimic Documentation

Welcome to **OpenMimic**, the 100% free and open-source, visual-first synthetic data and production traffic generation platform.

OpenMimic allows engineers, architects, and data teams to generate realistic, mathematically sound data streams directly into Kafka, PostgreSQL, S3, or stdout without relying on static mock files or fragile bash scripts.

---

## ☕ Support Open-Source & Buy Us a Coffee

OpenMimic is maintained as an open-source community effort under the **Apache 2.0 license**.

If OpenMimic helps your engineering workflows:
- ☕ **[Buy Me a Coffee](https://buymeacoffee.com/techwasti)**
- 💖 **[Sponsor on GitHub](https://github.com/sponsors/Techwasti)**
- ⭐ **[Star on GitHub](https://github.com/Techwasti/OpenMimic)**

---

## ⚡ Quick Navigation

- [Local Setup & Quickstart Guide](guides/getting-started.html) — Launch with Docker Compose or from source.
- [Enterprise Single Sign-On (SSO) & RBAC Guide](guides/enterprise-sso.html) — Okta, Microsoft Entra ID, Keycloak, OIDC, and group role mappings.
- [Traffic Simulation & Load Generation Guide](guides/traffic-simulation.html) — Production traffic shaping, diurnal curves, Poisson arrivals, and Markov funnels.
- [Functions & Expression Reference](reference/functions.html) — Complete guide to `uuid`, `normalDistribution`, `weightedChoice`, `markovChain`, etc.
- [Configuration Reference (`mimic.json`)](reference/configuration.html) — Full Protocol Buffers schema guide.
- [Mimic Studio Manual](guides/mimic-studio.html) — Designing schemas, live previews, and observability dashboards.
- [Mimic Copilot Manual](guides/mimic-copilot.html) — Natural-language data synthesis.
- [Architecture & Design Concepts](architecture/) — Engine internals, logical clock, and sinks.

---

## 🚀 Why OpenMimic?

| Feature | OpenMimic | Commercial Data Generators | Static Mock Scripts |
| :--- | :--- | :--- | :--- |
| **License** | **100% Free Apache 2.0** | Proprietary / Closed (\$15k+/yr) | In-house |
| **Visual Studio UI** | ✓ Built-in React 18 UI | ✕ None (JSON Only) | ✕ None |
| **AI Copilot** | ✓ Natural Language to Proto3 | ✕ None | ✕ None |
| **Observability** | ✓ Live Throughput, p99, Errors | Limited CLI metrics | Manual logging |
| **Traffic Shaping** | ✓ Diurnal, Poisson, Flash Bursts | Basic rate limit | Fixed sleeps |
| **Relational Entity Pools**| ✓ Shared Pools & Markov Funnels | State tracking | Complex custom code |
| **SQL DDL Import** | ✓ One-Click `CREATE TABLE` parser | ✕ None | ✕ None |
| **Universal Sinks** | ✓ Kafka, Postgres, S3, Stdout | Kafka, Postgres, S3 | Single-target |

---

## 📦 Monorepo Architecture

OpenMimic is structured as an integrated monorepo:

- **`proto/`**: Single source of truth Protocol Buffers (`mimic_config.proto`).
- **`engine/`**: High-performance Go generation engine with gRPC and HTTP bridges.
- **`studio/`**: Visual UI studio written with React 18, Vite, and Mantine.
- **`copilot/`**: Python FastAPI + LangChain microservice for natural-language configuration synthesis.
