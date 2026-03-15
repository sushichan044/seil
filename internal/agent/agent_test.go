package agent_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sushichan044/seil/internal/agent"
)

func TestParse(t *testing.T) {
	assert.Equal(t, agent.AgentClaude, agent.Parse("Claude"))
	assert.Equal(t, agent.AgentCodex, agent.Parse("codex"))
	assert.Equal(t, agent.AgentUnknown, agent.Parse("unknown"))
}

func TestDetectFromLookup(t *testing.T) {
	t.Run("prefers AI_AGENT over auto detection", func(t *testing.T) {
		detected := agent.DetectFromLookup(func(key string) (string, bool) {
			values := map[string]string{
				"AI_AGENT":        "claude",
				"CODEX_THREAD_ID": "thread",
			}
			value, ok := values[key]
			return value, ok
		})

		assert.Equal(t, agent.AgentClaude, detected)
	})

	t.Run("detects codex from env var", func(t *testing.T) {
		detected := agent.DetectFromLookup(func(key string) (string, bool) {
			values := map[string]string{
				"CODEX_SANDBOX": "danger-full-access",
			}
			value, ok := values[key]
			return value, ok
		})

		assert.Equal(t, agent.AgentCodex, detected)
	})

	t.Run("detects cursor from env var", func(t *testing.T) {
		detected := agent.DetectFromLookup(func(key string) (string, bool) {
			values := map[string]string{
				"CURSOR_AGENT": "1",
			}
			value, ok := values[key]
			return value, ok
		})

		assert.Equal(t, agent.AgentCursor, detected)
	})

	t.Run("detects pi from path pattern", func(t *testing.T) {
		detected := agent.DetectFromLookup(func(key string) (string, bool) {
			values := map[string]string{
				"PATH": "/tmp/.pi/agent/bin:/usr/bin",
			}
			value, ok := values[key]
			return value, ok
		})

		assert.Equal(t, agent.AgentPi, detected)
	})

	t.Run("returns unknown when no agent is detected", func(t *testing.T) {
		detected := agent.DetectFromLookup(func(_ string) (string, bool) {
			return "", false
		})

		assert.Equal(t, agent.AgentUnknown, detected)
	})
}
