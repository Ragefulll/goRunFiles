package config

import (
	"time"

	"gopkg.in/gcfg.v1"
)

const Banner = "\n\n\n\n╱╭━━━╮╱╭━━━╮╱╭━━━━╮╱╭━━━╮╱╭━━━╮\n╱┃╭━╮┃╱┃╭━╮┃╱┃╭╮╭╮┃╱┃╭━╮┃╱╰╮╭╮┃\n╱┃┃╱┃┃╱┃╰━╯┃╱╰╯┃┃╰╯╱╰╯╭╯┃╱╱┃┃┃┃\n╱┃╰━╯┃╱┃╭╮╭╯╱╱╱┃┃╱╱╱╭╮╰╮┃╱╱┃┃┃┃\n╱┃╭━╮┃╱┃┃┃╰╮╱╱╱┃┃╱╱╱┃╰━╯┃╱╭╯╰╯┃\n╱╰╯╱╰╯╱╰╯╰━╯╱╱╱╰╯╱╱╱╰━━━╯╱╰━━━╯\n\n\n"
const DefaultConfigName = "config.ini"

const (
	TypeExe = "exe"
	TypeCmd = "cmd"
	TypeBat = "bat"
)

// ProcessItem Один процесс
type ProcessItem struct {
	Process      string
	CheckProcess string
	CheckCmdline string
	Path         string
	Command      string
	Args         string
	Type         string // exe | cmd
	Pid          int
}

type Settings struct {
	CheckTiming time.Duration
	RestartTiming time.Duration
	LaunchInNewConsole bool
	AutoCloseErrorDialogs bool
	ErrorWindowTitles string
}

// Config Вся конфигурация
type Config struct {
	Process  map[string]*ProcessItem
	Settings Settings
}

func Load(path string) (Config, error) {
	var cfg Config
	if err := gcfg.ReadFileInto(&cfg, path); err != nil {
		return Config{}, err
	}
	return cfg, nil
}
