package app

// Status represents the runtime state we display for a process.
type Status string

const (
	StatusUnknown Status = "unknown"
	StatusRunning Status = "running"
	StatusStarted Status = "started"
	StatusStopped Status = "stopped"
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
	default:
		return "☠︎ UNKNOWN"
	}
}
