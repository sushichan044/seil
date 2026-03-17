package config_test

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/seil/internal/config"
)

func TestFindConfigFile(t *testing.T) {
	t.Run("finds config in ancestor directory without git root", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		require.NoError(t, fs.MkdirAll("/project/nested/deeper", 0o755))
		require.NoError(t, afero.WriteFile(fs, "/project/.seil.yml", []byte("setup: {}\n"), 0o600))

		got, err := config.FindConfigFile(fs, "/project/nested/deeper")
		require.NoError(t, err)
		assert.Equal(t, "/project/.seil.yml", got)
	})

	t.Run("finds config when starting from file path", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		require.NoError(t, fs.MkdirAll("/project/nested", 0o755))
		require.NoError(t, afero.WriteFile(fs, "/project/.seil.yml", []byte("setup: {}\n"), 0o600))
		require.NoError(t, afero.WriteFile(fs, "/project/nested/main.go", []byte("package main\n"), 0o600))

		got, err := config.FindConfigFile(fs, "/project/nested/main.go")
		require.NoError(t, err)
		assert.Equal(t, "/project/.seil.yml", got)
	})

	t.Run("returns error when config does not exist up to fs root", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		require.NoError(t, fs.MkdirAll("/project/nested/deeper", 0o755))

		_, err := config.FindConfigFile(fs, "/project/nested/deeper")
		assert.ErrorIs(t, err, config.ErrConfigNotFound)
	})
}
