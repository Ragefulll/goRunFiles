param(
  [string]$Config = "",
  [string]$Exe = "",
  [switch]$NoGenerate,
  [switch]$Gui
)

$ErrorActionPreference = "Stop"

if ($Gui) {
  Push-Location .\cmd\goRunFilesWails
  wails dev
  Pop-Location
  exit $LASTEXITCODE
}

if (-not $NoGenerate) {
  go generate .\cmd\goRunFiles
}

if ($Exe -ne "") {
  & $Exe
  exit $LASTEXITCODE
}

if ($Config -ne "") {
  go run .\cmd\goRunFiles $Config
  exit $LASTEXITCODE
}

go run .\cmd\goRunFiles
