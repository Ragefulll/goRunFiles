go mod download
go mod tidy
go install github.com/wailsapp/wails/v2/cmd/wails@latest
winget install -e --id MSYS2.MSYS2
```
C:\msys64\usr\bin\bash.exe -lc "pacman -S --noconfirm mingw-w64-ucrt-x86_64-gcc"
$env:Path = "C:\msys64\ucrt64\bin;$env:Path"
gcc --version
```

Перед go run:
```.\run.ps1```

Для сборки exe:
```.\build.ps1```

Для корректного `NET` (ETW на Windows) сборка должна быть с `CGO_ENABLED=1` и доступным `gcc` (MinGW-w64 в `PATH`).
Скрипты `run.ps1` и `build.ps1` теперь включают это автоматически и проверяют наличие `gcc`.
