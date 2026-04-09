$ErrorActionPreference = "Stop"

$ROOT = $PSScriptRoot
$BACKEND_DIR  = Join-Path $ROOT "backend"
$FRONTEND_DIR = Join-Path $ROOT "frontend"
$DEPLOY_DIR   = Join-Path $ROOT "deployments"
$OUTPUT_TAR   = Join-Path $ROOT "shundao_production.tar"

$BACKEND_IMAGE  = "fusion-backend:latest"
$FRONTEND_IMAGE = "fusion-frontend:latest"

Write-Host "============= SHUNDAO SOLAR - Docker Production Build =============="
Write-Host ""
Write-Host "[1/4] Building Backend Image ($BACKEND_IMAGE)..."
docker build -t $BACKEND_IMAGE $BACKEND_DIR
if ($LASTEXITCODE -ne 0) { Write-Error "Backend build FAILED!"; exit 1 }
Write-Host "      -> Backend Image built OK"

Write-Host "[2/4] Building Frontend Image ($FRONTEND_IMAGE)..."
docker build -t $FRONTEND_IMAGE $FRONTEND_DIR
if ($LASTEXITCODE -ne 0) { Write-Error "Frontend build FAILED!"; exit 1 }
Write-Host "      -> Frontend Image built OK"

Write-Host "[3/4] Saving Images to $OUTPUT_TAR ..."
docker save $BACKEND_IMAGE $FRONTEND_IMAGE -o $OUTPUT_TAR
if ($LASTEXITCODE -ne 0) { Write-Error "docker save FAILED!"; exit 1 }
Write-Host "      -> Exported OK"

Write-Host "[4/4] Preparing deployment package..."
$PKG_DIR = Join-Path $ROOT "shundao_deploy_package"
New-Item -ItemType Directory -Path $PKG_DIR -Force | Out-Null

Copy-Item $OUTPUT_TAR $PKG_DIR -Force
Copy-Item (Join-Path $ROOT ".env.example") $PKG_DIR -Force
Copy-Item (Join-Path $DEPLOY_DIR "docker-compose.prod.yml") $PKG_DIR -Force
$configsDst = Join-Path $PKG_DIR "configs"
if (Test-Path (Join-Path $ROOT "configs")) {
    Copy-Item -Recurse -Force (Join-Path $ROOT "configs") $configsDst
}

Write-Host ""
Write-Host "BUILD COMPLETED! Directory: shundao_deploy_package"
Write-Host ""
Write-Host "INSTALLATION INSTRUCTIONS (Linux Server):"
Write-Host " 1. Copy shundao_deploy_package to Server"
Write-Host " 2. SSH to Server"
Write-Host " 3. docker load -i shundao_production.tar"
Write-Host " 4. cp .env.example .env"
Write-Host " 5. docker compose -f docker-compose.prod.yml --env-file .env up -d"
