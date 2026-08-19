param([Parameter(Mandatory=$true)][string]$Version)
$ErrorActionPreference='Stop'
$root=Split-Path $PSScriptRoot -Parent
Set-Location $root
if (git status --porcelain) { throw 'Working tree is not clean.' }
& "$PSScriptRoot\build-installer.ps1" -Version $Version
git tag -a "v$Version" -m "Mongo Developer v$Version"
Write-Host "Release artifacts are in dist\. Push with: git push origin main --tags" -ForegroundColor Green
