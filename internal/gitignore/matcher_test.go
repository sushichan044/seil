package gitignore_test

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/seil/internal/gitignore"
)

func TestNewMatcherFromRoot(t *testing.T) {
	t.Run("ignores files matching patterns in root .gitignore", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		require.NoError(t, afero.WriteFile(fs, "/repo/.gitignore", []byte("*.log\n"), 0o600))

		m, err := gitignore.NewMatcherFromRoot(fs, "/repo")
		require.NoError(t, err)

		assert.True(t, m.IsIgnored("app.log"))
		assert.False(t, m.IsIgnored("main.go"))
	})

	t.Run("nested .gitignore only applies to files under its directory", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		require.NoError(t, afero.WriteFile(fs, "/repo/.gitignore", []byte("*.log\n"), 0o600))
		require.NoError(t, fs.MkdirAll("/repo/sub", 0o700))
		require.NoError(t, afero.WriteFile(fs, "/repo/sub/.gitignore", []byte("*.tmp\n"), 0o600))

		m, err := gitignore.NewMatcherFromRoot(fs, "/repo")
		require.NoError(t, err)

		// Root .gitignore matches everywhere
		assert.True(t, m.IsIgnored("app.log"))
		// sub/.gitignore matches only files under sub/
		assert.True(t, m.IsIgnored("sub/cache.tmp"))
		// sub/.gitignore does NOT match files outside sub/
		assert.False(t, m.IsIgnored("cache.tmp"))
		assert.False(t, m.IsIgnored("main.go"))
	})

	t.Run("returns matcher with no patterns when no .gitignore exists", func(t *testing.T) {
		fs := afero.NewMemMapFs()
		require.NoError(t, fs.MkdirAll("/repo", 0o700))

		m, err := gitignore.NewMatcherFromRoot(fs, "/repo")
		require.NoError(t, err)

		assert.False(t, m.IsIgnored("anything.go"))
	})
}
