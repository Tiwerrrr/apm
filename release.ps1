param (
    [Parameter(Mandatory=$false)]
    [string]$Version
)

if (-not $Version) {
    Write-Host "[ERROR] Please specify the release version!" -ForegroundColor Red
    Write-Host "Example: .\release.ps1 v1.0.1"
    exit 1
}

$ErrorActionPreference = "Stop"

Write-Host "[1/4] Building portable version (apm.exe)..." -ForegroundColor Cyan
go build -o apm.exe .
if ($LASTEXITCODE -ne 0 -and $LASTEXITCODE -ne $null) {
    Write-Host "[ERROR] Failed to build apm.exe" -ForegroundColor Red
    exit 1
}

Write-Host "[2/4] Building installer (apm-installer.exe)..." -ForegroundColor Cyan
go build -ldflags "-H=windowsgui" -o apm-installer.exe ./cmd/bootstrap
if ($LASTEXITCODE -ne 0 -and $LASTEXITCODE -ne $null) {
    Write-Host "[ERROR] Failed to build apm-installer.exe" -ForegroundColor Red
    exit 1
}

Write-Host "[DEBUG] Checking for GitHub CLI (gh)..." -ForegroundColor Cyan
if (-not (Get-Command "gh" -ErrorAction SilentlyContinue)) {
    Write-Host "[ERROR] GitHub CLI (gh) not found!" -ForegroundColor Red
    Write-Host "Please download it from: https://cli.github.com/"
    Write-Host "After installation, run: gh auth login"
    exit 1
}

Write-Host "[3/4] Committing and pushing code to GitHub..." -ForegroundColor Cyan
git add .
git commit -m "chore: release $Version"
git push origin main

Write-Host "[4/4] Creating GitHub Release $Version and uploading assets..." -ForegroundColor Cyan
gh release create $Version apm.exe apm-installer.exe --title "APM $Version" --notes "Automated release of Awesome Package Manager $Version"

if ($LASTEXITCODE -eq 0 -or $LASTEXITCODE -eq $null) {
    Write-Host ""
    Write-Host "[SUCCESS] YAY! Release $Version created successfully!" -ForegroundColor Green
    Write-Host "Files apm.exe and apm-installer.exe are now available on GitHub Releases."
} else {
    Write-Host ""
    Write-Host "[ERROR] Failed to create GitHub release." -ForegroundColor Red
}
