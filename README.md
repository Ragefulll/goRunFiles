go mod download
go mod tidy
go install github.com/wailsapp/wails/v2/cmd/wails@latest
winget install -e --id MSYS2.MSYS2
```
C:\msys64\usr\bin\bash.exe -lc "pacman -S --noconfirm mingw-w64-ucrt-x86_64-gcc"
$env:Path = "C:\msys64\ucrt64\bin;$env:Path"
gcc --version
```

```
go run github.com/Velocidex/etw/examples/tracer@latest -events network -kernel_event_type_filter "Send|Recv|TCP|UDP" "{9E814AAD-3204-11D2-9A82-006008A86939}"
```

Запуск:
```.\run.ps1```

Запуск UI:
```.\run.ps1 - Gui```

Для сборки exe:
```.\build.ps1```

Для сборки exe с UI:
```.\build.ps1 -Gui```

Для корректного `NET` (ETW на Windows) сборка должна быть с `CGO_ENABLED=1` и доступным `gcc` (MinGW-w64 в `PATH`).
Скрипты `run.ps1` и `build.ps1` теперь включают это автоматически и проверяют наличие `gcc`.
