@echo off
chcp 65001 >nul
setlocal

if "%~1"=="" (
    echo ❌ Ошибка: Укажи версию для релиза!
    echo Пример: release.bat v1.0.1
    exit /b 1
)

set VERSION=%~1

echo 🚀 [1/4] Сборка портативной версии (apm.exe)...
go build -o apm.exe .
if %ERRORLEVEL% NEQ 0 (
    echo ❌ Ошибка сборки apm.exe
    exit /b 1
)

echo 🚀 [2/4] Сборка установщика (apm-installer.exe)...
go build -ldflags "-H=windowsgui" -o apm-installer.exe ./cmd/bootstrap
if %ERRORLEVEL% NEQ 0 (
    echo ❌ Ошибка сборки apm-installer.exe
    exit /b 1
)

echo 📡 Проверка наличия GitHub CLI (gh)...
where gh >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo ❌ Ошибка: Утилита 'gh' ^(GitHub CLI^) не найдена!
    echo Скачай ее здесь: https://cli.github.com/
    echo После установки выполни в консоли: gh auth login
    exit /b 1
)

echo 📦 [3/4] Сохранение и отправка кода на GitHub...
git add .
git commit -m "chore: release %VERSION%"
git push origin main

echo 🏷️ [4/4] Создание релиза %VERSION% и загрузка файлов на GitHub...
gh release create %VERSION% apm.exe apm-installer.exe --title "APM %VERSION%" --notes "Новая версия Awesome Package Manager %VERSION%"

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✅ УРА! Релиз %VERSION% успешно создан! 
    echo 🔗 Файлы apm.exe и apm-installer.exe доступны на GitHub.
) else (
    echo.
    echo ❌ Ошибка при создании релиза на GitHub.
)
