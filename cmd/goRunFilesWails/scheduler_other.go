//go:build !windows

package main

import (
	"fmt"

	"goRunFiles/internal/config"
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

func (g *GUI) GetSchedulerStatus() (TaskStatus, error) {
	return TaskStatus{TaskName: "goRunFilesWails_AutoStart", Error: "scheduler supported on Windows only"}, nil
}

func (g *GUI) InstallScheduler() error {
	return fmt.Errorf("scheduler supported on Windows only")
}

func (g *GUI) RemoveScheduler() error {
	return fmt.Errorf("scheduler supported on Windows only")
}

func updateSchedulerScriptIfInstalled(_ config.Config) error {
	return nil
}
