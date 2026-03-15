package agent

import (
	"os"
	"strings"
)

type Agent string

const (
	AgentUnknown  Agent = ""
	AgentClaude   Agent = "claude"
	AgentCursor   Agent = "cursor"
	AgentDevin    Agent = "devin"
	AgentReplit   Agent = "replit"
	AgentGemini   Agent = "gemini"
	AgentCodex    Agent = "codex"
	AgentAuggie   Agent = "auggie"
	AgentOpenCode Agent = "opencode"
	AgentKiro     Agent = "kiro"
	AgentGoose    Agent = "goose"
	AgentPi       Agent = "pi"
)

func Parse(raw string) Agent {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case string(AgentClaude):
		return AgentClaude
	case string(AgentCursor):
		return AgentCursor
	case string(AgentDevin):
		return AgentDevin
	case string(AgentReplit):
		return AgentReplit
	case string(AgentGemini):
		return AgentGemini
	case string(AgentCodex):
		return AgentCodex
	case string(AgentAuggie):
		return AgentAuggie
	case string(AgentOpenCode):
		return AgentOpenCode
	case string(AgentKiro):
		return AgentKiro
	case string(AgentGoose):
		return AgentGoose
	case string(AgentPi):
		return AgentPi
	default:
		return AgentUnknown
	}
}

func Detect() Agent {
	return DetectFromLookup(os.LookupEnv)
}

func DetectFromLookup(lookup func(string) (string, bool)) Agent {
	if value := envValue(lookup, "AI_AGENT"); value != "" {
		return Parse(value)
	}

	switch {
	case hasAnyEnv(lookup, "CLAUDECODE", "CLAUDE_CODE"):
		return AgentClaude
	case hasAnyEnv(lookup, "REPL_ID"):
		return AgentReplit
	case hasAnyEnv(lookup, "GEMINI_CLI"):
		return AgentGemini
	case hasAnyEnv(lookup, "CODEX_SANDBOX", "CODEX_THREAD_ID"):
		return AgentCodex
	case hasAnyEnv(lookup, "OPENCODE"):
		return AgentOpenCode
	case pathLooksLikePi(envValue(lookup, "PATH")):
		return AgentPi
	case hasAnyEnv(lookup, "AUGMENT_AGENT"):
		return AgentAuggie
	case hasAnyEnv(lookup, "GOOSE_PROVIDER"):
		return AgentGoose
	case containsFold(envValue(lookup, "EDITOR"), "devin"):
		return AgentDevin
	case hasAnyEnv(lookup, "CURSOR_AGENT"):
		return AgentCursor
	case containsFold(envValue(lookup, "TERM_PROGRAM"), "kiro"):
		return AgentKiro
	default:
		return AgentUnknown
	}
}

func envValue(lookup func(string) (string, bool), key string) string {
	value, ok := lookup(key)
	if !ok {
		return ""
	}
	return strings.TrimSpace(value)
}

func hasAnyEnv(lookup func(string) (string, bool), keys ...string) bool {
	for _, key := range keys {
		if envValue(lookup, key) != "" {
			return true
		}
	}
	return false
}

func containsFold(value, needle string) bool {
	return strings.Contains(strings.ToLower(value), strings.ToLower(needle))
}

func pathLooksLikePi(path string) bool {
	return strings.Contains(path, ".pi/agent") || strings.Contains(path, `.pi\agent`)
}
