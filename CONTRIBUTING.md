# Contributing to OpenMimic

Thank you for your interest in contributing to **OpenMimic**! We are building the leading open-source, visual-first synthetic data and production traffic generation platform, and we welcome contributions from developers, data engineers, and documentation writers of all skill levels.

---

## 📜 Code of Conduct

We are committed to providing a welcoming, inclusive, and harassment-free community for everyone. Be respectful, constructive, and open to feedback.

---

## 🚀 How to Get Started

### 1. Fork & Clone the Repository
```bash
git clone https://github.com/<your-username>/OpenMimic.git
cd OpenMimic
```

### 2. Run the Stack Locally
You can use the one-click launcher:
- **Windows (PowerShell):** `.\run-local.ps1`
- **Windows (CMD):** `.\run-local.bat`
- **Linux/macOS:** `./run-local.sh`
- **Docker Compose:** `docker compose up -d`

---

## 🏗️ Repository Architecture

OpenMimic is an integrated monorepo:

- **`proto/`**: Single source of truth Protocol Buffers (`mimic_config.proto`). Any new generator function or sink configuration **must be defined here first**.
- **`engine/`**: Core streaming generation engine written in Go 1.23+.
  - `engine/internal/generator/`: Function generators (`uuid`, `normalDistribution`, `markovChain`, `entityPool`).
  - `engine/internal/sink/`: Sinks (`kafka`, `postgres`, `stdout`, `s3`).
  - `engine/internal/traffic/`: Traffic shaping models (`diurnal`, `poisson`, `burst`).
  - `engine/internal/importer/`: SQL DDL and OpenAPI schema parsers.
- **`studio/`**: Frontend Visual Studio built with React 18, Vite, TypeScript, and Mantine UI.
- **`copilot/`**: AI translation service built with Python 3.11+, FastAPI, and LangChain.
- **`website/`**: Product landing website.
- **`docs/`**: Jekyll documentation deployed to GitHub Pages.

---

## 🧪 Testing Guidelines

Before opening a pull request, ensure all tests pass:

### 1. Test the Go Engine
```bash
cd engine
go test -v -count=1 ./...
```

### 2. Test the Copilot Service
```bash
python -m unittest copilot/test_copilot.py
```

### 3. Build Mimic Studio
```bash
cd studio
npm run build
```

---

## 💡 Types of Contributions Welcome

1. **New Sinks:**
   - Apache Iceberg, Delta Lake, Snowflake, ClickHouse, DuckDB, Redis streams (`XADD`), AWS SQS, or GCP PubSub.
2. **New Generator Functions:**
   - Geospatial GPS coordinates (bounding boxes, routes).
   - Domain-specific fakers (financial credit card numbers, IBAN, medical ICD-10 codes).
   - Chaos & corrupted records (null injection, malformed JSON).
3. **Studio UI Improvements:**
   - Drag-and-drop node graph canvas.
   - Dark/Light mode theme toggle.
4. **Documentation & Tutorials:**
   - Guides for connecting OpenMimic to Apache Flink, Spark Streaming, or Kafka Connect.

---

## 📬 Pull Request Workflow

1. Create a feature branch:
   ```bash
   git checkout -b feature/my-new-sink
   ```
2. Commit your changes with clear, descriptive commit messages.
3. Push to your fork:
   ```bash
   git push origin feature/my-new-sink
   ```
4. Open a Pull Request on the main repository describing your changes, motivation, and tests performed.

---

## ☕ Support the Project

If you love OpenMimic, consider buying us a coffee or becoming a sponsor:
- ☕ **Buy Me a Coffee:** [buymeacoffee.com/techwasti](https://buymeacoffee.com/techwasti)
- 💖 **GitHub Sponsors:** [github.com/sponsors/Techwasti](https://github.com/sponsors/Techwasti)

