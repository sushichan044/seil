package version_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sushichan044/seil/internal/version"
)

func TestGet(t *testing.T) {
	v := version.Get()
	assert.NotEmpty(t, v)
}
