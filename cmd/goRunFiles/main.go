package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"goRunFiles/internal/app"
	"goRunFiles/internal/config"
)

var buildVersion = generatedVersion

func main() {
	configPath := config.DefaultConfigName
	if exePath, err := os.Executable(); err == nil {
		exeConfig := filepath.Join(filepath.Dir(exePath), config.DefaultConfigName)
		if _, err := os.Stat(exeConfig); err == nil {
			configPath = exeConfig
		}
	}

	log.Println(config.Banner)

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Printf("%s [ART3D-CHEKER]: Ошибка загрузки конфига: %v", app.LogTag, err)
		return
	}

	application := app.New(cfg, log.Default(), buildVersion)
	if err := application.Run(context.Background()); err != nil {
		log.Printf("%s [ART3D-CHEKER]: Приложение остановлено: %v", app.LogTag, err)
	}
}
