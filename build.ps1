param(
  [string]$Out = "goRunFiles.exe"
)

$ErrorActionPreference = "Stop"

$genScript = Join-Path $PSScriptRoot "scripts\gen_version.ps1"
if (!(Test-Path $genScript)) {
  throw "scripts\\gen_version.ps1 not found"
}

$newVersion = & $genScript -Increment $true

Write-Host "Building version $newVersion"
go build -ldflags "-X main.buildVersion=$newVersion" -o $Out .\cmd\goRunFiles
