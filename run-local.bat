@echo off
title OpenMimic One-Click Launcher
echo =========================================
echo   OpenMimic One-Click Local Launcher
echo =========================================

REM 1. Start OpenMimic Engine
echo [1/3] Starting OpenMimic Go Engine (:50051 & :8080)...
start "OpenMimic Engine" /min cmd /c "cd engine && go run .\cmd\mimic -grpc-port 50051 -http-port 8080"

REM 2. Start Mimic Copilot Service
echo [2/3] Starting Mimic Copilot Service (:8000)...
if exist ".venv\Scripts\python.exe" (
    start "Mimic Copilot" /min cmd /c ".\.venv\Scripts\python.exe -m uvicorn copilot.main:app --host 0.0.0.0 --port 8000"
) else (
    start "Mimic Copilot" /min cmd /c "python -m uvicorn copilot.main:app --host 0.0.0.0 --port 8000"
)

REM 3. Start Mimic Studio UI
echo [3/3] Starting Mimic Studio UI (:3000)...
start "Mimic Studio" /min cmd /c "cd studio && npm run dev"

timeout /t 3 /nobreak >nul

echo =========================================
echo   OpenMimic Stack is now Running!
echo =========================================
echo   Mimic Studio UI:    http://localhost:3000
echo   Engine HTTP Bridge: http://localhost:8080
echo   Engine gRPC:        localhost:50051
echo   Copilot API:        http://localhost:8000
echo =========================================

start http://localhost:3000

