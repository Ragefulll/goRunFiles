param(
  [string]$Out = "goRunFiles.exe",
  [switch]$Gui
)

$ErrorActionPreference = "Stop"

$genScript = Join-Path $PSScriptRoot "scripts\gen_version.ps1"
if (!(Test-Path $genScript)) {
  throw "scripts\\gen_version.ps1 not found"
}

$newVersion = & $genScript -Increment $true

Write-Host "Building version $newVersion"

if ($Gui) {
  Push-Location .\cmd\goRunFilesWails
  wails build -ldflags "-X main.buildVersion=$newVersion -H windowsgui"
  Pop-Location
  exit $LASTEXITCODE
}

go build -ldflags "-X main.buildVersion=$newVersion" -o $Out .\cmd\goRunFiles
