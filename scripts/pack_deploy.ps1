# Script đóng gói Docker Image để deploy offline
# Yêu cầu: Đã cài Docker Desktop trên Windows

$ErrorActionPreference = "Stop"

# Chuyển về thư mục gốc của project (cha của thư mục scripts này)
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectRoot = Split-Path -Parent $scriptDir
Set-Location $projectRoot

Write-Host ">>> Thu muc lam viec hien tai: $projectRoot" -ForegroundColor Cyan

Write-Host ">>> Bat dau xay dung Docker Images..." -ForegroundColor Green

# 1. Build Backend
Write-Host "Dang build Backend..."
Push-Location "backend"
try {
    docker build -t fusion-backend:latest -f Dockerfile .
} finally {
    Pop-Location
}

# 2. Build Frontend
Write-Host "Dang build Frontend..."
Push-Location "frontend"
try {
    docker build -t fusion-frontend:latest -f Dockerfile .
} finally {
    Pop-Location
}

# 3. Pull Ngrok (để đảm bảo có image mới nhất)
Write-Host "Dang pull Ngrok..."
# Them retry neu mang yeu
for ($i=1; $i -le 3; $i++) {
    try {
        docker pull ngrok/ngrok:latest
        break
    } catch {
        Write-Warning "Pull Ngrok that bai (lan $i/3). Dang thu lai..."
        Start-Sleep -Seconds 5
    }
}

# 4. Save Images to file (Lưu file tại thư mục gốc của project luôn cho dễ tìm)
$outputFile = Join-Path $projectRoot "fusion_images.tar"
Write-Host ">>> Dang luu cac images vao file: $outputFile ..." -ForegroundColor Yellow
docker save -o "$outputFile" fusion-backend:latest fusion-frontend:latest ngrok/ngrok:latest

if (Test-Path "$outputFile") {
    Write-Host ">>> THANH CONG!" -ForegroundColor Green
    Write-Host "File '$outputFile' da duoc tao."
    Write-Host "Kich thuoc: $( (Get-Item $outputFile).Length / 1MB ) MB"
} else {
    Write-Error "LOI: Khong tim thay file '$outputFile' sau khi save!"
}
Write-Host "--------------------------------------------------------"
Write-Host "HUONG DAN TIEP THEO:"
Write-Host "1. Copy 2 file sau len Linux Server:"
Write-Host "   - $outputFile"
Write-Host "   - deployments/docker-compose.prod.yml (Doi ten thanh docker-compose.yml)"
Write-Host "2. Tren Linux, chay lenh:"
Write-Host "   docker load -i fusion_images.tar"
Write-Host "   docker-compose up -d"
Write-Host "--------------------------------------------------------"
