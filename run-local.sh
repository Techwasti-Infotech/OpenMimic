#!/usr/bin/env bash
set -e

echo "========================================="
echo "  OpenMimic One-Click Local Launcher"
echo "========================================="

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 1. Start OpenMimic Go Engine
echo "[1/3] Starting OpenMimic Go Engine (:50051 & :8080)..."
(cd "$ROOT_DIR/engine" && go run ./cmd/mimic -grpc-port 50051 -http-port 8080) &
ENGINE_PID=$!

# 2. Start Mimic Copilot Service
echo "[2/3] Starting Mimic Copilot Service (:8000)..."
if [ -f "$ROOT_DIR/.venv/bin/python" ]; then
    (cd "$ROOT_DIR" && .venv/bin/python -m uvicorn copilot.main:app --host 0.0.0.0 --port 8000) &
else
    (cd "$ROOT_DIR" && python3 -m uvicorn copilot.main:app --host 0.0.0.0 --port 8000) &
fi
COPILOT_PID=$!

# 3. Start Mimic Studio UI
echo "[3/3] Starting Mimic Studio UI (:3000)..."
(cd "$ROOT_DIR/studio" && npm run dev) &
STUDIO_PID=$!

trap "kill $ENGINE_PID $COPILOT_PID $STUDIO_PID 2>/dev/null" EXIT

sleep 3

echo "========================================="
echo "  OpenMimic Stack is now Running!"
echo "========================================="
echo "  Mimic Studio UI:    http://localhost:3000"
echo "  Engine HTTP Bridge: http://localhost:8080"
echo "  Engine gRPC:        localhost:50051"
echo "  Copilot API:        http://localhost:8000"
echo "========================================="

# Open browser if possible
if which xdg-open > /dev/null; then
    xdg-open "http://localhost:3000"
elif which open > /dev/null; then
    open "http://localhost:3000"
fi

wait

