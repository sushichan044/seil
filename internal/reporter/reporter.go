package reporter

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/sushichan044/seil/internal/agent"
	"github.com/sushichan044/seil/internal/run"
)

type Reporter interface {
	Report(results []run.Result, stdout io.Writer, stderr io.Writer) (int, error)
}

type Name string

const (
	ReporterAuto    Name = "auto"
	ReporterDefault Name = "default"
	ReporterJSON    Name = "json"
	ReporterClaude  Name = "claude"
)

//nolint:gochecknoglobals // ReporterNames is a list of all valid reporter names.
var ReporterNames = []string{
	string(ReporterAuto),
	string(ReporterDefault),
	string(ReporterJSON),
	string(ReporterClaude),
}

type HumanReporter struct{}

type JSONReporter struct{}

type groupedResults struct {
	Failure []run.Result `json:"failure"`
	Success []run.Result `json:"success"`
	Skipped []run.Result `json:"skipped"`
}

func ParseName(raw string) Name {
	normalized := Name(strings.ToLower(strings.TrimSpace(raw)))
	switch normalized {
	case ReporterAuto, ReporterDefault, ReporterJSON, ReporterClaude:
		return normalized
	default:
		return ""
	}
}

// UnmarshalText implements the [encoding.TextUnmarshaler] interface.
//
// This is called from kong when parsing CLI flags.
func (n *Name) UnmarshalText(text []byte) error {
	parsed := ParseName(string(text))
	if parsed == "" {
		return fmt.Errorf("invalid reporter %q", string(text))
	}
	*n = parsed
	return nil
}

func Resolve(name Name, detectedAgent agent.Agent) Reporter {
	switch name {
	case ReporterJSON:
		return JSONReporter{}
	case ReporterDefault:
		return HumanReporter{}
	case ReporterClaude:
		return ClaudeReporter{}
	case ReporterAuto, "":
		switch detectedAgent {
		case agent.AgentClaude:
			return ClaudeReporter{}
		case agent.AgentUnknown,
			agent.AgentCursor,
			agent.AgentDevin,
			agent.AgentReplit,
			agent.AgentGemini,
			agent.AgentCodex,
			agent.AgentAuggie,
			agent.AgentOpenCode,
			agent.AgentKiro,
			agent.AgentGoose,
			agent.AgentPi:
			return HumanReporter{}
		}
	}

	return HumanReporter{}
}

func (HumanReporter) Report(results []run.Result, stdout io.Writer, _ io.Writer) (int, error) {
	grouped := groupResults(results)

	if _, err := fmt.Fprintf(stdout, "--- Failures (%d) ---\n", len(grouped.Failure)); err != nil {
		return 0, err
	}
	for _, result := range grouped.Failure {
		if err := writeResult(stdout, result); err != nil {
			return 0, err
		}
	}

	if _, err := fmt.Fprintf(stdout, "\n--- Successes (%d) ---\n", len(grouped.Success)); err != nil {
		return 0, err
	}
	for _, result := range grouped.Success {
		if err := writeResult(stdout, result); err != nil {
			return 0, err
		}
	}

	if _, err := fmt.Fprintf(stdout, "\n--- Skipped (%d) ---\n", len(grouped.Skipped)); err != nil {
		return 0, err
	}
	for _, result := range grouped.Skipped {
		if err := writeResult(stdout, result); err != nil {
			return 0, err
		}
	}

	if _, err := fmt.Fprintf(stdout, "\n---\n%s\n", summaryLine(grouped)); err != nil {
		return 0, err
	}
	return defaultExitCode(grouped), nil
}

func (JSONReporter) Report(results []run.Result, stdout io.Writer, _ io.Writer) (int, error) {
	grouped := groupResults(results)
	enc := json.NewEncoder(stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(grouped); err != nil {
		return 0, err
	}
	return defaultExitCode(grouped), nil
}

func groupResults(results []run.Result) groupedResults {
	grouped := groupedResults{
		Failure: []run.Result{},
		Success: []run.Result{},
		Skipped: []run.Result{},
	}
	for _, result := range results {
		switch result.Status {
		case run.StatusFailure:
			grouped.Failure = append(grouped.Failure, result)
		case run.StatusSuccess:
			grouped.Success = append(grouped.Success, result)
		case run.StatusSkipped:
			grouped.Skipped = append(grouped.Skipped, result)
		}
	}
	return grouped
}

func writeResult(w io.Writer, result run.Result) error {
	_, err := fmt.Fprintf(w, "\nhook: %s\nstatus: %s\nlog: %s\n",
		result.Name, result.Status, result.LogFile)
	return err
}

func defaultExitCode(grouped groupedResults) int {
	if len(grouped.Failure) > 0 {
		return 1
	}
	return 0
}

func summaryLine(grouped groupedResults) string {
	return fmt.Sprintf("%d succeeded, %d failed, %d skipped",
		len(grouped.Success), len(grouped.Failure), len(grouped.Skipped))
}
