---
layout: default
title: "Local Setup & Quickstart Guide"
---

# Local Setup & Quickstart Guide

This guide will walk you through setting up and running the complete OpenMimic stack locally on Windows, macOS, or Linux.

---

## âš¡ Option 1: One-Click Local Launcher (Fastest)

OpenMimic provides automated one-click launcher scripts that check your dependencies, compile the Go binaries, and launch **Mimic Studio**, **Engine**, and **Copilot** simultaneously.

### On Windows:
```powershell
# In PowerShell:
.\run-local.ps1

# Or in Command Prompt (CMD):
run-local.bat
```

### On macOS / Linux:
```bash
chmod +x run-local.sh
./run-local.sh
```

Once launched, the services will be running at:
- ðŸŽ¨ **Mimic Studio (Web UI):** [http://localhost:3000](http://localhost:3000)
- ðŸ”Œ **Engine HTTP/REST Bridge:** [http://localhost:8080](http://localhost:8080)
- âš¡ **Engine gRPC:** `localhost:50051`
- ðŸ¤– **Mimic Copilot Service:** [http://localhost:8000](http://localhost:8000)

---

## ðŸ³ Option 2: Docker Compose Setup

OpenMimic provides a preconfigured multi-container stack with the Go Engine, Python Copilot, and an embedded Redpanda/Kafka broker.

1. **Clone the repository:**
   ```bash
   git clone https://github.com/Techwasti-Infotech/OpenMimic.git
   cd OpenMimic
   ```

2. **Start all services:**
   ```bash
   docker compose up -d
   ```

3. **Verify running containers:**
   ```bash
   docker compose ps
   ```
   - **Kafka / Redpanda Broker:** `localhost:9092`
   - **OpenMimic Engine:** `localhost:8080` (HTTP Bridge) & `localhost:50051` (gRPC)
   - **Mimic Copilot:** `localhost:8000`

4. **Launch Mimic Studio:**
   ```bash
   cd studio
   npm install
   npm run dev
   ```
   Open **http://localhost:3000** in your browser.

---

## ðŸ’» Option 3: Running Components Individually from Source

### 1. Build and Run the OpenMimic Go Engine
```powershell
cd engine
# Run unit tests to verify
go test -v -count=1 ./...

# Build the binaries
go build -o bin\mimic.exe .\cmd\mimic
go build -o bin\mimic-cli.exe .\cmd\mimic-cli

# Start the engine with gRPC and HTTP bridges
.\bin\mimic.exe -grpc-port 50051 -http-port 8080
```
- gRPC server will listen on `localhost:50051`.
- REST / HTTP bridge for Studio will listen on `localhost:8080`.

---

### 2. Start the Mimic Copilot Service (Python FastAPI)
```powershell
# In the repository root
.\.venv\Scripts\activate
uvicorn copilot.main:app --host 0.0.0.0 --port 8000 --reload
```
- Copilot API will be available at `http://localhost:8000`.
- Health check: `curl http://localhost:8000/health` $\rightarrow$ `{"status":"ok"}`.

---

### 3. Start Mimic Studio (React 18 + Mantine)
```powershell
cd studio
npm install
npm run dev
```
- Studio UI will be accessible at **http://localhost:3000**.

---

## ðŸ› ï¸ Using the GitOps CLI (`mimic-cli`)

For CI/CD pipelines and headless terminal execution:

- **Validate Config & Constraints:**
  ```powershell
  .\engine\bin\mimic-cli.exe validate --config sample_order_flow.json --constraints
  ```
- **Stream Event Preview to Terminal:**
  ```powershell
  .\engine\bin\mimic-cli.exe preview --config sample_order_flow.json --count 5
  ```
- **Dry-Run Event Generation (Null Sink):**
  ```powershell
  .\engine\bin\mimic-cli.exe run --config sample_order_flow.json --dry-run
  ```
- **Export Config as Template:**
  ```powershell
  .\engine\bin\mimic-cli.exe export --config sample_order_flow.json --to-template
  ```
- **Auto-Discover from SQL DDL:**
  ```powershell
  .\engine\bin\mimic-cli.exe discover --sql "CREATE TABLE users (id UUID, email VARCHAR, amount DECIMAL);"
  ```

---

## ðŸ§ª Testing Your First Pipeline

1. In **Mimic Studio** (`http://localhost:3000`), navigate to **Generator Designer**.
2. Click **Preview (10)** to generate the first 10 events directly from the engine in real time.
3. Switch to the **Mimic Copilot** tab and submit:
   ```text
   Generate a stream of 500 transaction events with a 2% failure rate and normal distribution amount
   ```
4. Explore **Constraints & Solvers**, **Quality & Fidelity**, **Privacy & PII Scanner**, and **Stream Topology & DAG**!

---

## â˜• Support Open-Source

OpenMimic is 100% free software. If this helps your data engineering workflows, consider supporting the project:
ðŸ‘‰ [Buy Me a Coffee](https://buymeacoffee.com/techwasti) | [GitHub Sponsors](https://github.com/sponsors/Techwasti)
