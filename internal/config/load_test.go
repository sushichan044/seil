package config_test

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/seil/internal/config"
)

func TestLoad(t *testing.T) {
	t.Run("new empty config resolves RootDir from root", func(t *testing.T) {
		resolved := config.NewEmpty("/project")

		assert.Equal(t, "/project", resolved.RootDir())
		assert.Equal(t, "/project/.seil.yml", resolved.ConfigPath())
		assert.Empty(t, resolved.Config.Setup.Jobs)
		assert.Empty(t, resolved.Config.Teardown.Jobs)
		assert.Empty(t, resolved.Config.PostEdit.Jobs)
	})

	t.Run("loads valid config file and resolves CWD", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		path := "/project/.seil.yml"
		content := []byte(`
setup:
  jobs:
    - run: echo hello
`)
		require.NoError(t, afero.WriteFile(fs, path, content, 0o600))

		resolved, err := config.Load(fs, path)
		require.NoError(t, err)

		assert.Equal(t, "/project", resolved.RootDir())
		assert.Equal(t, "/project/.seil.yml", resolved.ConfigPath())
		require.Len(t, resolved.Config.Setup.Jobs, 1)
		assert.Equal(t, "echo hello", resolved.Config.Setup.Jobs[0].Run)
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		_, err := config.Load(fs, "/nonexistent/path/.seil.yml")
		assert.Error(t, err)
	})

	t.Run("returns error for invalid YAML", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		path := "/project/.seil.yml"
		require.NoError(t, afero.WriteFile(fs, path, []byte("{bad yaml"), 0o600))

		_, err := config.Load(fs, path)
		assert.Error(t, err)
	})

	t.Run("returns error when validation fails", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		path := "/project/.seil.yml"
		content := []byte(`setup:
  jobs:
	- run: echo hello
	  glob: "**/*.js" # glob is not allowed for setup jobs
`)
		require.NoError(t, afero.WriteFile(fs, path, content, 0o600))

		_, err := config.Load(fs, path)
		assert.Error(t, err)
	})
}
