$ErrorActionPreference = 'Stop'
Set-Location (Split-Path $PSScriptRoot -Parent)
if (-not (Get-Command wails -ErrorAction SilentlyContinue)) {
    Write-Host 'Installing Wails v2.14.0...' -ForegroundColor Cyan
    go install github.com/wailsapp/wails/v2/cmd/wails@v2.14.0
}
wails doctor
wails dev
