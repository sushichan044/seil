package run_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

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
	r := run.Skipped("myjob")
	assert.Equal(t, "myjob", r.Name)
	assert.Equal(t, run.StatusSkipped, r.Status)
	assert.Empty(t, r.LogFile)
}
