package agent_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sushichan044/seil/internal/agent"
)

func TestParse(t *testing.T) {
	assert.Equal(t, agent.AgentClaude, agent.Parse("Claude"))
	assert.Equal(t, agent.AgentClaude, agent.Parse("  CLAUDE  "))
	assert.Equal(t, agent.AgentCursor, agent.Parse("cursor"))
	assert.Equal(t, agent.AgentDevin, agent.Parse("devin"))
	assert.Equal(t, agent.AgentReplit, agent.Parse("replit"))
	assert.Equal(t, agent.AgentGemini, agent.Parse("gemini"))
	assert.Equal(t, agent.AgentCodex, agent.Parse("codex"))
	assert.Equal(t, agent.AgentAuggie, agent.Parse("auggie"))
	assert.Equal(t, agent.AgentOpenCode, agent.Parse("opencode"))
	assert.Equal(t, agent.AgentKiro, agent.Parse("kiro"))
	assert.Equal(t, agent.AgentGoose, agent.Parse("goose"))
	assert.Equal(t, agent.AgentPi, agent.Parse("pi"))
	assert.Equal(t, agent.AgentUnknown, agent.Parse("unknown"))
}

func TestDetect(_ *testing.T) {
	_ = agent.Detect()
}

func TestDetectFromLookup(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want agent.Agent
	}{
		{
			name: "prefers AI_AGENT over auto detection",
			env:  map[string]string{"AI_AGENT": "claude", "CODEX_THREAD_ID": "thread"},
			want: agent.AgentClaude,
		},
		{
			name: "detects claude from CLAUDECODE",
			env:  map[string]string{"CLAUDECODE": "1"},
			want: agent.AgentClaude,
		},
		{
			name: "detects claude from CLAUDE_CODE",
			env:  map[string]string{"CLAUDE_CODE": "1"},
			want: agent.AgentClaude,
		},
		{
			name: "detects replit from REPL_ID",
			env:  map[string]string{"REPL_ID": "my-repl"},
			want: agent.AgentReplit,
		},
		{
			name: "detects gemini from GEMINI_CLI",
			env:  map[string]string{"GEMINI_CLI": "1"},
			want: agent.AgentGemini,
		},
		{
			name: "detects codex from CODEX_SANDBOX",
			env:  map[string]string{"CODEX_SANDBOX": "danger-full-access"},
			want: agent.AgentCodex,
		},
		{
			name: "detects codex from CODEX_THREAD_ID",
			env:  map[string]string{"CODEX_THREAD_ID": "thread-123"},
			want: agent.AgentCodex,
		},
		{
			name: "detects opencode from OPENCODE",
			env:  map[string]string{"OPENCODE": "1"},
			want: agent.AgentOpenCode,
		},
		{
			name: "detects pi from unix path pattern",
			env:  map[string]string{"PATH": "/tmp/.pi/agent/bin:/usr/bin"},
			want: agent.AgentPi,
		},
		{
			name: "detects pi from windows path pattern",
			env:  map[string]string{"PATH": `C:\Users\user\.pi\agent\bin`},
			want: agent.AgentPi,
		},
		{
			name: "detects auggie from AUGMENT_AGENT",
			env:  map[string]string{"AUGMENT_AGENT": "1"},
			want: agent.AgentAuggie,
		},
		{
			name: "detects goose from GOOSE_PROVIDER",
			env:  map[string]string{"GOOSE_PROVIDER": "openai"},
			want: agent.AgentGoose,
		},
		{
			name: "detects devin from EDITOR containing devin",
			env:  map[string]string{"EDITOR": "/usr/local/bin/devin-editor"},
			want: agent.AgentDevin,
		},
		{
			name: "detects cursor from CURSOR_AGENT",
			env:  map[string]string{"CURSOR_AGENT": "1"},
			want: agent.AgentCursor,
		},
		{
			name: "detects kiro from TERM_PROGRAM containing kiro",
			env:  map[string]string{"TERM_PROGRAM": "kiro"},
			want: agent.AgentKiro,
		},
		{
			name: "returns unknown when no agent env present",
			env:  map[string]string{},
			want: agent.AgentUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detected := agent.DetectFromLookup(func(key string) (string, bool) {
				value, ok := tt.env[key]
				return value, ok
			})
			assert.Equal(t, tt.want, detected)
		})
	}
}
