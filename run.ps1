param(
  [string]$Config = "",
  [string]$Exe = "",
  [switch]$NoGenerate
)

$ErrorActionPreference = "Stop"

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
