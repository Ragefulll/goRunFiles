package config

import (
	"os"
	"strings"
)

// RepairFile tries to fix common Windows path escaping issues by quoting values.
// Returns true if file was modified.
func RepairFile(path string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	lines := strings.Split(string(data), "\n")
	changed := false

	for i, line := range lines {
		trim := strings.TrimSpace(line)
		if trim == "" || strings.HasPrefix(trim, ";") || strings.HasPrefix(trim, "#") || strings.HasPrefix(trim, "[") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if val == "" {
			continue
		}
		// Quote values for known keys if they include backslashes/spaces/commas.
		if key == "path" || key == "process" || key == "command" || key == "checkProcess" ||
			key == "checkCmdline" || key == "args" || key == "errorWindowTitles" {
			quoted := val
			if strings.HasPrefix(quoted, `"`) && strings.HasSuffix(quoted, `"`) {
				inner := strings.TrimSuffix(strings.TrimPrefix(quoted, `"`), `"`)
				inner = escapeBackslashes(inner)
				quoted = `"` + strings.ReplaceAll(inner, `"`, `\"`) + `"`
				if quoted != val {
					lines[i] = parts[0] + "=" + quoted
					changed = true
				}
				continue
			}
			if strings.ContainsAny(val, `\ ,`) {
				val = escapeBackslashes(val)
				val = `"` + strings.ReplaceAll(val, `"`, `\"`) + `"`
				lines[i] = parts[0] + "=" + val
				changed = true
			}
		}
	}

	if !changed {
		return false, nil
	}
	return true, os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
}

func escapeBackslashes(s string) string {
	var b strings.Builder
	b.Grow(len(s) * 2)
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' {
			if i+1 < len(s) && s[i+1] == '\\' {
				b.WriteByte('\\')
				b.WriteByte('\\')
				i++
				continue
			}
			b.WriteByte('\\')
			b.WriteByte('\\')
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}
