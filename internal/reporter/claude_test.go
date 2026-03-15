package reporter_test

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/exp/golden"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/seil/internal/reporter"
)

func TestClaudeReporter_ReportStdout(t *testing.T) {
	var stdout strings.Builder
	var stderr strings.Builder

	exitCode, err := reporter.ClaudeReporter{}.Report(sampleResults(), &stdout, &stderr)

	require.NoError(t, err)
	assert.Equal(t, 2, exitCode)
	golden.RequireEqual(t, []byte(stdout.String()))
}

func TestClaudeReporter_ReportStderr(t *testing.T) {
	var stdout strings.Builder
	var stderr strings.Builder

	exitCode, err := reporter.ClaudeReporter{}.Report(sampleResults(), &stdout, &stderr)

	require.NoError(t, err)
	assert.Equal(t, 2, exitCode)
	golden.RequireEqual(t, []byte(stderr.String()))
}
