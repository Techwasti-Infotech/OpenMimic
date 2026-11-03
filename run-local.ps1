Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "  OpenMimic One-Click Local Launcher" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan

$Root = Split-Path -Parent $MyInvocation.MyCommand.Path

# 1. Start OpenMimic Go Engine
Write-Host "[1/3] Building & Starting OpenMimic Go Engine (:50051 & :8080)..." -ForegroundColor Yellow
Start-Process -FilePath "powershell.exe" -ArgumentList "-NoExit", "-Command", "cd '$Root\engine'; go run .\cmd\mimic -grpc-port 50051 -http-port 8080" -WindowStyle Minimized

# 2. Start Mimic Copilot Service
Write-Host "[2/3] Starting Mimic Copilot Service (:8000)..." -ForegroundColor Yellow
$PythonPath = "$Root\.venv\Scripts\python.exe"
if (Test-Path $PythonPath) {
    Start-Process -FilePath "powershell.exe" -ArgumentList "-NoExit", "-Command", "cd '$Root'; & '$PythonPath' -m uvicorn copilot.main:app --host 0.0.0.0 --port 8000" -WindowStyle Minimized
} else {
    Start-Process -FilePath "powershell.exe" -ArgumentList "-NoExit", "-Command", "cd '$Root'; python -m uvicorn copilot.main:app --host 0.0.0.0 --port 8000" -WindowStyle Minimized
}

# 3. Start Mimic Studio UI
Write-Host "[3/3] Starting Mimic Studio UI (:3000)..." -ForegroundColor Yellow
$NpmCmd = (Get-Command npm.cmd -ErrorAction SilentlyContinue).Source
if (-not $NpmCmd -and (Test-Path "C:\nvm4w\nodejs\npm.cmd")) {
    $NpmCmd = "C:\nvm4w\nodejs\npm.cmd"
}
if ($NpmCmd) {
    Start-Process -FilePath "powershell.exe" -ArgumentList "-NoExit", "-Command", "cd '$Root\studio'; & '$NpmCmd' run dev" -WindowStyle Minimized
} else {
    Start-Process -FilePath "powershell.exe" -ArgumentList "-NoExit", "-Command", "cd '$Root\studio'; npm run dev" -WindowStyle Minimized
}

Start-Sleep -Seconds 3

Write-Host ""
Write-Host "=========================================" -ForegroundColor Green
Write-Host "  OpenMimic Stack is now Running!" -ForegroundColor Green
Write-Host "=========================================" -ForegroundColor Green
Write-Host "  Mimic Studio UI:    http://localhost:3000" -ForegroundColor Cyan
Write-Host "  Engine HTTP Bridge: http://localhost:8080" -ForegroundColor Cyan
Write-Host "  Engine gRPC:        localhost:50051" -ForegroundColor Cyan
Write-Host "  Copilot API:        http://localhost:8000" -ForegroundColor Cyan
Write-Host ""
Write-Host "Opening Mimic Studio in your browser..." -ForegroundColor Gray
Start-Process "http://localhost:3000"

