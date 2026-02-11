# Script to run both Backend and Frontend
$ErrorActionPreference = "Stop"

Write-Host "[READY] Starting Fusion-Shundao System..." -ForegroundColor Green

# 0. Kill existing processes to avoid conflicts (Force Kill)
Write-Host "[CLEANUP] Stopping old processes..." -ForegroundColor Yellow
Stop-Process -Name "ngrok" -Force -ErrorAction SilentlyContinue
Stop-Process -Name "fusion_test" -Force -ErrorAction SilentlyContinue

# 1. Start Backend (Go)
Write-Host "[PROCESS] Building Backend..." -ForegroundColor Cyan
Push-Location backend
go build -o ../fusion_test.exe ./cmd/server
if ($LASTEXITCODE -ne 0) { Pop-Location; Write-Error "Build Failed!"; exit 1 }
Pop-Location

Write-Host "[PROCESS] Launching Backend (Go)..." -ForegroundColor Cyan
Start-Process powershell -ArgumentList "-NoExit", "-Command", "& {Write-Host '=== BACKEND (GO) ===' -ForegroundColor Yellow; ./fusion_test.exe}"

# 2. Wait for Backend to initialize (optional, but good for cleanliness)
Start-Sleep -Seconds 2

# 3. Start Frontend (React/Vite)
Write-Host "[PROCESS] Launching Frontend (React)..." -ForegroundColor Cyan
Start-Process powershell -ArgumentList "-NoExit", "-Command", "& {Write-Host '=== FRONTEND (REACT) ===' -ForegroundColor Cyan; cd frontend; npm run dev}"

Write-Host "[PROCESS] Launching Ngrok Tunnel (Frontend)..." -ForegroundColor Cyan
Start-Process powershell -ArgumentList "-NoExit", "-Command", "& {ngrok http --domain=prescholastic-hurtlingly-latia.ngrok-free.dev 9000}"

Write-Host "[SUCCESS] System started! Check the THREE new PowerShell windows." -ForegroundColor Green
Write-Host "[HOST]   - Backend: http://localhost:5039" 
Write-Host "[HOST]   - Frontend: http://localhost:9000"
Write-Host "[PUBLIC] - Ngrok: https://prescholastic-hurtlingly-latia.ngrok-free.dev"
Write-Host "===================================[FINISHED]==================================="
