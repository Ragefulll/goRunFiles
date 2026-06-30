package runner

import (
	"fmt"
	"log"
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
		if err := display.MoveProcessWindowToScreen(pid, screen); err != nil {
			log.Printf("moveWindowAsync: pid=%d screen=%d error: %v", pid, screen, err)
		}
	}()
}

func injectWindowPosition(args []string, screen int, processPath string) []string {
	log.Printf("[injectWindowPosition] screen=%d args=%q", screen, args)
	if screen <= 0 || len(args) == 0 {
		log.Printf("[injectWindowPosition] early return: screen=%d args_len=%d", screen, len(args))
		return args
	}
	screens, err := display.ListScreens()
	log.Printf("[injectWindowPosition] ListScreens: count=%d err=%v", len(screens), err)
	if err != nil || screen > len(screens) {
		log.Printf("[injectWindowPosition] screen out of range or err: screen=%d screens_count=%d err=%v", screen, len(screens), err)
		return args
	}
	target := screens[screen-1]
	pos := fmt.Sprintf("--window-position=%d,%d", target.X, target.Y)
	log.Printf("[injectWindowPosition] target screen index=%d pos=%s", screen-1, pos)
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
	// Chrome single-instance workaround: force separate instance so that
	// --window-position is actually honoured and moveWindowAsync can find
	// the real browser PID in the process tree.
	if !hasUserDataDir {
		exe := strings.ToLower(filepath.Base(processPath))
		if exe == "chrome.exe" {
			userDir := fmt.Sprintf(`--user-data-dir=%s\art3d-chrome-%d`, os.TempDir(), screen)
			out = append(out, userDir, "--no-first-run")
			log.Printf("[injectWindowPosition] added %s for separate Chrome instance", userDir)
		}
	}
	log.Printf("[injectWindowPosition] final args=%q", out)
	return out
}

func splitArgs(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	return strings.Fields(raw)
}
