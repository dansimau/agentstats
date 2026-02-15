package hook

import (
	"fmt"
	"io"
)

// EventType indicates whether the hook is for the start or end of a prompt.
type EventType int

const (
	EventPromptStart EventType = iota
	EventPromptEnd
)

// HookInput is the normalized data extracted from a hook event.
type HookInput struct {
	SessionID  string
	Cwd        string
	PromptText string // empty for prompt-end events
	AgentType  string
	EventType  EventType
}

// Parser knows how to read a hook payload for a specific agent type.
type Parser interface {
	Parse(r io.Reader, eventType EventType) (*HookInput, error)
	AgentType() string
}

// ParserForAgent returns the Parser for the named agent type.
// Add new agents by implementing Parser and registering here.
func ParserForAgent(agentType string) (Parser, error) {
	switch agentType {
	case "claude-code", "":
		return &ClaudeCodeParser{}, nil
	default:
		return nil, fmt.Errorf("unknown agent type %q", agentType)
	}
}
