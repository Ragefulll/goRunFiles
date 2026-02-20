package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"goRunFiles/internal/config"
)

// Start launches the process described by item and returns a PID for cmd tasks.
func Start(item *config.ProcessItem, launchInNewConsole bool) (int, error) {
	processPath := filepath.Join(item.Path, item.Process)
	switch item.Type {
	case config.TypeExe:
		if _, err := os.Stat(processPath); err != nil {
			if os.IsNotExist(err) {
				return 0, fmt.Errorf("file %s does not exist", processPath)
			}
			return 0, err
		}

		args := splitArgs(item.Args)
		cmd := exec.Command(processPath, args...)
		cmd.Dir = filepath.Dir(processPath)
		hideWindow(cmd)

		if err := cmd.Start(); err != nil {
			return 0, err
		}
		return cmd.Process.Pid, nil
	case config.TypeCmd:
		var cmd *exec.Cmd
		if launchInNewConsole {
			cmd = exec.Command("cmd.exe", "/C", "start", "", "cmd.exe", "/C", item.Command)
		} else {
			cmd = exec.Command("cmd.exe", "/C", item.Command)
			hideWindow(cmd)
		}
		cmd.Dir = item.Path

		if err := cmd.Start(); err != nil {
			return 0, err
		}
		return cmd.Process.Pid, nil
	case config.TypeBat:
		if item.Process == "" {
			return 0, fmt.Errorf("bat process is empty")
		}
		if _, err := os.Stat(processPath); err != nil {
			if os.IsNotExist(err) {
				return 0, fmt.Errorf("file %s does not exist", processPath)
			}
			return 0, err
		}
		args := splitArgs(item.Args)
		var cmd *exec.Cmd
		if launchInNewConsole {
			callArgs := append([]string{"/C", "start", "", "cmd.exe", "/C", "call", processPath}, args...)
			cmd = exec.Command("cmd.exe", callArgs...)
		} else {
			callArgs := append([]string{"/C", "call", processPath}, args...)
			cmd = exec.Command("cmd.exe", callArgs...)
			hideWindow(cmd)
		}
		cmd.Dir = filepath.Dir(processPath)

		if err := cmd.Start(); err != nil {
			return 0, err
		}
		return cmd.Process.Pid, nil
	default:
		return 0, fmt.Errorf("unknown process type %q", item.Type)
	}
}

func splitArgs(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	return strings.Fields(raw)
}
