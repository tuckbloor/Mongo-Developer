param(
    [ValidatePattern('^\d+\.\d+\.\d+([-.][0-9A-Za-z.-]+)?$')]
    [string]$Version = '1.5.0'
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$root = Split-Path $PSScriptRoot -Parent
Set-Location $root

function Invoke-Native {
    param(
        [Parameter(Mandatory = $true)][string]$FilePath,
        [Parameter(ValueFromRemainingArguments = $true)][string[]]$Arguments
    )

    & $FilePath @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "$FilePath exited with code $LASTEXITCODE"
    }
}

function Require-Command {
    param(
        [Parameter(Mandatory = $true)][string]$Name,
        [Parameter(Mandatory = $true)][string]$InstallHint
    )
    $cmd = Get-Command $Name -ErrorAction SilentlyContinue
    if (-not $cmd) {
        throw "$Name is required. $InstallHint"
    }
    return $cmd.Source
}

& "$PSScriptRoot\bootstrap-node.ps1"

$goPath = Require-Command 'go.exe' 'Install a current Go release from go.dev.'

$wailsCommand = Get-Command 'wails.exe' -ErrorAction SilentlyContinue
if ($wailsCommand) {
    $wailsPath = $wailsCommand.Source
}
else {
    Write-Host 'Wails CLI not found; installing Wails v2.14.0...' -ForegroundColor Yellow
    Invoke-Native $goPath install github.com/wailsapp/wails/v2/cmd/wails@v2.14.0

    $goRoot = (& $goPath env GOPATH).Trim()
    if ($LASTEXITCODE -ne 0) {
        throw "go env GOPATH exited with code $LASTEXITCODE"
    }

    $wailsPath = Join-Path $goRoot 'bin\wails.exe'
    if (-not (Test-Path $wailsPath)) {
        throw "Wails installation completed but wails.exe was not found at $wailsPath"
    }
}

$makensisCommand = Get-Command 'makensis.exe' -ErrorAction SilentlyContinue
if ($makensisCommand) {
    $makensisPath = $makensisCommand.Source
}
elseif (Test-Path 'C:\Program Files (x86)\NSIS\makensis.exe') {
    $makensisPath = 'C:\Program Files (x86)\NSIS\makensis.exe'
}
elseif (Test-Path 'C:\Program Files\NSIS\makensis.exe') {
    $makensisPath = 'C:\Program Files\NSIS\makensis.exe'
}
else {
    throw 'NSIS is required. Install it with: winget install -e --id NSIS.NSIS'
}

$nsisDir = Split-Path $makensisPath -Parent
if (($env:Path -split ';') -notcontains $nsisDir) {
    $env:Path = "$nsisDir;$env:Path"
}

Write-Host "Wails: $wailsPath" -ForegroundColor DarkGray
Write-Host "NSIS:  $makensisPath" -ForegroundColor DarkGray

Write-Host 'Running release gate...' -ForegroundColor Cyan
& "$PSScriptRoot\test.ps1"

$distIndex = Join-Path $root 'frontend\dist\index.html'
if (-not (Test-Path $distIndex)) {
    throw "Frontend release assets are missing: $distIndex"
}

Write-Host 'Building Windows application + NSIS installer...' -ForegroundColor Cyan

# Frontend was already installed, tested and compiled by the release gate.
# Wails' documented -s flag skips its frontend install/build phase, avoiding
# a second npm dependency-resolution pass.
Invoke-Native $wailsPath build -clean -s -platform windows/amd64 -nsis

$binDir = Join-Path $root 'build\bin'
if (-not (Test-Path $binDir)) {
    throw "Wails build returned success but output directory is missing: $binDir"
}

$installer = Get-ChildItem -Path $binDir -Recurse -File -Filter '*.exe' |
    Where-Object { $_.Name -match '(?i)(installer|setup)' } |
    Sort-Object LastWriteTime -Descending |
    Select-Object -First 1

if (-not $installer) {
    throw 'No NSIS setup/installer executable was found in build\bin.'
}

$distDir = Join-Path $root 'dist'
New-Item -ItemType Directory -Force $distDir | Out-Null

$out = Join-Path $distDir "MongoDeveloper-Setup-$Version.exe"
Copy-Item $installer.FullName $out -Force

$hash = (Get-FileHash $out -Algorithm SHA256).Hash.ToLowerInvariant()
Set-Content "$out.sha256" "$hash  MongoDeveloper-Setup-$Version.exe" -Encoding ascii

Write-Host ''
Write-Host '============================================================' -ForegroundColor Green
Write-Host 'MONGO DEVELOPER INSTALLER BUILD COMPLETE' -ForegroundColor Green
Write-Host '============================================================' -ForegroundColor Green
Write-Host "Installer: $out" -ForegroundColor Green
Write-Host "SHA256:    $hash" -ForegroundColor Green
