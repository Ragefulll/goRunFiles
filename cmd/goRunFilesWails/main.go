package main

import (
	"context"
	"embed"
	"log"
	"os/exec"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"goRunFiles/internal/app"
	"goRunFiles/internal/config"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed frontend/*
var assets embed.FS

var buildVersion = "dev"

type GUI struct {
	mon        *app.App
	configPath string
	mu         sync.RWMutex
	snapshot   app.DisplaySnapshot
}

func main() {
	configPath := resolveConfigPath()
	cfg, err := config.Load(configPath)
	if err != nil {
		if _, repErr := config.RepairFile(configPath); repErr == nil {
			if cfg2, err2 := config.Load(configPath); err2 == nil {
				cfg = cfg2
			} else {
				log.Printf("%s [ART3D-CHEKER]: Ошибка загрузки конфига: %v", app.LogTag, err2)
				return
			}
		} else {
			log.Printf("%s [ART3D-CHEKER]: Ошибка загрузки конфига: %v", app.LogTag, err)
			return
		}
	}

	gui := &GUI{
		mon:        app.New(cfg, log.Default(), buildVersion),
		configPath: configPath,
	}

	err = wails.Run(&options.App{
		Title:  "ART3D Process Monitor",
		Width:  1700,
		Height: 900,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: func(ctx context.Context) {
			go func() {
				_ = gui.mon.RunWithObserver(ctx, gui.updateSnapshot)
			}()
		},
		OnShutdown: func(ctx context.Context) {
			// Keep child processes running after UI closes.
		},
		Bind: []interface{}{gui},
	})
	if err != nil {
		log.Printf("%s [ART3D-CHEKER]: Wails error: %v", app.LogTag, err)
	}
}

func resolveConfigPath() string {
	configPath := config.DefaultConfigName
	if exePath, err := os.Executable(); err == nil {
		exeConfig := filepath.Join(filepath.Dir(exePath), config.DefaultConfigName)
		if _, err := os.Stat(exeConfig); err == nil {
			configPath = exeConfig
		}
	}
	if cwd, err := os.Getwd(); err == nil {
		cwdConfig := filepath.Join(cwd, config.DefaultConfigName)
		if _, err := os.Stat(cwdConfig); err == nil {
			configPath = cwdConfig
		}
	}
	if exePath, err := os.Executable(); err == nil {
		dir := filepath.Dir(exePath)
		for i := 0; i < 5; i++ {
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			candidate := filepath.Join(parent, config.DefaultConfigName)
			if _, err := os.Stat(candidate); err == nil {
				configPath = candidate
				break
			}
			dir = parent
		}
	}
	return configPath
}

func (g *GUI) updateSnapshot(s app.DisplaySnapshot) {
	g.mu.Lock()
	g.snapshot = s
	g.mu.Unlock()
}

// GetSnapshot returns the latest snapshot for UI polling.
func (g *GUI) GetSnapshot() app.DisplaySnapshot {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.snapshot
}

// Start starts a process by config name.
func (g *GUI) Start(name string) error {
	return g.mon.StartProcess(name)
}

// Stop stops a process by config name.
func (g *GUI) Stop(name string) error {
	return g.mon.StopProcess(name)
}

// Restart restarts a process by config name.
func (g *GUI) Restart(name string) error {
	return g.mon.RestartProcess(name)
}

// RestartAll restarts all enabled processes.
func (g *GUI) RestartAll() error {
	return g.mon.RestartAll()
}

// KillCMD force-kills all cmd.exe processes.
func (g *GUI) KillCMD() error {
	return killImage("cmd.exe")
}

// KillNode force-kills all node.exe processes.
func (g *GUI) KillNode() error {
	return killImage("node.exe")
}

func killImage(imageName string) error {
	if runtime.GOOS != "windows" {
		return nil
	}

	// Ignore "not found" cases (exit code 128) so button is idempotent.
	cmd := exec.Command("cmd", "/C", "taskkill /F /IM "+imageName+" >nul 2>&1 || exit /b 0")
	return cmd.Run()
}

// GetConfig returns the current config.ini content.
func (g *GUI) GetConfigModel() (config.ConfigDTO, error) {
	cfg, err := config.Load(g.configPath)
	if err != nil {
		return config.ConfigDTO{}, err
	}
	return config.ToDTO(cfg), nil
}

// SaveConfig writes config.ini and reloads it.
func (g *GUI) SaveConfigModel(dto config.ConfigDTO) error {
	if err := config.WriteFromDTO(g.configPath, dto); err != nil {
		return err
	}
	cfg, err := config.FromDTO(dto)
	if err != nil {
		return err
	}
	g.mon.UpdateConfig(cfg)
	return nil
}
