package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	updateVersionURL = "https://feed.art3d.ru/gorunfiles/version.txt"
	updateExeURL     = "https://feed.art3d.ru/gorunfiles/goRunFiles.exe"
	selfUpdateFlag   = "--self-update"
)

type UpdateStatus struct {
	Current  string  `json:"current"`
	Remote   string  `json:"remote"`
	Status   string  `json:"status"`
	Progress float64 `json:"progress"`
	Detail   string  `json:"detail"`
}

type updater struct {
	mu       sync.Mutex
	update   UpdateStatus
	checking bool
	ctx      context.Context
}

func (u *updater) updateNewExePath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(exe), filepath.Base(exe)+".new"), nil
}

func (u *updater) updateAppliedMarker() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return exe + ".updated", nil
}

func isSelfUpdate() bool {
	return len(os.Args) > 1 && os.Args[1] == selfUpdateFlag
}

func (u *updater) checkForUpdate() {
	if marker, err := u.updateAppliedMarker(); err == nil {
		if _, statErr := os.Stat(marker); statErr == nil {
			_ = os.Remove(marker)
			u.removeStaleUpdate()
			u.emit(UpdateStatus{Current: buildVersion, Status: "applied", Detail: "Обновление применено"})
			return
		}
	}
	u.emit(UpdateStatus{Current: buildVersion, Status: "check", Detail: "Проверка обновлений..."})
	remote, err := fetchRemoteVersion()
	if err != nil {
		u.emit(UpdateStatus{Current: buildVersion, Status: "error", Detail: "Не удалось проверить обновления: " + err.Error()})
		return
	}
	if _, ok := parseVersion(remote); !ok {
		u.emit(UpdateStatus{Current: buildVersion, Remote: remote, Status: "error", Detail: "Сервер вернул некорректную версию"})
		return
	}
	if !versionLess(buildVersion, remote) {
		u.emit(UpdateStatus{Current: buildVersion, Remote: remote, Status: "idle", Detail: "Версия актуальна"})
		return
	}
	u.removeStaleUpdate()
	newPath, err := u.updateNewExePath()
	if err != nil {
		u.emit(UpdateStatus{Current: buildVersion, Remote: remote, Status: "failed", Detail: err.Error()})
		return
	}
	u.emit(UpdateStatus{Current: buildVersion, Remote: remote, Status: "downloading", Detail: "Скачивание обновления..."})
	if err := downloadToFile(updateExeURL, newPath, func(done, total int64) {
		pct := 0.0
		if total > 0 {
			pct = float64(done) / float64(total) * 100
		}
		u.emit(UpdateStatus{Current: buildVersion, Remote: remote, Status: "downloading", Progress: pct,
			Detail: fmt.Sprintf("%.1f / %.1f МБ", float64(done)/1024/1024, float64(total)/1024/1024)})
	}); err != nil {
		_ = os.Remove(newPath)
		u.emit(UpdateStatus{Current: buildVersion, Remote: remote, Status: "failed", Detail: err.Error()})
		return
	}
	if !isExecutableFile(newPath) {
		_ = os.Remove(newPath)
		u.emit(UpdateStatus{Current: buildVersion, Remote: remote, Status: "failed", Detail: "Скачанный файл повреждён, обновление отменено"})
		return
	}
	u.emit(UpdateStatus{Current: buildVersion, Remote: remote, Status: "restarting", Detail: "Применение обновления..."})
	if err := launchUpdater(newPath); err != nil {
		u.emit(UpdateStatus{Current: buildVersion, Remote: remote, Status: "failed", Detail: err.Error()})
		return
	}
	time.Sleep(1200 * time.Millisecond)
	os.Exit(0)
}

func (u *updater) emit(s UpdateStatus) {
	u.mu.Lock()
	u.update = s
	u.mu.Unlock()
	if u.ctx != nil {
		wruntime.EventsEmit(u.ctx, "update-status", s)
	}
}

