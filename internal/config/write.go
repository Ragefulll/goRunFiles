package config

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// WriteFromDTO writes config.ini content from ConfigDTO.
func WriteFromDTO(path string, dto ConfigDTO) error {
	var b strings.Builder

	names := make([]string, 0, len(dto.Processes))
	procByName := make(map[string]ProcessDTO, len(dto.Processes))
	for _, p := range dto.Processes {
		name := strings.TrimSpace(p.Name)
		if name == "" {
			continue
		}
		procByName[name] = p
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		p := procByName[name]
		b.WriteString(fmt.Sprintf("[process %q]\n", name))
		b.WriteString(fmt.Sprintf("disabled=%v\n", p.Disabled))
		if p.Path != "" {
			b.WriteString(fmt.Sprintf("path=%s\n", quoteIfNeeded(p.Path)))
		}
		if p.Command != "" {
			b.WriteString(fmt.Sprintf("command=%s\n", quoteIfNeeded(p.Command)))
		}
		if p.Process != "" {
			b.WriteString(fmt.Sprintf("process=%s\n", quoteIfNeeded(p.Process)))
		}
		if p.CheckProcess != "" {
			b.WriteString(fmt.Sprintf("checkProcess=%s\n", quoteIfNeeded(p.CheckProcess)))
		}
		if p.CheckCmdline != "" {
			b.WriteString(fmt.Sprintf("checkCmdline=%s\n", quoteIfNeeded(p.CheckCmdline)))
		}
		if p.Args != "" {
			b.WriteString(fmt.Sprintf("args=%s\n", quoteIfNeeded(p.Args)))
		}
		if p.Type != "" {
			b.WriteString(fmt.Sprintf("type=%s\n", p.Type))
		}
		b.WriteString(fmt.Sprintf("monitorHang=%v\n", p.MonitorHang))
		if strings.TrimSpace(p.HangTimeout) != "" {
			b.WriteString(fmt.Sprintf("hangTimeout=%s\n", p.HangTimeout))
		}
		b.WriteString("\n")
	}

	b.WriteString("[settings]\n")
	if strings.TrimSpace(dto.Settings.CheckTiming) != "" {
		b.WriteString(fmt.Sprintf("checkTiming=%s\n", dto.Settings.CheckTiming))
	}
	if strings.TrimSpace(dto.Settings.RestartTiming) != "" {
		b.WriteString(fmt.Sprintf("restartTiming=%s\n", dto.Settings.RestartTiming))
	}
	b.WriteString(fmt.Sprintf("launchInNewConsole=%v\n", dto.Settings.LaunchInNewConsole))
	b.WriteString(fmt.Sprintf("autoCloseErrorDialogs=%v\n", dto.Settings.AutoCloseErrorDialogs))
	if strings.TrimSpace(dto.Settings.ErrorWindowTitles) != "" {
		b.WriteString(fmt.Sprintf("errorWindowTitles=%s\n", quoteIfNeeded(dto.Settings.ErrorWindowTitles)))
	}

	return os.WriteFile(path, []byte(b.String()), 0644)
}

func quoteIfNeeded(s string) string {
	if s == "" {
		return "\"\""
	}
	need := false
	for _, r := range s {
		if r == ' ' || r == '\\' || r == '"' || r == ',' {
			need = true
			break
		}
	}
	if !need {
		return s
	}
	escaped := strings.ReplaceAll(s, `"`, `\"`)
	return `"` + escaped + `"`
}
