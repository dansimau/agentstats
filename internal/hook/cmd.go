package hook

import (
	"fmt"
	"os"

	"github.com/dansimau/agentstats/internal/db"
	"github.com/spf13/cobra"
)

// NewHookCmd returns the 'hook' subcommand (and its children).
func NewHookCmd() *cobra.Command {
	var agentType string
	var dbPath string

	hookCmd := &cobra.Command{
		Use:   "hook",
		Short: "Receive hook events from AI coding agents",
		// Don't show usage on error — hook errors go to stderr only.
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	run := func(eventType EventType) func(cmd *cobra.Command, args []string) {
		return func(cmd *cobra.Command, args []string) {
			// Hooks must always exit 0.
			if err := handleHook(dbPath, agentType, eventType); err != nil {
				fmt.Fprintln(os.Stderr, "agentstats hook error:", err)
			}
		}
	}

	startCmd := &cobra.Command{
		Use:   "prompt-start",
		Short: "Record the start of a prompt (UserPromptSubmit event)",
		Run:   run(EventPromptStart),
	}

	endCmd := &cobra.Command{
		Use:   "prompt-end",
		Short: "Record the end of a prompt (Stop event)",
		Run:   run(EventPromptEnd),
	}

	hookCmd.PersistentFlags().StringVar(&agentType, "agent", "claude-code", "Agent type (e.g. claude-code)")
	hookCmd.PersistentFlags().StringVar(&dbPath, "db", "", "Path to database (default: XDG data dir)")

	hookCmd.AddCommand(startCmd, endCmd)
	return hookCmd
}

func handleHook(dbPath, agentType string, eventType EventType) error {
	if dbPath == "" {
		dbPath = db.DefaultPath()
	}

	database, err := db.Open(dbPath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer database.Close()

	parser, err := ParserForAgent(agentType)
	if err != nil {
		return err
	}

	input, err := parser.Parse(os.Stdin, eventType)
	if err != nil {
		return fmt.Errorf("parse hook input: %w", err)
	}

	switch eventType {
	case EventPromptStart:
		return RecordPromptStart(database, input)
	case EventPromptEnd:
		return RecordPromptEnd(database, input)
	default:
		return fmt.Errorf("unknown event type %d", eventType)
	}
}
