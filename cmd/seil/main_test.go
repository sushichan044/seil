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
	return runSeilWithEnv(t, dir, nil, args...)
}

func runSeilWithEnv(t *testing.T, dir string, env map[string]string, args ...string) runResult {
	t.Helper()
	cmd := exec.Command(seilBin, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Env = filteredEnv(os.Environ())
	for key, value := range env {
		cmd.Env = append(cmd.Env, key+"="+value)
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

func filteredEnv(environ []string) []string {
	excluded := map[string]struct{}{
		"AI_AGENT":        {},
		"CLAUDECODE":      {},
		"CLAUDE_CODE":     {},
		"REPL_ID":         {},
		"GEMINI_CLI":      {},
		"CODEX_SANDBOX":   {},
		"CODEX_THREAD_ID": {},
		"OPENCODE":        {},
		"AUGMENT_AGENT":   {},
		"GOOSE_PROVIDER":  {},
		"CURSOR_AGENT":    {},
	}

	filtered := make([]string, 0, len(environ))
	for _, entry := range environ {
		key, _, found := strings.Cut(entry, "=")
		if !found {
			continue
		}
		if _, ok := excluded[key]; ok {
			continue
		}
		filtered = append(filtered, entry)
	}
	return filtered
}

type hookResultJSON struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	LogFile string `json:"log_file"`
}

type groupedResultsJSON struct {
	Failure []hookResultJSON `json:"failure"`
	Success []hookResultJSON `json:"success"`
	Skipped []hookResultJSON `json:"skipped"`
}

func writeConfigYML(t *testing.T, dir, fileName, hookName, command string) {
	t.Helper()
	content := fmt.Sprintf(`
post_edit:
  jobs:
    - name: %s
      glob: '**/*.go'
      run: '%s'
`, hookName, command)
	err := os.WriteFile(filepath.Join(dir, fileName), []byte(content), 0o644)
	require.NoError(t, err)
}

func writeDefaultConfigYML(t *testing.T, dir, hookName, command string) {
	t.Helper()
	writeConfigYML(t, dir, ".seil.yml", hookName, command)
}

func writeLegacyConfigYML(t *testing.T, dir, hookName, command string) {
	t.Helper()
	writeConfigYML(t, dir, "seil.yml", hookName, command)
}

func writeSetupConfigYML(t *testing.T, dir, fileName, hookName, command string) {
	t.Helper()
	content := fmt.Sprintf(`
setup:
  jobs:
    - name: %s
      run: '%s'
`, hookName, command)
	err := os.WriteFile(filepath.Join(dir, fileName), []byte(content), 0o644)
	require.NoError(t, err)
}

func writeDefaultSetupConfigYML(t *testing.T, dir, hookName, command string) {
	t.Helper()
	writeSetupConfigYML(t, dir, ".seil.yml", hookName, command)
}

func writeTeardownConfigYML(t *testing.T, dir, fileName, hookName, command string) {
	t.Helper()
	content := fmt.Sprintf(`
teardown:
  jobs:
    - name: %s
      run: '%s'
`, hookName, command)
	err := os.WriteFile(filepath.Join(dir, fileName), []byte(content), 0o644)
	require.NoError(t, err)
}

func writeDefaultTeardownConfigYML(t *testing.T, dir, hookName, command string) {
	t.Helper()
	writeTeardownConfigYML(t, dir, ".seil.yml", hookName, command)
}

// TestPostEdit_JSON_Schema verifies that --reporter json outputs a valid JSON array with the correct schema.
func TestPostEdit_JSON_Schema(t *testing.T) {
	tmpDir := t.TempDir()

	writeDefaultConfigYML(t, tmpDir, "greet", "echo hello")

	targetFile := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(targetFile, []byte("package main\n"), 0o644)
	require.NoError(t, err)

	cfgPath := filepath.Join(tmpDir, ".seil.yml")
	result := runSeil(t, "", "-c", cfgPath, "--reporter", "json", "post-edit", targetFile)

	assert.Equal(t, 0, result.ExitCode, "stderr: %s", result.Stderr)

	var results groupedResultsJSON
	err = json.Unmarshal([]byte(result.Stdout), &results)
	require.NoError(t, err, "stdout should be valid JSON: %s", result.Stdout)
	require.Len(t, results.Success, 1)

	assert.Equal(t, "greet", results.Success[0].Name)
	assert.Equal(t, "success", results.Success[0].Status)
	assert.NotEmpty(t, results.Success[0].LogFile)
}

// TestPostEdit_TextFormat_IsDefault verifies that auto reporter uses human-readable text by default.
func TestPostEdit_TextFormat_IsDefault(t *testing.T) {
	tmpDir := t.TempDir()

	writeDefaultConfigYML(t, tmpDir, "greet", "echo hello")

	targetFile := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(targetFile, []byte("package main\n"), 0o644)
	require.NoError(t, err)

	cfgPath := filepath.Join(tmpDir, ".seil.yml")
	result := runSeil(t, "", "-c", cfgPath, "post-edit", targetFile)

	assert.Equal(t, 0, result.ExitCode, "stderr: %s", result.Stderr)
	assert.False(t, strings.HasPrefix(result.Stdout, "["), "stdout should not be a JSON array: %s", result.Stdout)
	assert.Contains(t, result.Stdout, "--- Failures (0) ---")
}

// TestConfig_ExplicitPath verifies that -c <path> loads config from outside the working directory.
func TestConfig_ExplicitPath(t *testing.T) {
	configDir := t.TempDir()
	workDir := filepath.Join(configDir, "workspace")
	err := os.Mkdir(workDir, 0o755)
	require.NoError(t, err)
	writeLegacyConfigYML(t, configDir, "lint", "echo linted")

	targetFile := filepath.Join(workDir, "app.go")
	err = os.WriteFile(targetFile, []byte("package main\n"), 0o644)
	require.NoError(t, err)

	cfgPath := filepath.Join(configDir, "seil.yml")
	result := runSeil(t, workDir, "-c", cfgPath, "--reporter", "json", "post-edit", targetFile)

	assert.Equal(t, 0, result.ExitCode, "stderr: %s", result.Stderr)

	var results groupedResultsJSON
	err = json.Unmarshal([]byte(result.Stdout), &results)
	require.NoError(t, err, "stdout should be valid JSON: %s", result.Stdout)
	require.NotEmpty(t, results.Success)

	assert.Equal(t, "lint", results.Success[0].Name)
}

// TestSetup_JSON_Schema verifies that setup with --reporter json outputs a valid JSON array.
func TestSetup_JSON_Schema(t *testing.T) {
	tmpDir := t.TempDir()

	writeDefaultSetupConfigYML(t, tmpDir, "greet", "echo hello")

	cfgPath := filepath.Join(tmpDir, ".seil.yml")
	result := runSeil(t, "", "-c", cfgPath, "--reporter", "json", "setup")

	assert.Equal(t, 0, result.ExitCode, "stderr: %s", result.Stderr)

	var results groupedResultsJSON
	err := json.Unmarshal([]byte(result.Stdout), &results)
	require.NoError(t, err, "stdout should be valid JSON: %s", result.Stdout)
	require.Len(t, results.Success, 1)

	assert.Equal(t, "greet", results.Success[0].Name)
	assert.Equal(t, "success", results.Success[0].Status)
}

// TestTeardown_JSON_Schema verifies that teardown with --reporter json outputs a valid JSON array.
func TestTeardown_JSON_Schema(t *testing.T) {
	tmpDir := t.TempDir()

	writeDefaultTeardownConfigYML(t, tmpDir, "cleanup", "echo cleaned")

	cfgPath := filepath.Join(tmpDir, ".seil.yml")
	result := runSeil(t, "", "-c", cfgPath, "--reporter", "json", "teardown")

	assert.Equal(t, 0, result.ExitCode, "stderr: %s", result.Stderr)

	var results groupedResultsJSON
	err := json.Unmarshal([]byte(result.Stdout), &results)
	require.NoError(t, err, "stdout should be valid JSON: %s", result.Stdout)
	require.Len(t, results.Success, 1)

	assert.Equal(t, "cleanup", results.Success[0].Name)
	assert.Equal(t, "success", results.Success[0].Status)
}

// TestSetup_TextFormat_IsDefault verifies that setup without --reporter uses human-readable text.
func TestSetup_TextFormat_IsDefault(t *testing.T) {
	tmpDir := t.TempDir()

	writeDefaultSetupConfigYML(t, tmpDir, "greet", "echo hello")

	cfgPath := filepath.Join(tmpDir, ".seil.yml")
	result := runSeil(t, "", "-c", cfgPath, "setup")

	assert.Equal(t, 0, result.ExitCode, "stderr: %s", result.Stderr)
	assert.False(t, strings.HasPrefix(result.Stdout, "["), "stdout should not be a JSON array: %s", result.Stdout)
	assert.Contains(t, result.Stdout, "--- Failures (0) ---")
}

// TestConfig_AutoDiscovery verifies that without -c the config is discovered from the working directory.
func TestConfig_AutoDiscovery(t *testing.T) {
	tmpRepoRoot := t.TempDir()

	// Create .git dir so git.FindRepoRootFrom recognizes tmpDir as a repo root.
	err := os.Mkdir(filepath.Join(tmpRepoRoot, ".git"), 0o755)
	require.NoError(t, err)

	writeDefaultConfigYML(t, tmpRepoRoot, "fmt", "echo formatted")

	targetFile := filepath.Join(tmpRepoRoot, "main.go")
	err = os.WriteFile(targetFile, []byte("package main\n"), 0o644)
	require.NoError(t, err)

	result := runSeil(t, tmpRepoRoot, "--reporter", "json", "post-edit", targetFile)

	assert.Equal(t, 0, result.ExitCode, "stderr: %s", result.Stderr)

	var results groupedResultsJSON
	err = json.Unmarshal([]byte(result.Stdout), &results)
	require.NoError(t, err, "stdout should be valid JSON: %s", result.Stdout)
	require.NotEmpty(t, results.Success)

	assert.Equal(t, "fmt", results.Success[0].Name)
}

func TestPostEdit_AutoDiscovery_MissingConfig_IsNoOpForJSONReporter(t *testing.T) {
	tmpDir := t.TempDir()

	targetFile := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(targetFile, []byte("package main\n"), 0o644)
	require.NoError(t, err)

	result := runSeil(t, tmpDir, "--reporter", "json", "post-edit", targetFile)

	assert.Equal(t, 0, result.ExitCode, "stderr: %s", result.Stderr)
	assert.Empty(t, result.Stderr)

	var results groupedResultsJSON
	err = json.Unmarshal([]byte(result.Stdout), &results)
	require.NoError(t, err, "stdout should be valid JSON: %s", result.Stdout)
	assert.Empty(t, results.Failure)
	assert.Empty(t, results.Success)
	assert.Empty(t, results.Skipped)
}

func TestSetup_AutoDiscovery_MissingConfig_IsNoOp(t *testing.T) {
	tmpDir := t.TempDir()

	result := runSeil(t, tmpDir, "setup")

	assert.Equal(t, 0, result.ExitCode, "stderr: %s", result.Stderr)
	assert.Empty(t, result.Stderr)
	assert.Contains(t, result.Stdout, "0 succeeded, 0 failed, 0 skipped")
}

func TestTeardown_AutoDiscovery_MissingConfig_IsNoOp(t *testing.T) {
	tmpDir := t.TempDir()

	result := runSeil(t, tmpDir, "teardown")

	assert.Equal(t, 0, result.ExitCode, "stderr: %s", result.Stderr)
	assert.Empty(t, result.Stderr)
	assert.Contains(t, result.Stdout, "0 succeeded, 0 failed, 0 skipped")
}

func TestPostEdit_AutoDiscovery_MissingConfig_IsNoOpForClaudeReporter(t *testing.T) {
	tmpDir := t.TempDir()

	targetFile := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(targetFile, []byte("package main\n"), 0o644)
	require.NoError(t, err)

	result := runSeilWithEnv(t, tmpDir, map[string]string{
		"AI_AGENT": "claude",
	}, "post-edit", targetFile)

	assert.Equal(t, 0, result.ExitCode, "stderr: %s", result.Stderr)
	assert.Empty(t, result.Stderr)
	assert.Equal(t, "0 succeeded, 0 failed, 0 skipped\n", result.Stdout)
}

func TestConfig_ExplicitMissingPath_Fails(t *testing.T) {
	tmpDir := t.TempDir()
	missingCfgPath := filepath.Join(tmpDir, "missing.yml")

	result := runSeil(t, tmpDir, "-c", missingCfgPath, "setup")

	assert.NotEqual(t, 0, result.ExitCode)
	assert.Contains(t, result.Stderr, "missing.yml")
}

// TestPostEdit_JSON_Failure_ExitCode verifies that reporter json exits with code 1 when a hook fails.
func TestPostEdit_JSON_Failure_ExitCode(t *testing.T) {
	tmpDir := t.TempDir()

	writeDefaultConfigYML(t, tmpDir, "error", "exit 1")

	targetFile := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(targetFile, []byte("package main\n"), 0o644)
	require.NoError(t, err)

	cfgPath := filepath.Join(tmpDir, ".seil.yml")
	result := runSeil(t, "", "-c", cfgPath, "--reporter", "json", "post-edit", targetFile)

	assert.Equal(t, 1, result.ExitCode)

	var results groupedResultsJSON
	err = json.Unmarshal([]byte(result.Stdout), &results)
	require.NoError(t, err, "stdout should be valid JSON: %s", result.Stdout)
	require.Len(t, results.Failure, 1)

	assert.Equal(t, "error", results.Failure[0].Name)
	assert.Equal(t, "failure", results.Failure[0].Status)
}

func TestPostEdit_AIClaude_UsesClaudeReporter(t *testing.T) {
	tmpDir := t.TempDir()

	writeDefaultConfigYML(t, tmpDir, "error", "echo boom && exit 1")

	targetFile := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(targetFile, []byte("package main\n"), 0o644)
	require.NoError(t, err)

	cfgPath := filepath.Join(tmpDir, ".seil.yml")
	result := runSeilWithEnv(t, "", map[string]string{
		"AI_AGENT": "claude",
	}, "-c", cfgPath, "post-edit", targetFile)

	assert.Equal(t, 2, result.ExitCode)
	assert.Equal(t, "0 succeeded, 1 failed, 0 skipped\n", result.Stdout)
	assert.Contains(t, result.Stderr, "hook: error")
	assert.NotContains(t, result.Stdout, "--- Failures")
}

func TestSetup_AIClaude_UsesSameReporterSelection(t *testing.T) {
	tmpDir := t.TempDir()

	writeDefaultSetupConfigYML(t, tmpDir, "error", "echo boom && exit 1")

	cfgPath := filepath.Join(tmpDir, ".seil.yml")
	result := runSeilWithEnv(t, "", map[string]string{
		"AI_AGENT": "claude",
	}, "-c", cfgPath, "setup")

	assert.Equal(t, 2, result.ExitCode)
	assert.Equal(t, "0 succeeded, 1 failed, 0 skipped\n", result.Stdout)
	assert.Contains(t, result.Stderr, "hook: error")
}

func TestPostEdit_ExplicitJSONReporterOverridesClaudeAutoSelection(t *testing.T) {
	tmpDir := t.TempDir()

	writeDefaultConfigYML(t, tmpDir, "error", "exit 1")

	targetFile := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(targetFile, []byte("package main\n"), 0o644)
	require.NoError(t, err)

	cfgPath := filepath.Join(tmpDir, ".seil.yml")
	result := runSeilWithEnv(t, "", map[string]string{
		"AI_AGENT": "claude",
	}, "-c", cfgPath, "--reporter", "json", "post-edit", targetFile)

	assert.Equal(t, 1, result.ExitCode)
	assert.Empty(t, result.Stderr)

	var results groupedResultsJSON
	err = json.Unmarshal([]byte(result.Stdout), &results)
	require.NoError(t, err)
	require.Len(t, results.Failure, 1)
}

func TestExplicitClaudeReporterForcesClaudeBehavior(t *testing.T) {
	tmpDir := t.TempDir()

	writeDefaultConfigYML(t, tmpDir, "error", "echo boom && exit 1")

	targetFile := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(targetFile, []byte("package main\n"), 0o644)
	require.NoError(t, err)

	cfgPath := filepath.Join(tmpDir, ".seil.yml")
	result := runSeil(t, "", "-c", cfgPath, "--reporter", "claude", "post-edit", targetFile)

	assert.Equal(t, 2, result.ExitCode)
	assert.Equal(t, "0 succeeded, 1 failed, 0 skipped\n", result.Stdout)
	assert.Contains(t, result.Stderr, "hook: error")
}

func TestInvalidReporterFailsParsing(t *testing.T) {
	tmpDir := t.TempDir()

	writeDefaultSetupConfigYML(t, tmpDir, "greet", "echo hello")

	cfgPath := filepath.Join(tmpDir, ".seil.yml")
	result := runSeil(t, "", "-c", cfgPath, "--reporter", "wat", "setup")

	assert.NotEqual(t, 0, result.ExitCode)
	assert.Contains(t, result.Stderr, "invalid reporter")
}

func TestAIAgent_OverridesAutoDetectedAgent(t *testing.T) {
	tmpDir := t.TempDir()

	writeDefaultConfigYML(t, tmpDir, "error", "echo boom && exit 1")

	targetFile := filepath.Join(tmpDir, "main.go")
	err := os.WriteFile(targetFile, []byte("package main\n"), 0o644)
	require.NoError(t, err)

	cfgPath := filepath.Join(tmpDir, ".seil.yml")
	result := runSeilWithEnv(t, "", map[string]string{
		"AI_AGENT":        "claude",
		"CODEX_THREAD_ID": "thread-1",
	}, "-c", cfgPath, "post-edit", targetFile)

	assert.Equal(t, 2, result.ExitCode)
	assert.Contains(t, result.Stderr, "hook: error")
}
