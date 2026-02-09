# Script to run both Backend and Frontend
$ErrorActionPreference = "Stop"

Write-Host "🚀 Starting Fusion-Shundao System..." -ForegroundColor Green

# 1. Start Backend (Go)
Write-Host "   + Building Backend..." -ForegroundColor Cyan
go build -o fusion_test.exe ./src
if ($LASTEXITCODE -ne 0) { Write-Error "Build Failed!"; exit 1 }

Write-Host "   + Launching Backend (Go)..." -ForegroundColor Cyan
Start-Process powershell -ArgumentList "-NoExit", "-Command", "& {Write-Host '=== BACKEND (GO) ===' -ForegroundColor Yellow; ./fusion_test.exe}"

# 2. Wait for Backend to initialize (optional, but good for cleanliness)
Start-Sleep -Seconds 2

# 3. Start Frontend (React/Vite)
Write-Host "   + Launching Frontend (React)..." -ForegroundColor Cyan
Start-Process powershell -ArgumentList "-NoExit", "-Command", "& {Write-Host '=== FRONTEND (REACT) ===' -ForegroundColor Cyan; cd UI; npm run dev}"

Write-Host "✅ System started! Check the two new PowerShell windows." -ForegroundColor Green
Write-Host "   - Backend: http://localhost:5039" 
Write-Host "   - Frontend: http://localhost:5173"
