package hook

import (
	"database/sql"
	"fmt"

	"github.com/dansimau/agentstats/internal/gitx"
	"github.com/dansimau/agentstats/internal/project"
	"github.com/google/uuid"
)

// RecordPromptStart persists the start of a prompt.
func RecordPromptStart(db *sql.DB, input *HookInput) error {
	proj, err := project.Upsert(db, input.Cwd)
	if err != nil {
		return fmt.Errorf("upsert project: %w", err)
	}

	// INSERT OR IGNORE: session may already exist (multiple prompts per session).
	if _, err := db.Exec(
		`INSERT OR IGNORE INTO sessions (id, project_id, agent_type) VALUES (?, ?, ?)`,
		input.SessionID, proj.ID, input.AgentType,
	); err != nil {
		return fmt.Errorf("upsert session: %w", err)
	}

	hashStart := gitx.HeadHash(input.Cwd)
	var hashVal interface{}
	if hashStart != "" {
		hashVal = hashStart
	}

	promptID := uuid.New().String()
	var promptText interface{}
	if input.PromptText != "" {
		promptText = input.PromptText
	}

	if _, err := db.Exec(
		`INSERT INTO prompts (id, session_id, project_id, prompt_text, submitted_at, git_hash_start, agent_type)
		 VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP, ?, ?)`,
		promptID, input.SessionID, proj.ID, promptText, hashVal, input.AgentType,
	); err != nil {
		return fmt.Errorf("insert prompt: %w", err)
	}

	return nil
}

// RecordPromptEnd marks the most recent open prompt in this session as complete.
func RecordPromptEnd(db *sql.DB, input *HookInput) error {
	hashEnd := gitx.HeadHash(input.Cwd)
	var hashVal interface{}
	if hashEnd != "" {
		hashVal = hashEnd
	}

	result, err := db.Exec(
		`UPDATE prompts
		 SET completed_at = CURRENT_TIMESTAMP,
		     git_hash_end = ?
		 WHERE id = (
		     SELECT id FROM prompts
		     WHERE session_id = ? AND completed_at IS NULL
		     ORDER BY submitted_at DESC
		     LIMIT 1
		 )`,
		hashVal, input.SessionID,
	)
	if err != nil {
		return fmt.Errorf("update prompt: %w", err)
	}

	// 0 rows affected is a silent no-op (Stop fires even with no preceding prompt).
	_ = result
	return nil
}
