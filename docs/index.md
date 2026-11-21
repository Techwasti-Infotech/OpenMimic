# OpenMimic Documentation

Welcome to **OpenMimic**, the 100% free and open-source, visual-first synthetic data and production traffic generation platform.

OpenMimic allows engineers, architects, and data teams to generate realistic, mathematically sound data streams directly into Kafka, PostgreSQL, S3, or stdout without relying on static mock files or fragile bash scripts.

---

## â˜• Support Open-Source & Buy Us a Coffee

OpenMimic is maintained as an open-source community effort under the **Apache 2.0 license**.

If OpenMimic helps your engineering workflows:
- â˜• **[Buy Me a Coffee](https://buymeacoffee.com/techwasti)**
- ðŸ’– **[Sponsor on GitHub](https://github.com/sponsors/Techwasti)**
- â­ **[Star on GitHub](https://github.com/Techwasti/OpenMimic)**

---

## âš¡ Quick Navigation

- [Local Setup & Quickstart Guide](guides/getting-started.html) â€” Launch with Docker Compose or from source.
- [Enterprise Single Sign-On (SSO) & RBAC Guide](guides/enterprise-sso.html) â€” Okta, Microsoft Entra ID, Keycloak, OIDC, and group role mappings.
- [Model Context Protocol (MCP) Server](guides/mcp-server.html)
- [Traffic Simulation & Load Generation Guide](guides/traffic-simulation.html) â€” Production traffic shaping, diurnal curves, Poisson arrivals, and Markov funnels.
- [Functions & Expression Reference](reference/functions.html) â€” Complete guide to `uuid`, `normalDistribution`, `weightedChoice`, `markovChain`, etc.
- [Configuration Reference (`mimic.json`)](reference/configuration.html) â€” Full Protocol Buffers schema guide.
- [Mimic Studio Manual](guides/mimic-studio.html) â€” Designing schemas, live previews, and observability dashboards.
- [Mimic Copilot Manual](guides/mimic-copilot.html) â€” Natural-language data synthesis.
- [Architecture & Design Concepts](architecture/) â€” Engine internals, logical clock, and sinks.

---

## ðŸš€ Why OpenMimic?

| Feature | OpenMimic | Commercial Data Generators | Static Mock Scripts |
| :--- | :--- | :--- | :--- |
| **License** | **100% Free Apache 2.0** | Proprietary / Closed (\$15k+/yr) | In-house |
| **Visual Studio UI** | âœ“ Built-in React 18 UI | âœ• None (JSON Only) | âœ• None |
| **AI Copilot** | âœ“ Natural Language to Proto3 | âœ• None | âœ• None |
| **Observability** | âœ“ Live Throughput, p99, Errors | Limited CLI metrics | Manual logging |
| **Traffic Shaping** | âœ“ Diurnal, Poisson, Flash Bursts | Basic rate limit | Fixed sleeps |
| **Relational Entity Pools**| âœ“ Shared Pools & Markov Funnels | State tracking | Complex custom code |
| **SQL DDL Import** | âœ“ One-Click `CREATE TABLE` parser | âœ• None | âœ• None |
| **Universal Sinks** | âœ“ Kafka, Postgres, S3, Stdout | Kafka, Postgres, S3 | Single-target |

---

## ðŸ“¦ Monorepo Architecture

OpenMimic is structured as an integrated monorepo:

- **`proto/`**: Single source of truth Protocol Buffers (`mimic_config.proto`).
- **`engine/`**: High-performance Go generation engine with gRPC and HTTP bridges.
- **`studio/`**: Visual UI studio written with React 18, Vite, and Mantine.
- **`copilot/`**: Python FastAPI + LangChain microservice for natural-language configuration synthesis.
