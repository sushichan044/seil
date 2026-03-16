package run

// Status represents the outcome of a job execution.
type Status string

const (
	StatusSuccess Status = "success"
	StatusFailure Status = "failure"
	StatusSkipped Status = "skipped"
)

// Result holds the outcome of a single job run.
type Result struct {
	Name    string `json:"name"`
	Status  Status `json:"status"`
	LogFile string `json:"log_file"`
	err     error
}

func Success(name, logFile string) Result {
	return Result{Name: name, Status: StatusSuccess, LogFile: logFile}
}

func Failure(name, logFile string, err error) Result {
	return Result{Name: name, Status: StatusFailure, LogFile: logFile, err: err}
}

func Skipped(name string) Result {
	return Result{Name: name, Status: StatusSkipped}
}
