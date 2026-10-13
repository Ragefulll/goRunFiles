# GoFiles

[![Downloads](https://img.shields.io/github/downloads/Ragefulll/goRunFiles/total?style=for-the-badge)](https://github.com/Ragefulll/goRunFiles/releases)
[![Release](https://img.shields.io/github/v/release/Ragefulll/goRunFiles?style=for-the-badge&label=Latest%20release)](https://github.com/Ragefulll/goRunFiles/releases/latest)
[![Discussions](https://img.shields.io/badge/Join-the%20Discussion-2D9F2D?style=for-the-badge&logo=github&logoColor=white)](https://github.com/Ragefulll/goRunFiles/releases/discussions)


## GO install

🖥️ **GO Command**:

```go mod download```

```go mod tidy```

```go install github.com/wailsapp/wails/v2/cmd/wails@latest```

```winget install -e --id MSYS2.MSYS2```

```
C:\msys64\usr\bin\bash.exe -lc "pacman -S --noconfirm mingw-w64-ucrt-x86_64-gcc"
$env:Path = "C:\msys64\ucrt64\bin;$env:Path"
gcc --version
```

🖥️ **NETWORD TEST**

```
go run github.com/Velocidex/etw/examples/tracer@latest -events network -kernel_event_type_filter "Send|Recv|TCP|UDP" "{9E814AAD-3204-11D2-9A82-006008A86939}"
```

## Usage
🖥️ **Запуск WEB версии, для вёрстки:**
```
cd cmd\goRunFilesWails
wails dev -browser
```

🖥️ **Запуск:**
```.\run.ps1```

🖥️ **Запуск UI:**
```.\run.ps1 - Gui```

🖥️ **Для сборки exe:**
```.\build.ps1```

🖥️ **Для сборки exe с UI:**
```.\build.ps1 -Gui```

> [!Note]
> Чтобы снизить ложные срабатывания антивирусов, exe собирается с полными метаданными (версия, компания, описание) и манифестом. Для Авторити-подписи (рекомендуется) задайте `$env:CODESIGN_THUMBPRINT="<thumbprint>"` (сертификат из хранилища) или `$env:CODESIGN_PFX="путь\cert.pfx"` / `$env:CODESIGN_PFX_PASS="пароль"` — `build.ps1` подпишет exe через `signtool`.

> [!Warning]
> Для корректного `NET` (ETW на Windows) сборка должна быть с `CGO_ENABLED=1` и доступным `gcc` (MinGW-w64 в `PATH`).
> Скрипты `run.ps1` и `build.ps1` теперь включают это автоматически и проверяют наличие `gcc`.