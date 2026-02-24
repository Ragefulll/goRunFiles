//go:build !windows

package process

import "os/exec"

func hideCmdWindow(cmd *exec.Cmd) {}
