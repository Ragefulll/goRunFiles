package app

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mattn/go-runewidth"
	"goRunFiles/internal/config"
)

func (a *App) render(statuses []procStatus) {
	enableANSI()

	now := time.Now().Format("2006-01-02 15:04:05")
	clearConsole()
	var b strings.Builder
	b.Grow(1024)

	b.WriteString(config.Banner)
	fmt.Fprintf(&b, "%s Monitor  %s\n", LogTag, now)
	if a.version != "" {
		if ansiEnabled {
			fmt.Fprintf(&b, "Version: \x1b[32m%s\x1b[0m\n", a.version)
		} else {
			fmt.Fprintf(&b, "Version: %s\n", a.version)
		}
	}
	b.WriteString("\n")

	headers := []string{"NAME", "TYPE", "STATUS", "PID", "STARTED", "UPTIME", "TARGET", "ERROR"}
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, s := range statuses {
		cols := []string{
			s.Name,
			s.Type,
			s.Status.Icon(),
			s.pidString(),
			s.StartedAt,
			s.Uptime,
			s.Target,
			s.Err,
		}
		for i, c := range cols {
			if w := runewidth.StringWidth(c); w > widths[i] {
				widths[i] = w
			}
		}
	}

	maxErr := 60
	if widths[7] > maxErr {
		widths[7] = maxErr
	}

	b.WriteString(formatRow(headers, widths))
	b.WriteString("\n")
	b.WriteString(formatRow(dividerRow(widths), widths))
	b.WriteString("\n")

	for _, s := range statuses {
		row := []string{
			s.Name,
			s.Type,
			s.Status.Icon(),
			s.pidString(),
			s.StartedAt,
			s.Uptime,
			s.Target,
			truncateDisplay(s.Err, widths[7]),
		}
		line := formatRowWithColors(row, widths, func(col int, text string) string {
			if !ansiEnabled {
				return text
			}
			if col != 2 {
				return text
			}
			return colorizeStatus(s.Status, text)
		})
		if ansiEnabled && s.Hung {
			line = "\x1b[41m" + line + "\x1b[49m"
		}
		b.WriteString(line)
		b.WriteString("\n")
	}

	frame := b.String()
	lines := strings.Split(frame, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	maxWidth := 0
	for i := range lines {
		w := runewidth.StringWidth(lines[i])
		if w > maxWidth {
			maxWidth = w
		}
	}
	if a.lastRenderWidth > maxWidth {
		maxWidth = a.lastRenderWidth
	}
	a.lastRenderWidth = maxWidth

	for i := range lines {
		lines[i] = padRight(lines[i], maxWidth)
	}

	if !ansiEnabled && a.lastRenderLines > len(lines) {
		for i := 0; i < a.lastRenderLines-len(lines); i++ {
			lines = append(lines, padRight("", maxWidth))
		}
	}
	a.lastRenderLines = len(lines)

	frame = strings.Join(lines, "\n") + "\n"
	_, _ = os.Stdout.WriteString(frame)
}

func formatRow(cols []string, widths []int) string {
	out := ""
	for i, c := range cols {
		if i > 0 {
			out += "  "
		}
		out += padRight(c, widths[i])
	}
	return out
}

func formatRowWithColors(cols []string, widths []int, colorFn func(col int, text string) string) string {
	out := ""
	for i, c := range cols {
		if i > 0 {
			out += "  "
		}
		padded := padRight(c, widths[i])
		if colorFn != nil {
			padded = colorFn(i, padded)
		}
		out += padded
	}
	return out
}

func dividerRow(widths []int) []string {
	out := make([]string, len(widths))
	for i, w := range widths {
		out[i] = strings.Repeat("-", w)
	}
	return out
}

func colorizeStatus(s Status, text string) string {
	switch s {
	case StatusRunning:
		return "\x1b[32m" + text + "\x1b[39m"
	case StatusStarted:
		return "\x1b[33m" + text + "\x1b[39m"
	case StatusStopped, StatusUnknown:
		return "\x1b[31m" + text + "\x1b[39m"
	default:
		return text
	}
}
