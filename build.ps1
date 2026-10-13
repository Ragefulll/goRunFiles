param(
  [string]$Out = "goRunFiles.exe",
  [switch]$Gui,
  [switch]$Current
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

function Test-SigningConfigured {
  return -not [string]::IsNullOrWhiteSpace($env:CODESIGN_THUMBPRINT) -or -not [string]::IsNullOrWhiteSpace($env:CODESIGN_PFX)
}

function Get-SignTool {
  $kitRoot = Join-Path ${env:ProgramFiles(x86)} "Windows Kits\10\bin"
  if (Test-Path $kitRoot) {
    $found = Get-ChildItem $kitRoot -Recurse -Filter "signtool.exe" -ErrorAction SilentlyContinue |
      Sort-Object FullName -Descending | Select-Object -First 1
    if ($found) { return $found.FullName }
  }
  $cmd = Get-Command "signtool" -ErrorAction SilentlyContinue
  if ($cmd) { return $cmd.Source }
  return $null
}

function Invoke-ExeSigning {
  param([string]$ExePath)
  if ([string]::IsNullOrWhiteSpace($ExePath) -or -not (Test-Path $ExePath)) { return }
  if (-not (Test-SigningConfigured)) {
    Write-Host "Code signing skipped (set CODESIGN_THUMBPRINT or CODESIGN_PFX to sign)." -ForegroundColor DarkGray
    return
  }
  $signtool = Get-SignTool
  if (-not $signtool) {
    Write-Warning "Code signing is configured but signtool.exe was not found. Skipping signature."
    return
  }
  $sArgs = @("sign", "/fd", "SHA256")
  if (-not [string]::IsNullOrWhiteSpace($env:CODESIGN_THUMBPRINT)) {
    $sArgs += @("/sha1", $env:CODESIGN_THUMBPRINT)
  } elseif (-not [string]::IsNullOrWhiteSpace($env:CODESIGN_PFX)) {
    $sArgs += @("/f", $env:CODESIGN_PFX)
    if (-not [string]::IsNullOrWhiteSpace($env:CODESIGN_PFX_PASS)) {
      $sArgs += @("/p", $env:CODESIGN_PFX_PASS)
    }
  }
  if ([string]::IsNullOrWhiteSpace($env:CODESIGN_TIMESTAMP)) {
    $sArgs += @("/tr", "http://timestamp.digicert.com", "/td", "SHA256")
  }
  $sArgs += @($ExePath)
  & $signtool $sArgs
  if ($LASTEXITCODE -ne 0) {
    Write-Warning "signtool failed with exit code $LASTEXITCODE"
  }
}

function Sync-WailsVersion {
  param([string]$Version)
  $wailsConfig = Join-Path $PSScriptRoot "cmd\goRunFilesWails\wails.json"
  if (-not (Test-Path $wailsConfig)) { return }
  $c = Get-Content $wailsConfig -Raw | ConvertFrom-Json
  if (-not $c.info) { $c.info = [pscustomobject]@{} }
  $c.info.productVersion = $Version
  $c | ConvertTo-Json -Depth 10 | Set-Content $wailsConfig -Encoding UTF8
}

$genScript = Join-Path $PSScriptRoot "scripts\gen_version.ps1"
if (!(Test-Path $genScript)) {
  throw "scripts\\gen_version.ps1 not found"
}

if ($Current) {
    $newVersion = & $genScript -Increment $false
} else {
    $newVersion = & $genScript -Increment $true
}

Write-Host "Building version $newVersion"

if ($Gui) {
  Enable-CgoIfWindows
  Sync-WailsVersion $newVersion
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
    $rootVersion = Join-Path $PSScriptRoot "version.txt"
    if (Test-Path $rootVersion) {
      Copy-Item $rootVersion (Join-Path (Split-Path -Parent $guiExe) "version.txt") -Force
    }
    Invoke-ExeSigning $guiExe
  }
  exit $LASTEXITCODE
}

Enable-CgoIfWindows
go build -trimpath -ldflags "-X main.buildVersion=$newVersion" -o $Out .\cmd\goRunFiles
if ($LASTEXITCODE -eq 0) {
  Copy-ConfigNextToExe $Out
  Invoke-ExeSigning $Out
}
