package main_test

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var himoBin string //nolint:gochecknoglobals // shared across tests via TestMain

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "himo-test-bin-*")
	if err != nil {
		panic("failed to create temp dir: " + err.Error())
	}

	himoBin = filepath.Join(dir, "himo")
	out, err := exec.Command("go", "build", "-o", himoBin, ".").CombinedOutput()
	if err != nil {
		os.RemoveAll(dir)
		panic("go build failed:\n" + string(out))
	}

	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

type runResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func runHimo(t *testing.T, dir string, args ...string) runResult {
	t.Helper()
	cmd := exec.Command(himoBin, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("unexpected error running himo: %v", err)
		}
	}
	return runResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
	}
}

type hookResultJSON struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	ExitCode int    `json:"exit_code"`
	LogPath  string `json:"log_path"`
	Summary  string `json:"summary"`
}

func writeHimoYML(t *testing.T, dir, hookName, command string) {
	t.Helper()
	content := "post_edit:\n  hooks:\n    " + hookName + ":\n      glob: '**/*.go'\n      command: '" + command + "'\n"
	err := os.WriteFile(filepath.Join(dir, "himo.yml"), []byte(content), 0o644)
	require.NoError(t, err)
}

// TestPostEdit_JSON_Schema verifies that --json outputs a valid JSON array with the correct schema.
func TestPostEdit_JSON_Schema(t *testing.T) {
	tmpDir := t.TempDir()

	writeHimoYML(t, tmpDir, "greet", "echo hello")

	targetFile := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(targetFile, []byte("package main\n"), 0o644)
	require.NoError(t, err)

	cfgPath := filepath.Join(tmpDir, "himo.yml")
	result := runHimo(t, "", "-c", cfgPath, "post-edit", "--json", targetFile)

	assert.Equal(t, 0, result.ExitCode, "stderr: %s", result.Stderr)

	var results []hookResultJSON
	err = json.Unmarshal([]byte(result.Stdout), &results)
	require.NoError(t, err, "stdout should be valid JSON: %s", result.Stdout)
	require.Len(t, results, 1)

	assert.Equal(t, "greet", results[0].Name)
	assert.Equal(t, "success", results[0].Status)
	assert.Equal(t, 0, results[0].ExitCode)
	assert.NotEmpty(t, results[0].LogPath)
	assert.Equal(t, "hello", results[0].Summary)
}

// TestPostEdit_TextFormat_IsDefault verifies that without --json the output is human-readable text.
func TestPostEdit_TextFormat_IsDefault(t *testing.T) {
	tmpDir := t.TempDir()

	writeHimoYML(t, tmpDir, "greet", "echo hello")

	targetFile := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(targetFile, []byte("package main\n"), 0o644)
	require.NoError(t, err)

	cfgPath := filepath.Join(tmpDir, "himo.yml")
	result := runHimo(t, "", "-c", cfgPath, "post-edit", targetFile)

	assert.Equal(t, 0, result.ExitCode, "stderr: %s", result.Stderr)
	assert.False(t, strings.HasPrefix(result.Stdout, "["), "stdout should not be a JSON array: %s", result.Stdout)
	assert.Contains(t, result.Stdout, "=== post-edit hooks result ===")
}

// TestConfig_ExplicitPath verifies that -c <path> loads config from outside the working directory.
func TestConfig_ExplicitPath(t *testing.T) {
	configDir := t.TempDir()
	workDir := t.TempDir()

	writeHimoYML(t, configDir, "lint", "echo linted")

	targetFile := filepath.Join(workDir, "app.go")
	err := os.WriteFile(targetFile, []byte("package main\n"), 0o644)
	require.NoError(t, err)

	cfgPath := filepath.Join(configDir, "himo.yml")
	result := runHimo(t, workDir, "-c", cfgPath, "post-edit", "--json", targetFile)

	assert.Equal(t, 0, result.ExitCode, "stderr: %s", result.Stderr)

	var results []hookResultJSON
	err = json.Unmarshal([]byte(result.Stdout), &results)
	require.NoError(t, err, "stdout should be valid JSON: %s", result.Stdout)
	require.NotEmpty(t, results)

	assert.Equal(t, "lint", results[0].Name)
}

// TestConfig_AutoDiscovery verifies that without -c the config is discovered from the working directory.
func TestConfig_AutoDiscovery(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .git dir so git.FindRepoRootFrom recognizes tmpDir as a repo root.
	err := os.Mkdir(filepath.Join(tmpDir, ".git"), 0o755)
	require.NoError(t, err)

	writeHimoYML(t, tmpDir, "fmt", "echo formatted")

	targetFile := filepath.Join(tmpDir, "main.go")
	err = os.WriteFile(targetFile, []byte("package main\n"), 0o644)
	require.NoError(t, err)

	result := runHimo(t, tmpDir, "post-edit", "--json", targetFile)

	assert.Equal(t, 0, result.ExitCode, "stderr: %s", result.Stderr)

	var results []hookResultJSON
	err = json.Unmarshal([]byte(result.Stdout), &results)
	require.NoError(t, err, "stdout should be valid JSON: %s", result.Stdout)
	require.NotEmpty(t, results)

	assert.Equal(t, "fmt", results[0].Name)
}
