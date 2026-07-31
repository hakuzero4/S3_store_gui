$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot
Write-Host "==> Building frontend..." -ForegroundColor Cyan
Push-Location web
if (-not (Test-Path node_modules)) { npm install }
npm run build
if ($LASTEXITCODE -ne 0) { throw "frontend build failed" }
Pop-Location
Write-Host "==> Copying dist into embed package..." -ForegroundColor Cyan
$distTarget = "internal/static/dist"
if (Test-Path $distTarget) { Remove-Item -Recurse -Force $distTarget }
Copy-Item -Recurse "web/dist" $distTarget
Write-Host "==> Building Go exe..." -ForegroundColor Cyan
New-Item -ItemType Directory -Force -Path dist | Out-Null
$env:CGO_ENABLED = "0"
go build -trimpath -ldflags "-s -w -X main.version=0.1.0" -o dist/s3store.exe ./cmd/s3store
if ($LASTEXITCODE -ne 0) { throw "go build failed" }
Write-Host "Done: dist/s3store.exe" -ForegroundColor Green
Get-Item dist/s3store.exe | Format-List Name, Length, LastWriteTime
