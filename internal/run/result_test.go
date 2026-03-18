package run_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/seil/internal/run"
)

func TestSuccess(t *testing.T) {
	r := run.Success("myjob", "/tmp/myjob.log")
	assert.Equal(t, "myjob", r.Name)
	assert.Equal(t, run.StatusSuccess, r.Status)
	assert.Equal(t, "/tmp/myjob.log", r.LogFile)
}

func TestFailure(t *testing.T) {
	err := errors.New("exit status 1")
	r := run.Failure("myjob", "/tmp/myjob.log", err)
	assert.Equal(t, "myjob", r.Name)
	assert.Equal(t, run.StatusFailure, r.Status)
	assert.Equal(t, "/tmp/myjob.log", r.LogFile)
}

func TestSkipped(t *testing.T) {
	reason := run.SkipReason{Code: run.SkipReasonGlobNoMatch, Message: "glob pattern did not match"}
	r := run.Skipped("myjob", reason)
	assert.Equal(t, "myjob", r.Name)
	assert.Equal(t, run.StatusSkipped, r.Status)
	assert.Empty(t, r.LogFile)
	require.NotNil(t, r.SkipReason)
	assert.Equal(t, run.SkipReasonGlobNoMatch, r.SkipReason.Code)
	assert.Equal(t, "glob pattern did not match", r.SkipReason.Message)
}
