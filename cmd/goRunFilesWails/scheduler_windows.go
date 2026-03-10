//go:build windows

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"goRunFiles/internal/config"
)

const (
	schedulerTaskName   = "goRunFilesWails_AutoStart"
	schedulerScriptName = "goRunFiles.watch.ps1"
)

type TaskStatus struct {
	Installed      bool   `json:"installed"`
	Running        bool   `json:"running"`
	State          string `json:"state"`
	LastRunTime    string `json:"lastRunTime"`
	LastTaskResult int    `json:"lastTaskResult"`
	TaskName       string `json:"taskName"`
	Error          string `json:"error"`
}

type schedulerQuery struct {
	Exists         bool   `json:"Exists"`
	State          string `json:"State"`
	LastRunTime    string `json:"LastRunTime"`
	LastTaskResult int    `json:"LastTaskResult"`
}

func (g *GUI) GetSchedulerStatus() (TaskStatus, error) {
	status := TaskStatus{TaskName: schedulerTaskName}
	info, err := queryScheduler()
	if err != nil {
		status.Error = err.Error()
		return status, nil
	}
	status.Installed = info.Exists
	status.State = info.State
	status.LastRunTime = info.LastRunTime
	status.LastTaskResult = info.LastTaskResult
	status.Running = strings.EqualFold(info.State, "running")
	return status, nil
}

func (g *GUI) InstallScheduler() error {
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("exe path: %w", err)
	}
	exePath, _ = filepath.Abs(exePath)
	restartOnExit := true
	if cfg, cfgErr := config.Load(g.configPath); cfgErr == nil {
		restartOnExit = cfg.Settings.AutoRestartOnExit
	}
	scriptPath, err := ensureWatchdogScript(exePath, restartOnExit)
	if err != nil {
		return err
	}
	cmd := buildSchedulerCommand(scriptPath, exePath, restartOnExit)
	args := []string{
		"/Create",
		"/TN", schedulerTaskName,
		"/SC", "ONLOGON",
		"/RL", "HIGHEST",
		"/F",
		"/TR", cmd,
	}
	out, err := exec.Command("schtasks", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("schtasks create: %v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func (g *GUI) RemoveScheduler() error {
	_ = exec.Command("schtasks", "/End", "/TN", schedulerTaskName).Run()
	out, err := exec.Command("schtasks", "/Delete", "/TN", schedulerTaskName, "/F").CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if strings.Contains(strings.ToLower(msg), "cannot find") || strings.Contains(strings.ToLower(msg), "не найден") {
			return nil
		}
		return fmt.Errorf("schtasks delete: %v: %s", err, msg)
	}
	return nil
}

func updateSchedulerScriptIfInstalled(cfg config.Config) error {
	info, err := queryScheduler()
	if err != nil || !info.Exists {
		return nil
	}
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("exe path: %w", err)
	}
	exePath, _ = filepath.Abs(exePath)
	scriptPath, err := ensureWatchdogScript(exePath, cfg.Settings.AutoRestartOnExit)
	if err != nil {
		return err
	}
	cmd := buildSchedulerCommand(scriptPath, exePath, cfg.Settings.AutoRestartOnExit)
	out, err := exec.Command("schtasks", "/Change", "/TN", schedulerTaskName, "/TR", cmd).CombinedOutput()
	if err != nil {
		return fmt.Errorf("schtasks change: %v: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func ensureWatchdogScript(exePath string, restartOnExit bool) (string, error) {
	dir := filepath.Dir(exePath)
	scriptPath := filepath.Join(dir, schedulerScriptName)
	content := buildWatchdogScript(exePath, restartOnExit)
	if existing, err := os.ReadFile(scriptPath); err == nil {
		if bytes.Equal(existing, []byte(content)) {
			return scriptPath, nil
		}
	}
	if err := os.WriteFile(scriptPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("write watchdog script: %w", err)
	}
	return scriptPath, nil
}

func buildWatchdogScript(exePath string, restartOnExit bool) string {
	safeExe := strings.ReplaceAll(exePath, "'", "''")
	restartFlag := "0"
	if restartOnExit {
		restartFlag = "1"
	}
	return fmt.Sprintf(`param(
    [string]$ExePath = '%s',
    [int]$RestartOnExit = %s
)
$ErrorActionPreference = 'Stop'
if (-not (Test-Path -LiteralPath $ExePath)) { exit 2 }

if ($RestartOnExit -ne 1) {
    $p = Start-Process -FilePath $ExePath -PassThru
    try { $p.WaitForExit() } catch { }
    exit 0
}

while ($true) {
    $existing = Get-Process -ErrorAction SilentlyContinue | Where-Object { $_.Path -eq $ExePath } | Select-Object -First 1
    if ($null -ne $existing) {
        try { $existing.WaitForExit() } catch { Start-Sleep -Seconds 1 }
        Start-Sleep -Seconds 1
        continue
    }
    $p = Start-Process -FilePath $ExePath -PassThru
    try { $p.WaitForExit() } catch { Start-Sleep -Seconds 1 }
    Start-Sleep -Seconds 1
}
`, safeExe, restartFlag)
}

func buildSchedulerCommand(scriptPath, exePath string, restartOnExit bool) string {
	quotedScript := fmt.Sprintf("\"%s\"", scriptPath)
	quotedExe := fmt.Sprintf("\"%s\"", exePath)
	flag := "0"
	if restartOnExit {
		flag = "1"
	}
	return fmt.Sprintf("powershell -NoProfile -ExecutionPolicy Bypass -File %s -ExePath %s -RestartOnExit %s", quotedScript, quotedExe, flag)
}

func queryScheduler() (schedulerQuery, error) {
	ps := fmt.Sprintf(`$ErrorActionPreference = 'Stop'
try {
  $task = Get-ScheduledTask -TaskName '%s'
  $info = Get-ScheduledTaskInfo -TaskName '%s'
  $obj = [pscustomobject]@{
    Exists = $true
    State = $task.State.ToString()
    LastRunTime = if ($info.LastRunTime) { $info.LastRunTime.ToString('yyyy-MM-dd HH:mm:ss') } else { '' }
    LastTaskResult = $info.LastTaskResult
  }
} catch {
  $obj = [pscustomobject]@{
    Exists = $false
    State = ''
    LastRunTime = ''
    LastTaskResult = 0
  }
}
$obj | ConvertTo-Json -Compress
`, schedulerTaskName, schedulerTaskName)

	out, err := exec.Command("powershell", "-NoProfile", "-Command", ps).Output()
	if err != nil {
		return schedulerQuery{}, fmt.Errorf("query scheduler: %w", err)
	}
	var info schedulerQuery
	if err := json.Unmarshal(bytes.TrimSpace(out), &info); err != nil {
		return schedulerQuery{}, fmt.Errorf("parse scheduler status: %w", err)
	}
	return info, nil
}
