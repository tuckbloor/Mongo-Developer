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

& "$PSScriptRoot\bootstrap-node.ps1"

$node = (Get-Command node.exe -ErrorAction Stop).Source
$npm = (Get-Command npm.cmd -ErrorAction Stop).Source

Write-Host 'Preparing Go modules...' -ForegroundColor Cyan
Invoke-Native go.exe mod tidy
Invoke-Native go.exe mod verify

Write-Host 'Go tests...' -ForegroundColor Cyan
Invoke-Native go.exe test ./internal/...

Write-Host 'Frontend clean install + tests + build...' -ForegroundColor Cyan
Push-Location (Join-Path $root 'frontend')
try {
    # Always clear the failed dependency tree from the previous build.
    if (Test-Path '.\node_modules') {
        Write-Host 'Removing old node_modules...' -ForegroundColor Yellow
        Remove-Item '.\node_modules' -Recurse -Force
    }

    # A lock produced by the previous broken install is not trusted.
    if (Test-Path '.\package-lock.json') {
        Write-Host 'Removing old package-lock.json...' -ForegroundColor Yellow
        Remove-Item '.\package-lock.json' -Force
    }

    Write-Host 'Installing frontend dependencies with isolated npm cache...' -ForegroundColor Cyan

    # npm/Arborist currently has an edgesOut crash in peer-set construction.
    # legacy-peer-deps bypasses that peer-tree path. The Windows Rolldown native
    # binding is also an explicit dependency in package.json rather than relying
    # on npm optional-dependency selection.
    Invoke-Native $npm install --legacy-peer-deps --include=optional --no-audit --no-fund

    $binding = '.\node_modules\@rolldown\binding-win32-x64-msvc'
    if (-not (Test-Path $binding)) {
        throw "Required Windows Rolldown native binding was not installed: $binding"
    }

    Invoke-Native $npm run test:smoke
    Invoke-Native $npm test -- --run
    Invoke-Native $npm run build

    $distIndex = '.\dist\index.html'
    if (-not (Test-Path $distIndex)) {
        throw "Frontend build returned success but $distIndex does not exist."
    }
}
finally {
    Pop-Location
}

Write-Host 'ALL RELEASE-GATE TESTS PASSED' -ForegroundColor Green