// GetUpdateStatus returns the latest update status.
func (u *updater) GetUpdateStatus() UpdateStatus {
	u.mu.Lock()
	defer u.mu.Unlock()
	return u.update
}

// AppVersion returns the current application version.
func (u *updater) AppVersion() string {
	return buildVersion
}

// CheckUpdates starts an asynchronous update check in the background.
func (u *updater) CheckUpdates() UpdateStatus {
	u.mu.Lock()
	if u.checking {
		st := u.update
		u.mu.Unlock()
		return st
	}
	u.checking = true
	st := u.update
	u.mu.Unlock()
	go u.checkForUpdate()
	return st
}

func fetchRemoteVersion() (string, error) {
	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Get(updateVersionURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(body)), nil
}

func versionLess(a, b string) bool {
	an, aok := parseVersion(a)
	bn, bok := parseVersion(b)
	if aok && bok {
		n := len(an)
		if len(bn) > n {
			n = len(bn)
		}
		for i := 0; i < n; i++ {
			x, y := 0, 0
			if i < len(an) {
				x = an[i]
			}
			if i < len(bn) {
				y = bn[i]
			}
			if x != y {
				return x < y
			}
		}
		return false
	}
	return strings.TrimSpace(a) < strings.TrimSpace(b)
}

func parseVersion(v string) ([]int, bool) {
	v = strings.TrimPrefix(strings.TrimSpace(v), "v")
	parts := strings.Split(v, ".")
	out := make([]int, 0, len(parts))
	for _, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return nil, false
		}
		out = append(out, n)
	}
	return out, len(out) > 0
}

func downloadToFile(url, dest string, progress func(done, total int64)) error {
	client := &http.Client{Timeout: 30 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("не удалось начать загрузку: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ошибка загрузки: HTTP %d", resp.StatusCode)
	}
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	buf := make([]byte, 256*1024)
	var written int64
	total := resp.ContentLength
	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, err := out.Write(buf[:n]); err != nil {
				return err
			}
			written += int64(n)
			if progress != nil {
				progress(written, total)
			}
		}
		if readErr == io.EOF {
			return nil
		}
		if readErr != nil {
			return readErr
		}
	}
}

func launchUpdater(newPath string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	return hideCommandWindow(exec.Command(newPath, selfUpdateFlag, exe)).Start()
}

func runSelfUpdate() error {
	if len(os.Args) < 3 {
		return errors.New("неверные аргументы обновления")
	}
	dest := os.Args[2]
	if err := waitWritable(dest); err != nil {
		return err
	}
	self, err := os.Executable()
	if err != nil {
		return err
	}
	deadline := time.Now().Add(2 * time.Minute)
	for {
		if err := copyFile(self, dest); err == nil {
			break
		} else if time.Now().After(deadline) {
			return fmt.Errorf("не удалось заменить %s: %w", dest, err)
		}
		time.Sleep(500 * time.Millisecond)
	}
	_ = os.WriteFile(dest+".updated", []byte(buildVersion), 0644)
	return startUpdateProcess(dest)
}

func startUpdateProcess(dest string) error {
	deadline := time.Now().Add(30 * time.Second)
	for {
		err := exec.Command(dest).Start()
		if err == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("не удалось запустить %s: %w", dest, err)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func waitWritable(dest string) error {
	deadline := time.Now().Add(2 * time.Minute)
	for {
		f, err := os.OpenFile(dest, os.O_WRONLY, 0)
		if err == nil {
			f.Close()
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("не удалось дождаться освобождения %s: %w", dest, err)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func isExecutableFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	head := make([]byte, 2)
	if _, err := io.ReadFull(f, head); err != nil {
		return false
	}
	return head[0] == 'M' && head[1] == 'Z'
}

func (u *updater) removeStaleUpdate() {
	if p, err := u.updateNewExePath(); err == nil {
		_ = os.Remove(p)
	}
}
