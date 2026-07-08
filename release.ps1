param (
    [Parameter(Mandatory=$false)]
    [string]$Version
)

if (-not $Version) {
    Write-Host "[ERROR] Укажи версию для релиза!" -ForegroundColor Red
    Write-Host "Пример: .\release.ps1 v1.0.1"
    exit 1
}

$ErrorActionPreference = "Stop"

Write-Host "[1/4] Сборка портативной версии (apm.exe)..." -ForegroundColor Cyan
go build -o apm.exe .
if ($LASTEXITCODE -ne 0 -and $LASTEXITCODE -ne $null) {
    Write-Host "[ERROR] Ошибка сборки apm.exe" -ForegroundColor Red
    exit 1
}

Write-Host "[2/4] Сборка установщика (apm-installer.exe)..." -ForegroundColor Cyan
go build -ldflags "-H=windowsgui" -o apm-installer.exe ./cmd/bootstrap
if ($LASTEXITCODE -ne 0 -and $LASTEXITCODE -ne $null) {
    Write-Host "[ERROR] Ошибка сборки apm-installer.exe" -ForegroundColor Red
    exit 1
}

Write-Host "[DEBUG] Проверка наличия GitHub CLI (gh)..." -ForegroundColor Cyan
if (-not (Get-Command "gh" -ErrorAction SilentlyContinue)) {
    Write-Host "[ERROR] Утилита 'gh' (GitHub CLI) не найдена!" -ForegroundColor Red
    Write-Host "Скачай ее здесь: https://cli.github.com/"
    Write-Host "После установки выполни в консоли: gh auth login"
    exit 1
}

Write-Host "[3/4] Сохранение и отправка кода на GitHub..." -ForegroundColor Cyan
git add .
git commit -m "chore: release $Version"
git push origin main

Write-Host "[4/4] Создание релиза $Version и загрузка файлов на GitHub..." -ForegroundColor Cyan
gh release create $Version apm.exe apm-installer.exe --title "APM $Version" --notes "Новая версия Awesome Package Manager $Version"

if ($LASTEXITCODE -eq 0 -or $LASTEXITCODE -eq $null) {
    Write-Host ""
    Write-Host "[SUCCESS] УРА! Релиз $Version успешно создан!" -ForegroundColor Green
    Write-Host "Файлы apm.exe и apm-installer.exe доступны на GitHub."
} else {
    Write-Host ""
    Write-Host "[ERROR] Ошибка при создании релиза на GitHub." -ForegroundColor Red
}
