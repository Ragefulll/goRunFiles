param(
  [string]$Config = "",
  [string]$Exe = "",
  [switch]$NoGenerate,
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

if ($Gui) {
  Enable-CgoIfWindows
  Push-Location .\cmd\goRunFilesWails
  wails dev
  Pop-Location
  exit $LASTEXITCODE
}

if (-not $NoGenerate) {
  Enable-CgoIfWindows
  go generate .\cmd\goRunFiles
}

if ($Exe -ne "") {
  & $Exe
  exit $LASTEXITCODE
}

if ($Config -ne "") {
  Enable-CgoIfWindows
  go run .\cmd\goRunFiles $Config
  exit $LASTEXITCODE
}

Enable-CgoIfWindows
go run .\cmd\goRunFiles
