package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/seil/internal/config"
)

func TestValidate_LogDir(t *testing.T) {
	t.Run("valid relative path is accepted", func(t *testing.T) {
		cfg := &config.Config{LogDir: ".seil/logs"}
		err := config.Validate(cfg)
		require.NoError(t, err)
	})

	t.Run("absolute path is rejected", func(t *testing.T) {
		cfg := &config.Config{LogDir: "/tmp/logs"}
		err := config.Validate(cfg)
		assert.ErrorContains(t, err, "log_dir: must be a relative path")
	})

	t.Run("path escaping config root is rejected", func(t *testing.T) {
		cfg := &config.Config{LogDir: "../outside"}
		err := config.Validate(cfg)
		assert.ErrorContains(t, err, "log_dir: must not escape config root")
	})

	t.Run("bare dotdot is rejected", func(t *testing.T) {
		cfg := &config.Config{LogDir: ".."}
		err := config.Validate(cfg)
		assert.ErrorContains(t, err, "log_dir: must not escape config root")
	})

	t.Run("empty log_dir is accepted", func(t *testing.T) {
		cfg := &config.Config{LogDir: ""}
		err := config.Validate(cfg)
		require.NoError(t, err)
	})
}
