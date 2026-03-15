package main_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var seilBin string //nolint:gochecknoglobals // shared across tests via TestMain

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "seil-test-bin-*")
	if err != nil {
		panic("failed to create temp dir: " + err.Error())
	}

	seilBin = filepath.Join(dir, "seil")
	out, err := exec.Command("go", "build", "-o", seilBin, ".").CombinedOutput()
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

func runSeil(t *testing.T, dir string, args ...string) runResult {
	t.Helper()
	cmd := exec.Command(seilBin, args...)
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
			t.Fatalf("unexpected error running seil: %v", err)
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

type groupedResultsJSON struct {
	Failure []hookResultJSON `json:"failure"`
	Success []hookResultJSON `json:"success"`
	Skipped []hookResultJSON `json:"skipped"`
}

func writeSeilYML(t *testing.T, dir, hookName, command string) {
	t.Helper()
	content := fmt.Sprintf(`
post_edit:
  jobs:
    - name: %s
      glob: '**/*.go'
      run: '%s'
`, hookName, command)
	err := os.WriteFile(filepath.Join(dir, "seil.yml"), []byte(content), 0o644)
	require.NoError(t, err)
}

func writeSetupSeilYML(t *testing.T, dir, hookName, command string) {
	t.Helper()
	content := fmt.Sprintf(`
setup:
  jobs:
    - name: %s
      run: '%s'
`, hookName, command)
	err := os.WriteFile(filepath.Join(dir, "seil.yml"), []byte(content), 0o644)
	require.NoError(t, err)
}

func writeTeardownSeilYML(t *testing.T, dir, hookName, command string) {
	t.Helper()
	content := fmt.Sprintf(`
teardown:
  jobs:
    - name: %s
      run: '%s'
`, hookName, command)
	err := os.WriteFile(filepath.Join(dir, "seil.yml"), []byte(content), 0o644)
	require.NoError(t, err)
}

// TestPostEdit_JSON_Schema verifies that --json outputs a valid JSON array with the correct schema.
func TestPostEdit_JSON_Schema(t *testing.T) {
	tmpDir := t.TempDir()

	writeSeilYML(t, tmpDir, "greet", "echo hello")

	targetFile := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(targetFile, []byte("package main\n"), 0o644)
	require.NoError(t, err)

	cfgPath := filepath.Join(tmpDir, "seil.yml")
	result := runSeil(t, "", "-c", cfgPath, "post-edit", "--json", targetFile)

	assert.Equal(t, 0, result.ExitCode, "stderr: %s", result.Stderr)

	var results groupedResultsJSON
	err = json.Unmarshal([]byte(result.Stdout), &results)
	require.NoError(t, err, "stdout should be valid JSON: %s", result.Stdout)
	require.Len(t, results.Success, 1)

	assert.Equal(t, "greet", results.Success[0].Name)
	assert.Equal(t, "success", results.Success[0].Status)
	assert.Equal(t, 0, results.Success[0].ExitCode)
	assert.NotEmpty(t, results.Success[0].LogPath)
	assert.Equal(t, "hello", results.Success[0].Summary)
}

// TestPostEdit_TextFormat_IsDefault verifies that without --json the output is human-readable text.
func TestPostEdit_TextFormat_IsDefault(t *testing.T) {
	tmpDir := t.TempDir()

	writeSeilYML(t, tmpDir, "greet", "echo hello")

	targetFile := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(targetFile, []byte("package main\n"), 0o644)
	require.NoError(t, err)

	cfgPath := filepath.Join(tmpDir, "seil.yml")
	result := runSeil(t, "", "-c", cfgPath, "post-edit", targetFile)

	assert.Equal(t, 0, result.ExitCode, "stderr: %s", result.Stderr)
	assert.False(t, strings.HasPrefix(result.Stdout, "["), "stdout should not be a JSON array: %s", result.Stdout)
	assert.Contains(t, result.Stdout, "=== post-edit hooks result ===")
}

// TestConfig_ExplicitPath verifies that -c <path> loads config from outside the working directory.
func TestConfig_ExplicitPath(t *testing.T) {
	configDir := t.TempDir()
	workDir := t.TempDir()

	writeSeilYML(t, configDir, "lint", "echo linted")

	targetFile := filepath.Join(workDir, "app.go")
	err := os.WriteFile(targetFile, []byte("package main\n"), 0o644)
	require.NoError(t, err)

	cfgPath := filepath.Join(configDir, "seil.yml")
	result := runSeil(t, workDir, "-c", cfgPath, "post-edit", "--json", targetFile)

	assert.Equal(t, 0, result.ExitCode, "stderr: %s", result.Stderr)

	var results groupedResultsJSON
	err = json.Unmarshal([]byte(result.Stdout), &results)
	require.NoError(t, err, "stdout should be valid JSON: %s", result.Stdout)
	require.NotEmpty(t, results.Success)

	assert.Equal(t, "lint", results.Success[0].Name)
}

// TestSetup_JSON_Schema verifies that setup --json outputs a valid JSON array.
func TestSetup_JSON_Schema(t *testing.T) {
	tmpDir := t.TempDir()

	writeSetupSeilYML(t, tmpDir, "greet", "echo hello")

	cfgPath := filepath.Join(tmpDir, "seil.yml")
	result := runSeil(t, "", "-c", cfgPath, "setup", "--json")

	assert.Equal(t, 0, result.ExitCode, "stderr: %s", result.Stderr)

	var results groupedResultsJSON
	err := json.Unmarshal([]byte(result.Stdout), &results)
	require.NoError(t, err, "stdout should be valid JSON: %s", result.Stdout)
	require.Len(t, results.Success, 1)

	assert.Equal(t, "greet", results.Success[0].Name)
	assert.Equal(t, "success", results.Success[0].Status)
	assert.Equal(t, 0, results.Success[0].ExitCode)
	assert.Equal(t, "hello", results.Success[0].Summary)
}

// TestTeardown_JSON_Schema verifies that teardown --json outputs a valid JSON array.
func TestTeardown_JSON_Schema(t *testing.T) {
	tmpDir := t.TempDir()

	writeTeardownSeilYML(t, tmpDir, "cleanup", "echo cleaned")

	cfgPath := filepath.Join(tmpDir, "seil.yml")
	result := runSeil(t, "", "-c", cfgPath, "teardown", "--json")

	assert.Equal(t, 0, result.ExitCode, "stderr: %s", result.Stderr)

	var results groupedResultsJSON
	err := json.Unmarshal([]byte(result.Stdout), &results)
	require.NoError(t, err, "stdout should be valid JSON: %s", result.Stdout)
	require.Len(t, results.Success, 1)

	assert.Equal(t, "cleanup", results.Success[0].Name)
	assert.Equal(t, "success", results.Success[0].Status)
	assert.Equal(t, 0, results.Success[0].ExitCode)
	assert.Equal(t, "cleaned", results.Success[0].Summary)
}

// TestSetup_TextFormat_IsDefault verifies that setup without --json outputs human-readable text.
func TestSetup_TextFormat_IsDefault(t *testing.T) {
	tmpDir := t.TempDir()

	writeSetupSeilYML(t, tmpDir, "greet", "echo hello")

	cfgPath := filepath.Join(tmpDir, "seil.yml")
	result := runSeil(t, "", "-c", cfgPath, "setup")

	assert.Equal(t, 0, result.ExitCode, "stderr: %s", result.Stderr)
	assert.False(t, strings.HasPrefix(result.Stdout, "["), "stdout should not be a JSON array: %s", result.Stdout)
	assert.Contains(t, result.Stdout, "=== setup hooks result ===")
}

// TestConfig_AutoDiscovery verifies that without -c the config is discovered from the working directory.
func TestConfig_AutoDiscovery(t *testing.T) {
	tmpDir := t.TempDir()

	// Create .git dir so git.FindRepoRootFrom recognizes tmpDir as a repo root.
	err := os.Mkdir(filepath.Join(tmpDir, ".git"), 0o755)
	require.NoError(t, err)

	writeSeilYML(t, tmpDir, "fmt", "echo formatted")

	targetFile := filepath.Join(tmpDir, "main.go")
	err = os.WriteFile(targetFile, []byte("package main\n"), 0o644)
	require.NoError(t, err)

	result := runSeil(t, tmpDir, "post-edit", "--json", targetFile)

	assert.Equal(t, 0, result.ExitCode, "stderr: %s", result.Stderr)

	var results groupedResultsJSON
	err = json.Unmarshal([]byte(result.Stdout), &results)
	require.NoError(t, err, "stdout should be valid JSON: %s", result.Stdout)
	require.NotEmpty(t, results.Success)

	assert.Equal(t, "fmt", results.Success[0].Name)
}

// TestPostEdit_JSON_Failure_ExitCode verifies that --json exits with code 1 when a hook fails.
func TestPostEdit_JSON_Failure_ExitCode(t *testing.T) {
	tmpDir := t.TempDir()

	writeSeilYML(t, tmpDir, "error", "exit 1")

	targetFile := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(targetFile, []byte("package main\n"), 0o644)
	require.NoError(t, err)

	cfgPath := filepath.Join(tmpDir, "seil.yml")
	result := runSeil(t, "", "-c", cfgPath, "post-edit", "--json", targetFile)

	assert.Equal(t, 1, result.ExitCode)

	var results groupedResultsJSON
	err = json.Unmarshal([]byte(result.Stdout), &results)
	require.NoError(t, err, "stdout should be valid JSON: %s", result.Stdout)
	require.Len(t, results.Failure, 1)

	assert.Equal(t, "error", results.Failure[0].Name)
	assert.Equal(t, "failure", results.Failure[0].Status)
}
