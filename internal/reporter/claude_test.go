package reporter_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/seil/internal/reporter"
)

func TestClaudeReporter_Report(t *testing.T) {
	var stdout strings.Builder
	var stderr strings.Builder

	exitCode, err := reporter.ClaudeReporter{}.Report(sampleResults(), &stdout, &stderr)

	require.NoError(t, err)
	assert.Equal(t, 2, exitCode)
	assert.Equal(t, "1 succeeded, 1 failed, 1 skipped\n", stdout.String())
	assert.Contains(t, stderr.String(), "hook: fail")
	assert.NotContains(t, stderr.String(), "hook: ok")
}
