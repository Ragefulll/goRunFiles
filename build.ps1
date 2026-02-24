param(
  [string]$Out = "goRunFiles.exe",
  [switch]$Gui
)

$ErrorActionPreference = "Stop"

function Enable-CgoIfWindows {
  if (-not $IsWindows) { return }
  $env:CGO_ENABLED = "1"
  try {
    $null = & gcc --version 2>$null
  } catch {
    throw "CGO_ENABLED=1 requires gcc in PATH (install MinGW-w64 and ensure gcc is available)."
  }
}

$genScript = Join-Path $PSScriptRoot "scripts\gen_version.ps1"
if (!(Test-Path $genScript)) {
  throw "scripts\\gen_version.ps1 not found"
}

$newVersion = & $genScript -Increment $true

Write-Host "Building version $newVersion"

if ($Gui) {
  Enable-CgoIfWindows
  Push-Location .\cmd\goRunFilesWails
  wails build -ldflags "-X main.buildVersion=$newVersion -H windowsgui"
  Pop-Location
  exit $LASTEXITCODE
}

Enable-CgoIfWindows
go build -ldflags "-X main.buildVersion=$newVersion" -o $Out .\cmd\goRunFiles
