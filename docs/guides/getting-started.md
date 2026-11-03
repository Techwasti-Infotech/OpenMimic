# Local Setup & Quickstart Guide

This guide will walk you through setting up and running the complete OpenMimic stack locally on Windows, macOS, or Linux.

---

## 📋 Prerequisites

Depending on your preferred setup method:
- **Docker & Docker Compose** (recommended for all-in-one local setup), OR:
- **Go 1.23+**
- **Python 3.11+**
- **Node.js 18+** & npm

---

## 🐳 Option 1: Fast Start with Docker Compose (Recommended)

OpenMimic provides a preconfigured multi-container stack with the Go Engine, Python Copilot, and an embedded Redpanda/Kafka broker.

1. **Clone the repository:**
   ```bash
   git clone https://github.com/Techwasti/OpenMimic.git
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

## 💻 Option 2: Running Directly from Source (Local Development)

### 1. Build and Run the OpenMimic Go Engine
```powershell
cd engine
# Run unit tests to verify
go test -v -count=1 ./...

# Build the binary
go build -o bin\mimic.exe .\cmd\mimic

# Start the engine with gRPC and HTTP bridges
.\bin\mimic.exe -grpc-port 50051 -http-port 8080
```
- gRPC server will listen on `localhost:50051`.
- REST / HTTP bridge for Studio will listen on `localhost:8080`.

---

### 2. Start the Mimic Copilot Service (Python FastAPI)
```powershell
# In the repository root
# If using the provided python venv:
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

### 4. Optional: Start the Product Documentation & Landing Website
```powershell
cd website
npx serve -p 3333 .
```
- Landing page will be available at **http://localhost:3333**.

---

## 🧪 Testing Your First Pipeline

1. In **Mimic Studio** (`http://localhost:3000`), navigate to **Generator Designer**.
2. Click **Preview (10)** to generate the first 10 events directly from the engine in real time.
3. Switch to the **Mimic Copilot** tab and submit:
   ```text
   Generate a stream of 500 transaction events with a 2% failure rate and normal distribution amount
   ```
4. View the validated configuration auto-populated in the Designer and start streaming!

---

## ☕ Support Open-Source

OpenMimic is 100% free software. If this helps your data engineering workflows, consider supporting the project:
👉 [Buy Me a Coffee](https://buymeacoffee.com/techwasti) | [GitHub Sponsors](https://github.com/sponsors/Techwasti)
