package config

import (
	"fmt"
	"strconv"
	"strings"
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
	Disabled     bool
	Process      string
	CheckProcess string
	CheckCmdline string
	MonitorHang  bool
	HangTimeout  Duration
	Path         string
	Command      string
	Args         string
	Type         string // exe | cmd
	Pid          int
}

type Settings struct {
	CheckTiming Duration
	RestartTiming Duration
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

// Duration supports values like "100ms", "0.1s", "1s", "2m", or plain numbers (seconds).
type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalText(text []byte) error {
	s := strings.TrimSpace(string(text))
	if s == "" {
		d.Duration = 0
		return nil
	}
	if parsed, err := time.ParseDuration(s); err == nil {
		d.Duration = parsed
		return nil
	}
	// Fallback: plain number treated as seconds (supports decimals).
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fmt.Errorf("invalid duration %q", s)
	}
	d.Duration = time.Duration(float64(time.Second) * f)
	return nil
}
