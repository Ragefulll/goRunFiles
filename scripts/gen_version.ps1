param(
  [bool]$Increment = $true
)

$ErrorActionPreference = "Stop"

$root = Split-Path $PSScriptRoot -Parent
$versionFile = Join-Path $root "version.txt"
$outFile = Join-Path $root "cmd\goRunFiles\generated_version.go"

if (!(Test-Path $versionFile)) {
  throw "version.txt not found"
}

$raw = (Get-Content $versionFile -Raw).Trim()
if ($raw -notmatch '^\d+\.\d+\.\d+$') {
  throw "version.txt must be in format x.y.z"
}

$parts = $raw.Split(".")
$major = [int]$parts[0]
$minor = [int]$parts[1]
$patch = [int]$parts[2]
if ($Increment) {
  $patch++
}

$newVersion = "$major.$minor.$patch"

if ($Increment) {
  Set-Content -Path $versionFile -Value $newVersion -Encoding ASCII
}

$content = @"
package main

const generatedVersion = "$newVersion"
"@

Set-Content -Path $outFile -Value $content -Encoding ASCII

Write-Output $newVersion
