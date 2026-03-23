package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/seil/internal/config"
)

func TestParseConfigYAML(t *testing.T) {
	t.Run("valid empty config", func(t *testing.T) {
		data := []byte(`{}`)
		cfg, err := config.ParseConfigYAML(data)
		require.NoError(t, err)
		assert.NotNil(t, cfg)
		assert.Empty(t, cfg.Setup.Jobs)
		assert.Empty(t, cfg.Teardown.Jobs)
		assert.Empty(t, cfg.PostEdit.Jobs)
	})

	t.Run("valid default config", func(t *testing.T) {
		cfg, err := config.GetDefault()
		require.NoError(t, err)
		assert.NotNil(t, cfg)
	})

	t.Run("valid config with all hooks", func(t *testing.T) {
		data := []byte(`
setup:
  jobs:
    - name: install
      run: npm install
teardown:
  jobs:
    - run: echo teardown
post_edit:
  jobs:
    - name: lint
      run: eslint
      glob: "**/*.ts"
`)
		cfg, err := config.ParseConfigYAML(data)
		require.NoError(t, err)

		require.Len(t, cfg.Setup.Jobs, 1)
		assert.Equal(t, "install", cfg.Setup.Jobs[0].Name)
		assert.Equal(t, "npm install", cfg.Setup.Jobs[0].Run)

		require.Len(t, cfg.Teardown.Jobs, 1)
		assert.Equal(t, "echo teardown", cfg.Teardown.Jobs[0].Run)

		require.Len(t, cfg.PostEdit.Jobs, 1)
		assert.Equal(t, "lint", cfg.PostEdit.Jobs[0].Name)
		assert.Equal(t, "eslint", cfg.PostEdit.Jobs[0].Run)
		assert.Equal(t, "**/*.ts", cfg.PostEdit.Jobs[0].Glob)
	})

	t.Run("valid config with multiple jobs", func(t *testing.T) {
		data := []byte(`
setup:
  jobs:
    - run: echo first
    - run: echo second
    - run: echo third
`)
		cfg, err := config.ParseConfigYAML(data)
		require.NoError(t, err)
		assert.Len(t, cfg.Setup.Jobs, 3)
	})

	t.Run("invalid YAML syntax", func(t *testing.T) {
		data := []byte(`{invalid yaml: [`)
		_, err := config.ParseConfigYAML(data)
		assert.Error(t, err)
	})

	t.Run("validation error: empty run field", func(t *testing.T) {
		data := []byte(`
setup:
  jobs:
    - run: ""
`)
		_, err := config.ParseConfigYAML(data)
		assert.Error(t, err)
	})

	t.Run("validation error: missing run field", func(t *testing.T) {
		data := []byte(`
setup:
  jobs:
    - name: no-run
`)
		_, err := config.ParseConfigYAML(data)
		assert.Error(t, err)
	})
}
