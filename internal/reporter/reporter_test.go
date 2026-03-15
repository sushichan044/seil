package reporter_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/seil/internal/agent"
	"github.com/sushichan044/seil/internal/reporter"
	"github.com/sushichan044/seil/internal/runner"
)

type groupedResultsJSON struct {
	Failure []runner.HookResult `json:"failure"`
	Success []runner.HookResult `json:"success"`
	Skipped []runner.HookResult `json:"skipped"`
}

func TestResolveReporter(t *testing.T) {
	assert.Equal(t, reporter.ReporterAuto, reporter.ParseName("auto"))
	assert.Equal(t, reporter.ReporterDefault, reporter.ParseName("default"))
	assert.Equal(t, reporter.ReporterJSON, reporter.ParseName("json"))
	assert.Equal(t, reporter.Name(""), reporter.ParseName("wat"))
	assert.IsType(t, reporter.JSONReporter{}, reporter.Resolve(reporter.ReporterJSON, agent.AgentClaude))
	assert.IsType(t, reporter.ClaudeReporter{}, reporter.Resolve(reporter.ReporterAuto, agent.AgentClaude))
	assert.IsType(t, reporter.HumanReporter{}, reporter.Resolve(reporter.ReporterAuto, agent.AgentCodex))
	assert.IsType(t, reporter.ClaudeReporter{}, reporter.Resolve("", agent.AgentClaude))
	assert.IsType(t, reporter.HumanReporter{}, reporter.Resolve("", agent.AgentCodex))
	assert.IsType(t, reporter.HumanReporter{}, reporter.Resolve(reporter.ReporterDefault, agent.AgentClaude))
	assert.IsType(t, reporter.ClaudeReporter{}, reporter.Resolve(reporter.ReporterClaude, agent.AgentCodex))
}

func TestHumanReporter_Report(t *testing.T) {
	var stdout strings.Builder

	exitCode, err := reporter.HumanReporter{}.Report(sampleResults(), &stdout, &strings.Builder{})

	require.NoError(t, err)
	assert.Equal(t, 1, exitCode)
	assert.Contains(t, stdout.String(), "--- Failures (1) ---")
	assert.Contains(t, stdout.String(), "--- Successes (1) ---")
	assert.Contains(t, stdout.String(), "--- Skipped (1) ---")
	assert.Contains(t, stdout.String(), "1 succeeded, 1 failed, 1 skipped")
	assert.NotContains(t, stdout.String(), "=== ")
}

func TestJSONReporter_Report(t *testing.T) {
	var stdout strings.Builder

	exitCode, err := reporter.JSONReporter{}.Report(sampleResults(), &stdout, &strings.Builder{})

	require.NoError(t, err)
	assert.Equal(t, 1, exitCode)

	var grouped groupedResultsJSON
	err = json.Unmarshal([]byte(stdout.String()), &grouped)
	require.NoError(t, err)
	require.Len(t, grouped.Failure, 1)
}

func sampleResults() []runner.HookResult {
	return []runner.HookResult{
		{Name: "ok", Status: runner.HookStatusSuccess, ExitCode: 0, LogPath: "/tmp/ok.log", Summary: "hello"},
		{Name: "fail", Status: runner.HookStatusFailure, ExitCode: 1, LogPath: "/tmp/fail.log", Summary: "boom"},
		{Name: "skip", Status: runner.HookStatusSkipped},
	}
}
