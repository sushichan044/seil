package reporter

import (
	"fmt"
	"io"

	"github.com/sushichan044/seil/internal/runner"
)

type ClaudeReporter struct{}

const claudeFailureExitCode = 2

func (ClaudeReporter) Report(results []runner.HookResult, stdout io.Writer, stderr io.Writer) (int, error) {
	grouped := groupResults(results)
	for _, result := range grouped.Failure {
		if err := writeHookResult(stderr, result); err != nil {
			return 0, err
		}
	}
	if _, err := fmt.Fprintf(stdout, "%s\n", summaryLine(grouped)); err != nil {
		return 0, err
	}
	return claudeExitCode(grouped), nil
}

func claudeExitCode(grouped groupedHookResults) int {
	if len(grouped.Failure) > 0 {
		return claudeFailureExitCode
	}
	return 0
}
