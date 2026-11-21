# OpenMimic Documentation

Welcome to **OpenMimic**, the 100% free and open-source, visual-first synthetic data and production traffic generation platform.

OpenMimic allows engineers, architects, and data teams to generate realistic, mathematically sound data streams directly into Kafka, PostgreSQL, S3, or stdout without relying on static mock files or fragile bash scripts.

---

## ☕ Support Open-Source & Buy Us a Coffee

OpenMimic is maintained as an open-source community effort under the **Apache 2.0 license**.

If OpenMimic helps your engineering workflows:
- ☕ **[Buy Me a Coffee](https://buymeacoffee.com/techwasti)**
- 💖 **[Sponsor on GitHub](https://github.com/sponsors/Techwasti)**
- ⭐ **[Star on GitHub](https://github.com/Techwasti-Infotech/OpenMimic)**

---

## ⚡ Quick Navigation

- [Local Setup & Quickstart Guide]({{ site.baseurl }}/guides/getting-started.html) — Launch with Docker Compose or from source.
- [Enterprise Single Sign-On (SSO) & RBAC Guide]({{ site.baseurl }}/guides/enterprise-sso.html) — Okta, Microsoft Entra ID, Keycloak, OIDC, and group role mappings.
- [Model Context Protocol (MCP) Server Guide]({{ site.baseurl }}/guides/mcp-server.html) — Autonomous LLM agent tool integration (Claude, Cursor, Docker).
- [Traffic Simulation & Load Generation Guide]({{ site.baseurl }}/guides/traffic-simulation.html) — Production traffic shaping, diurnal curves, Poisson arrivals, and Markov funnels.
- [Functions & Expression Reference]({{ site.baseurl }}/reference/functions.html) — Complete guide to `uuid`, `normalDistribution`, `weightedChoice`, `markovChain`, etc.
- [Configuration Reference (`mimic.json`)]({{ site.baseurl }}/reference/configuration.html) — Full Protocol Buffers schema guide.
- [Mimic Studio Manual]({{ site.baseurl }}/guides/mimic-studio.html) — Designing schemas, live previews, and observability dashboards.
- [Mimic Copilot Manual]({{ site.baseurl }}/guides/mimic-copilot.html) — Natural-language data synthesis.
- [Architecture & Design Concepts]({{ site.baseurl }}/architecture/) — Engine internals, logical clock, and sinks.

---

## 🚀 Why OpenMimic?

| Feature | OpenMimic | Commercial Data Generators | Static Mock Scripts |
| :--- | :--- | :--- | :--- |
| **License** | **100% Free Apache 2.0** | Proprietary / Closed ($15k+/yr) | In-house |
| **Visual Studio UI** | ✓ Built-in React 18 UI | ✕ None (JSON Only) | ✕ None |
| **AI Copilot** | ✓ Natural Language to Proto3 | ✕ None | ✕ None |
| **Observability** | ✓ Live Throughput, p99, Errors | Limited CLI metrics | Manual logging |
| **Traffic Shaping** | ✓ Diurnal, Poisson, Flash Bursts | Basic rate limit | Fixed sleeps |
| **Relational Entity Pools** | ✓ Shared Pools & Markov Funnels | State tracking | Complex custom code |
| **SQL DDL Import** | ✓ One-Click `CREATE TABLE` parser | ✕ None | ✕ None |
| **Universal Sinks** | ✓ Kafka, Postgres, S3, Iceberg, Delta | Kafka, Postgres, S3 | Single-target |
