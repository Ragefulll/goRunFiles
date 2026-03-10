param(
  [string]$Out = "goRunFiles.exe",
  [switch]$Gui
)

$ErrorActionPreference = "Stop"

$rootConfig = Join-Path $PSScriptRoot "config.ini"

function Copy-ConfigNextToExe {
  param([string]$ExePath)
  if (-not (Test-Path $rootConfig)) { return }
  if ([string]::IsNullOrWhiteSpace($ExePath)) { return }
  $destDir = Split-Path -Parent $ExePath
  if (-not (Test-Path $destDir)) {
    New-Item -ItemType Directory -Path $destDir | Out-Null
  }
  Copy-Item $rootConfig (Join-Path $destDir "config.ini") -Force
}

function Enable-CgoIfWindows {
  if (-not $IsWindows) { return }
  $env:CGO_ENABLED = "1"
  if (-not (Get-Command gcc -ErrorAction SilentlyContinue)) {
    $gccDirs = @(
      "C:\msys64\ucrt64\bin",
      "C:\msys64\mingw64\bin",
      "C:\mingw64\bin",
      "C:\Program Files\mingw-w64\mingw64\bin"
    )
    foreach ($dir in $gccDirs) {
      if (Test-Path (Join-Path $dir "gcc.exe")) {
        $env:PATH = "$dir;$env:PATH"
        break
      }
    }
  }
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
  if ($LASTEXITCODE -eq 0) {
    $wailsConfig = Join-Path $PSScriptRoot "cmd\goRunFilesWails\wails.json"
    $outputName = "goRunFiles"
    if (Test-Path $wailsConfig) {
      try {
        $outputName = (Get-Content $wailsConfig | ConvertFrom-Json).outputfilename
      } catch {
        $outputName = "goRunFiles"
      }
    }
    $exeName = $outputName
    if ($IsWindows) { $exeName = "$exeName.exe" }
    $guiExe = Join-Path $PSScriptRoot "cmd\goRunFilesWails\build\bin\$exeName"
    Copy-ConfigNextToExe $guiExe
  }
  exit $LASTEXITCODE
}

Enable-CgoIfWindows
go build -ldflags "-X main.buildVersion=$newVersion" -o $Out .\cmd\goRunFiles
if ($LASTEXITCODE -eq 0) {
  Copy-ConfigNextToExe $Out
}
