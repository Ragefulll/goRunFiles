@echo off
setlocal EnableExtensions EnableDelayedExpansion

set "SRC=%~dp0cmd\goRunFilesWails\build\bin"
set "DST=X:\feed.art3d.ru\gorunfiles"
set "VERSION=%~dp0version.txt"

if not exist "%SRC%" (
    echo Source not found: "%SRC%"
    exit /b 1
)

if not exist "%DST%" (
    mkdir "%DST%"
    if errorlevel 1 (
        echo Failed to create destination: "%DST%"
        exit /b 1
    )
)

if not exist "%VERSION%" (
    echo version.txt not found: "%VERSION%"
    exit /b 1
)

copy /y "%SRC%\goRunFiles.exe" "%DST%\goRunFiles.exe" >nul
if errorlevel 1 (
    echo Failed to copy goRunFiles.exe
    exit /b 1
)

copy /y "%VERSION%" "%DST%\version.txt" >nul
if errorlevel 1 (
    echo Failed to copy version.txt
    exit /b 1
)

echo Done: goRunFiles.exe and version.txt ^> "%DST%"
exit /b 0
