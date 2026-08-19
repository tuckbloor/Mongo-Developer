Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$projectRoot = Split-Path $PSScriptRoot -Parent
$version = '22.22.0'
$folderName = "node-v$version-win-x64"
$toolsDir = Join-Path $projectRoot '.tools'
$nodeDir = Join-Path $toolsDir $folderName
$nodeExe = Join-Path $nodeDir 'node.exe'
$npmCmd = Join-Path $nodeDir 'npm.cmd'

New-Item -ItemType Directory -Force $toolsDir | Out-Null

if (-not ((Test-Path $nodeExe) -and (Test-Path $npmCmd))) {
    $baseUrl = "https://nodejs.org/dist/v$version"
    $zipName = "$folderName.zip"
    $zipPath = Join-Path $toolsDir $zipName
    $sumsPath = Join-Path $toolsDir 'SHASUMS256.txt'

    if (-not ((Test-Path $zipPath) -and (Test-Path $sumsPath))) {
        Write-Host "Downloading pinned Node.js v$version..." -ForegroundColor Cyan
        Invoke-WebRequest -UseBasicParsing "$baseUrl/$zipName" -OutFile $zipPath
        Invoke-WebRequest -UseBasicParsing "$baseUrl/SHASUMS256.txt" -OutFile $sumsPath
    }
    else {
        Write-Host "Using bundled pinned Node.js archive v$version..." -ForegroundColor Cyan
    }

    $expectedLine = Get-Content $sumsPath |
        Where-Object { $_ -match [regex]::Escape($zipName) } |
        Select-Object -First 1

    if (-not $expectedLine) {
        throw "Could not find $zipName in Node.js SHASUMS256.txt"
    }

    $expected = ($expectedLine -split '\s+')[0].ToLowerInvariant()
    $actual = (Get-FileHash $zipPath -Algorithm SHA256).Hash.ToLowerInvariant()

    if ($actual -ne $expected) {
        Remove-Item $zipPath -Force -ErrorAction SilentlyContinue
        throw "Node.js archive checksum mismatch. Expected $expected but got $actual"
    }

    Write-Host 'Node.js archive checksum verified.' -ForegroundColor Green

    if (Test-Path $nodeDir) {
        Remove-Item $nodeDir -Recurse -Force
    }

    Expand-Archive -Path $zipPath -DestinationPath $toolsDir -Force
}

if (-not (Test-Path $nodeExe)) {
    throw "Pinned node.exe not found at $nodeExe"
}
if (-not (Test-Path $npmCmd)) {
    throw "Pinned npm.cmd not found at $npmCmd"
}

$env:Path = "$nodeDir;$env:Path"

# IMPORTANT: isolate npm state as well as the npm executable.
# This avoids a damaged machine-wide npm cache/config contaminating this build.
$env:npm_config_cache = Join-Path $toolsDir 'npm-cache'
$env:npm_config_update_notifier = 'false'
$env:npm_config_fund = 'false'
$env:npm_config_audit = 'false'

New-Item -ItemType Directory -Force $env:npm_config_cache | Out-Null

$nodeVersion = (& $nodeExe --version).Trim()
$npmVersion = (& $npmCmd --version).Trim()
if ($LASTEXITCODE -ne 0) {
    throw 'Pinned npm failed to start.'
}

Write-Host "Pinned toolchain: Node $nodeVersion / npm $npmVersion" -ForegroundColor DarkGray
Write-Host "Private npm cache: $env:npm_config_cache" -ForegroundColor DarkGray
