@echo off
setlocal
cd /d "%~dp0"

echo ============================================================
echo Mongo Developer V1.1.1 Test Discovery Hotfix
echo ============================================================
echo.

if exist "frontend\src\sql.node.test.mjs" (
    echo Removing stale V1.1.0 standalone test filename...
    del /f /q "frontend\src\sql.node.test.mjs"
)

if not exist "frontend\src\sql.node.check.mjs" (
    echo ERROR: frontend\src\sql.node.check.mjs is missing.
    exit /b 1
)

echo Hotfix applied successfully.
echo.
echo Now run:
echo   BUILD-INSTALLER.cmd
exit /b 0
