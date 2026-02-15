package hook

import (
	"encoding/json"
	"fmt"
	"io"
)

// claudeCodePayload is the JSON structure Claude Code sends to hooks.
type claudeCodePayload struct {
	SessionID   string `json:"session_id"`
	Cwd         string `json:"cwd"`
	HookEvent   string `json:"hook_event_name"`
	Prompt      string `json:"prompt"`         // present on UserPromptSubmit
	Transcript  string `json:"transcript_path"`
	Permission  string `json:"permission_mode"`
}

// ClaudeCodeParser implements Parser for Claude Code hooks.
type ClaudeCodeParser struct{}

func (p *ClaudeCodeParser) AgentType() string { return "claude-code" }

func (p *ClaudeCodeParser) Parse(r io.Reader, eventType EventType) (*HookInput, error) {
	var payload claudeCodePayload
	dec := json.NewDecoder(r)
	if err := dec.Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode claude-code payload: %w", err)
	}

	if payload.SessionID == "" {
		return nil, fmt.Errorf("missing session_id in hook payload")
	}
	if payload.Cwd == "" {
		return nil, fmt.Errorf("missing cwd in hook payload")
	}

	return &HookInput{
		SessionID:  payload.SessionID,
		Cwd:        payload.Cwd,
		PromptText: payload.Prompt,
		AgentType:  "claude-code",
		EventType:  eventType,
	}, nil
}
