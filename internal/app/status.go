package app

// Status represents the runtime state we display for a process.
type Status string

const (
	StatusUnknown  Status = "unknown"
	StatusRunning  Status = "running"
	StatusStarted  Status = "started"
	StatusStopped  Status = "stopped"
	StatusDisabled Status = "disabled"
)

// Icon returns the user-facing marker for a status.
func (s Status) Icon() string {
	switch s {
	case StatusRunning:
		return "★ WORK    ︎"
	case StatusStarted:
		return "☆︎ RUN     "
	case StatusStopped:
		return "✗︎ NRUN    "
	case StatusDisabled:
		return "⛔︎ DISABLED"
	default:
		return "☠︎ UNKNOWN "
	}
}
