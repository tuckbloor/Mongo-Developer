@echo off
setlocal
cd /d "%~dp0"

set "VERSION=%~1"
if "%VERSION%"=="" set "VERSION=1.5.0"

echo ============================================================
echo Mongo Developer Windows Installer Build
echo Version %VERSION%
echo ============================================================
echo.

powershell.exe -NoLogo -NoProfile -ExecutionPolicy Bypass -File "%~dp0scripts\build-installer.ps1" -Version "%VERSION%"
set "RC=%ERRORLEVEL%"

echo.
if not "%RC%"=="0" (
  echo BUILD FAILED with exit code %RC%.
  exit /b %RC%
)

echo BUILD SUCCEEDED.
exit /b 0
