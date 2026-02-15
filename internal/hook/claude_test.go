package hook_test

import (
	"strings"
	"testing"

	"github.com/dansimau/agentstats/internal/hook"
)

func TestClaudeCodeParser_PromptStart(t *testing.T) {
	json := `{
		"session_id": "abc-123",
		"cwd": "/home/user/myapp",
		"hook_event_name": "UserPromptSubmit",
		"prompt": "Add authentication",
		"transcript_path": "",
		"permission_mode": "default"
	}`

	p := &hook.ClaudeCodeParser{}
	input, err := p.Parse(strings.NewReader(json), hook.EventPromptStart)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if input.SessionID != "abc-123" {
		t.Errorf("SessionID: got %q", input.SessionID)
	}
	if input.Cwd != "/home/user/myapp" {
		t.Errorf("Cwd: got %q", input.Cwd)
	}
	if input.PromptText != "Add authentication" {
		t.Errorf("PromptText: got %q", input.PromptText)
	}
	if input.AgentType != "claude-code" {
		t.Errorf("AgentType: got %q", input.AgentType)
	}
	if input.EventType != hook.EventPromptStart {
		t.Errorf("EventType: got %v", input.EventType)
	}
}

func TestClaudeCodeParser_Stop(t *testing.T) {
	json := `{
		"session_id": "abc-123",
		"cwd": "/home/user/myapp",
		"hook_event_name": "Stop",
		"transcript_path": "",
		"permission_mode": "default"
	}`

	p := &hook.ClaudeCodeParser{}
	input, err := p.Parse(strings.NewReader(json), hook.EventPromptEnd)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if input.EventType != hook.EventPromptEnd {
		t.Errorf("EventType: got %v", input.EventType)
	}
	if input.PromptText != "" {
		t.Errorf("PromptText should be empty for Stop event, got %q", input.PromptText)
	}
}

func TestClaudeCodeParser_MissingSessionID(t *testing.T) {
	json := `{"cwd": "/tmp", "hook_event_name": "Stop"}`
	p := &hook.ClaudeCodeParser{}
	_, err := p.Parse(strings.NewReader(json), hook.EventPromptEnd)
	if err == nil {
		t.Error("expected error for missing session_id")
	}
}

func TestClaudeCodeParser_MissingCwd(t *testing.T) {
	json := `{"session_id": "abc", "hook_event_name": "Stop"}`
	p := &hook.ClaudeCodeParser{}
	_, err := p.Parse(strings.NewReader(json), hook.EventPromptEnd)
	if err == nil {
		t.Error("expected error for missing cwd")
	}
}

func TestParserForAgent(t *testing.T) {
	p, err := hook.ParserForAgent("claude-code")
	if err != nil {
		t.Fatalf("ParserForAgent(claude-code): %v", err)
	}
	if p.AgentType() != "claude-code" {
		t.Errorf("AgentType: got %q", p.AgentType())
	}

	_, err = hook.ParserForAgent("unknown-agent")
	if err == nil {
		t.Error("expected error for unknown agent")
	}
}
