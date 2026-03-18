package reporter_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/charmbracelet/x/exp/golden"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/seil/internal/agent"
	"github.com/sushichan044/seil/internal/reporter"
	"github.com/sushichan044/seil/internal/run"
)

type groupedResultsJSON struct {
	Failure []run.Result `json:"failure"`
	Success []run.Result `json:"success"`
	Skipped []run.Result `json:"skipped"`
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
	golden.RequireEqual(t, []byte(stdout.String()))
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

func sampleResults() []run.Result {
	return []run.Result{
		run.Success("ok", "/tmp/ok.log"),
		run.Failure("fail", "/tmp/fail.log", errors.New("exit status 1")),
		run.Skipped("skip", run.SkipReason{
			Code:    run.SkipReasonGlobNoMatch,
			Message: `glob pattern "**/*.ts" did not match`,
		}),
	}
}
