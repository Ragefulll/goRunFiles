param(
    [string]$ExePath = 'D:\dev\www\goRunFiles\cmd\goRunFilesWails\build\bin\goRunFiles.exe',
    [int]$RestartOnExit = 1
)
$ErrorActionPreference = 'Stop'
if (-not (Test-Path -LiteralPath $ExePath)) { exit 2 }

if ($RestartOnExit -ne 1) {
    $p = Start-Process -FilePath $ExePath -PassThru
    try { $p.WaitForExit() } catch { }
    exit 0
}

while ($true) {
    $existing = Get-Process -ErrorAction SilentlyContinue | Where-Object { $_.Path -eq $ExePath } | Select-Object -First 1
    if ($null -ne $existing) {
        try { $existing.WaitForExit() } catch { Start-Sleep -Seconds 1 }
        Start-Sleep -Seconds 1
        continue
    }
    $p = Start-Process -FilePath $ExePath -PassThru
    try { $p.WaitForExit() } catch { Start-Sleep -Seconds 1 }
    Start-Sleep -Seconds 1
}
