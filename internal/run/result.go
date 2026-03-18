package run

// Status represents the outcome of a job execution.
type Status string

const (
	StatusSuccess Status = "success"
	StatusFailure Status = "failure"
	StatusSkipped Status = "skipped"
)

// SkipReasonCode is a machine-readable code for why a job was skipped.
type SkipReasonCode string

const (
	SkipReasonGlobNoMatch      SkipReasonCode = "glob_no_match"
	SkipReasonGitignored       SkipReasonCode = "gitignored"
	SkipReasonOutsideWorkspace SkipReasonCode = "outside_workspace"
)

// SkipReason holds a machine-readable code and a human-readable message.
type SkipReason struct {
	Code    SkipReasonCode `json:"code"`
	Message string         `json:"message"`
}

// Result holds the outcome of a single job run.
type Result struct {
	Name       string      `json:"name"`
	Status     Status      `json:"status"`
	LogFile    string      `json:"log_file,omitempty"`
	SkipReason *SkipReason `json:"skip_reason,omitempty"`
	err        error
}

func Success(name, logFile string) Result {
	return Result{Name: name, Status: StatusSuccess, LogFile: logFile}
}

func Failure(name, logFile string, err error) Result {
	return Result{Name: name, Status: StatusFailure, LogFile: logFile, err: err}
}

func Skipped(name string, reason SkipReason) Result {
	return Result{Name: name, Status: StatusSkipped, SkipReason: &reason}
}
