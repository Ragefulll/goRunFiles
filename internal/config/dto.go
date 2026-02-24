package config

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// ProcessDTO is a UI-friendly view of ProcessItem.
type ProcessDTO struct {
	Name         string `json:"name"`
	Disabled     bool   `json:"disabled"`
	Type         string `json:"type"`
	Process      string `json:"process"`
	Path         string `json:"path"`
	Command      string `json:"command"`
	Args         string `json:"args"`
	CheckProcess string `json:"checkProcess"`
	CheckCmdline string `json:"checkCmdline"`
	MonitorHang  bool   `json:"monitorHang"`
	HangTimeout  string `json:"hangTimeout"`
}

// SettingsDTO is a UI-friendly view of Settings.
type SettingsDTO struct {
	CheckTiming           string `json:"checkTiming"`
	RestartTiming         string `json:"restartTiming"`
	LaunchInNewConsole    bool   `json:"launchInNewConsole"`
	AutoCloseErrorDialogs bool   `json:"autoCloseErrorDialogs"`
	ErrorWindowTitles     string `json:"errorWindowTitles"`
	UseETWNetwork         bool   `json:"useETWNetwork"`
	NetDebug              bool   `json:"netDebug"`
	NetUnit               string `json:"netUnit"`
	NetScale              string `json:"netScale"`
}

// ConfigDTO is a UI-friendly view of Config.
type ConfigDTO struct {
	Processes []ProcessDTO `json:"processes"`
	Settings  SettingsDTO  `json:"settings"`
}

// ToDTO converts Config to ConfigDTO.
func ToDTO(cfg Config) ConfigDTO {
	names := make([]string, 0, len(cfg.Process))
	for name := range cfg.Process {
		names = append(names, name)
	}
	sort.Strings(names)

	out := ConfigDTO{
		Processes: make([]ProcessDTO, 0, len(names)),
		Settings: SettingsDTO{
			CheckTiming:           durString(cfg.Settings.CheckTiming),
			RestartTiming:         durString(cfg.Settings.RestartTiming),
			LaunchInNewConsole:    cfg.Settings.LaunchInNewConsole,
			AutoCloseErrorDialogs: cfg.Settings.AutoCloseErrorDialogs,
			ErrorWindowTitles:     cfg.Settings.ErrorWindowTitles,
			UseETWNetwork:         cfg.Settings.UseETWNetwork,
			NetDebug:              cfg.Settings.NetDebug,
			NetUnit:               cfg.Settings.NetUnit,
			NetScale:              floatToString(cfg.Settings.NetScale),
		},
	}

	for _, name := range names {
		p := cfg.Process[name]
		out.Processes = append(out.Processes, ProcessDTO{
			Name:         name,
			Disabled:     p.Disabled,
			Type:         p.Type,
			Process:      p.Process,
			Path:         p.Path,
			Command:      p.Command,
			Args:         p.Args,
			CheckProcess: p.CheckProcess,
			CheckCmdline: p.CheckCmdline,
			MonitorHang:  p.MonitorHang,
			HangTimeout:  durString(p.HangTimeout),
		})
	}
	return out
}

// FromDTO converts ConfigDTO into Config.
func FromDTO(dto ConfigDTO) (Config, error) {
	cfg := Config{
		Process: make(map[string]*ProcessItem),
	}

	var d Duration
	if err := d.UnmarshalText([]byte(dto.Settings.CheckTiming)); err != nil {
		return Config{}, fmt.Errorf("checkTiming: %w", err)
	}
	cfg.Settings.CheckTiming = d

	if err := d.UnmarshalText([]byte(dto.Settings.RestartTiming)); err != nil {
		return Config{}, fmt.Errorf("restartTiming: %w", err)
	}
	cfg.Settings.RestartTiming = d

	cfg.Settings.LaunchInNewConsole = dto.Settings.LaunchInNewConsole
	cfg.Settings.AutoCloseErrorDialogs = dto.Settings.AutoCloseErrorDialogs
	cfg.Settings.ErrorWindowTitles = dto.Settings.ErrorWindowTitles
	cfg.Settings.UseETWNetwork = dto.Settings.UseETWNetwork
	cfg.Settings.NetDebug = dto.Settings.NetDebug
	cfg.Settings.NetUnit = strings.TrimSpace(dto.Settings.NetUnit)
	cfg.Settings.NetScale = parseFloatOrZero(dto.Settings.NetScale)

	for _, p := range dto.Processes {
		name := strings.TrimSpace(p.Name)
		if name == "" {
			return Config{}, fmt.Errorf("process name is empty")
		}
		if _, exists := cfg.Process[name]; exists {
			return Config{}, fmt.Errorf("duplicate process name: %s", name)
		}

		var ht Duration
		if err := ht.UnmarshalText([]byte(p.HangTimeout)); err != nil {
			return Config{}, fmt.Errorf("hangTimeout for %s: %w", name, err)
		}

		cfg.Process[name] = &ProcessItem{
			Disabled:     p.Disabled,
			Type:         strings.TrimSpace(p.Type),
			Process:      p.Process,
			Path:         p.Path,
			Command:      p.Command,
			Args:         p.Args,
			CheckProcess: p.CheckProcess,
			CheckCmdline: p.CheckCmdline,
			MonitorHang:  p.MonitorHang,
			HangTimeout:  ht,
		}
	}
	return cfg, nil
}

func durString(d Duration) string {
	if d.Duration == 0 {
		return ""
	}
	return d.Duration.String()
}

func floatToString(v float64) string {
	if v == 0 {
		return ""
	}
	return fmt.Sprintf("%.0f", v)
}

func parseFloatOrZero(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}
