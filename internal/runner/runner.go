package runner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"goRunFiles/internal/config"
	"goRunFiles/internal/display"
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
		args = injectWindowPosition(args, item.Screen, processPath)
		cmd := exec.Command(processPath, args...)
		cmd.Dir = filepath.Dir(processPath)
		hideWindow(cmd)

		if err := cmd.Start(); err != nil {
			return 0, err
		}
		go cmd.Wait()
		moveWindowAsync(cmd.Process.Pid, item.Screen)
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
		go cmd.Wait()
		moveWindowAsync(cmd.Process.Pid, item.Screen)
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
		go cmd.Wait()
		moveWindowAsync(cmd.Process.Pid, item.Screen)
		return cmd.Process.Pid, nil
	default:
		return 0, fmt.Errorf("unknown process type %q", item.Type)
	}
}

func moveWindowAsync(pid int, screen int) {
	if pid <= 0 || screen <= 0 {
		return
	}
	go func() {
		_ = display.MoveProcessWindowToScreen(pid, screen)
	}()
}

func injectWindowPosition(args []string, screen int, processPath string) []string {
	if screen <= 0 || len(args) == 0 {
		return args
	}
	screens, err := display.ListScreens()
	if err != nil || screen > len(screens) {
		return args
	}
	target := screens[screen-1]
	pos := fmt.Sprintf("--window-position=%d,%d", target.X, target.Y)
	out := make([]string, 0, len(args)+2)
	replaced := false
	hasUserDataDir := false
	for _, a := range args {
		if strings.HasPrefix(strings.ToLower(a), "--window-position=") {
			out = append(out, pos)
			replaced = true
		} else {
			out = append(out, a)
		}
		if strings.HasPrefix(strings.ToLower(a), "--user-data-dir") {
			hasUserDataDir = true
		}
	}
	if !replaced {
		out = append(out, pos)
	}
	if !hasUserDataDir {
		exe := strings.ToLower(filepath.Base(processPath))
		if exe == "chrome.exe" {
			userDir := fmt.Sprintf(`--user-data-dir=%s\art3d-chrome-%d`, os.TempDir(), screen)
			out = append(out, userDir, "--no-first-run")
		}
	}
	return out
}

func splitArgs(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var args []string
	var current strings.Builder
	inQuote := false
	quoteChar := byte(0)
	for i := 0; i < len(raw); i++ {
		ch := raw[i]
		if inQuote {
			if ch == quoteChar {
				inQuote = false
				continue
			}
			current.WriteByte(ch)
		} else {
			if ch == '"' || ch == '\'' {
				inQuote = true
				quoteChar = ch
				continue
			}
			if ch == ' ' || ch == '\t' {
				if current.Len() > 0 {
					args = append(args, current.String())
					current.Reset()
				}
				continue
			}
			current.WriteByte(ch)
		}
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	return args
}
