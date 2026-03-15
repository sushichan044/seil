package runner

// HookStatus represents the outcome of a hook execution.
type HookStatus string

const (
	HookStatusSuccess HookStatus = "success"
	HookStatusFailure HookStatus = "failure"
	HookStatusSkipped HookStatus = "skipped"
)

// HookResult holds the result of a single hook execution.
type HookResult struct {
	Name     string     `json:"name"`
	Status   HookStatus `json:"status"`
	ExitCode int        `json:"exit_code"`
	LogPath  string     `json:"log_path"`
	Summary  string     `json:"summary"`
}
