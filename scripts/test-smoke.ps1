@echo off
setlocal
cd /d "%~dp0.."

echo === Smoke + unit tests (campanhas / obras-plus) ===
go test ./server/repositories/... ./server/services/... ./server/templates/core/... ./server/... -count=1 -cover
if errorlevel 1 exit /b 1

echo.
echo === OK ===
