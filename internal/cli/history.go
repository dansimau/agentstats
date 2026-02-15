package cli

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/dansimau/agentstats/internal/db"
	"github.com/dansimau/agentstats/internal/project"
	"github.com/spf13/cobra"
)

// NewHistoryCmd returns the 'history' subcommand.
func NewHistoryCmd() *cobra.Command {
	var projectDir string
	var dbPath string
	var limit int

	cmd := &cobra.Command{
		Use:   "history",
		Short: "Show recent prompt history for a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHistory(dbPath, projectDir, limit)
		},
	}

	cmd.Flags().StringVarP(&projectDir, "project", "p", "", "Project directory (default: current directory)")
	cmd.Flags().StringVar(&dbPath, "db", "", "Path to database (default: XDG data dir)")
	cmd.Flags().IntVarP(&limit, "limit", "n", 50, "Number of prompts to show")
	return cmd
}

type promptRow struct {
	num         int
	submittedAt string
	duration    string // "-" for in-flight
	promptText  string
}

func runHistory(dbPath, projectDir string, limit int) error {
	if dbPath == "" {
		dbPath = db.DefaultPath()
	}
	if projectDir == "" {
		var err error
		projectDir, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("get cwd: %w", err)
		}
	}

	database, err := db.Open(dbPath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer database.Close()

	proj, err := project.Find(database, projectDir)
	if err != nil {
		return fmt.Errorf("find project: %w", err)
	}
	if proj == nil {
		fmt.Println("No project found for", projectDir)
		fmt.Println("Run an AI agent in this directory first to start tracking.")
		return nil
	}

	rows, err := queryHistory(database, proj.ID, limit)
	if err != nil {
		return fmt.Errorf("query history: %w", err)
	}

	if len(rows) == 0 {
		fmt.Println("No prompts recorded yet.")
		return nil
	}

	printHistory(rows)
	return nil
}

func queryHistory(database *sql.DB, projectID string, limit int) ([]promptRow, error) {
	sqlRows, err := database.Query(`
		SELECT
			ROW_NUMBER() OVER (ORDER BY submitted_at DESC) AS num,
			strftime('%Y-%m-%d %H:%M:%S', submitted_at) AS submitted_at,
			CASE
				WHEN completed_at IS NULL THEN NULL
				ELSE CAST(ROUND((julianday(completed_at) - julianday(submitted_at)) * 86400) AS INTEGER)
			END AS duration_secs,
			COALESCE(prompt_text, '')
		FROM prompts
		WHERE project_id = ?
		ORDER BY submitted_at DESC
		LIMIT ?
	`, projectID, limit)
	if err != nil {
		return nil, err
	}
	defer sqlRows.Close()

	var results []promptRow
	for sqlRows.Next() {
		var r promptRow
		var durationSecs sql.NullInt64
		var promptText string
		if err := sqlRows.Scan(&r.num, &r.submittedAt, &durationSecs, &promptText); err != nil {
			return nil, err
		}
		if durationSecs.Valid {
			r.duration = formatDuration(float64(durationSecs.Int64))
		} else {
			r.duration = "-"
		}
		r.promptText = truncate(promptText, 60)
		results = append(results, r)
	}
	return results, sqlRows.Err()
}

func printHistory(rows []promptRow) {
	// Column widths.
	const (
		numW      = 5
		timeW     = 19
		durationW = 10
	)

	header := fmt.Sprintf("%-*s  %-*s  %-*s  %s",
		numW, "#",
		timeW, "Time",
		durationW, "Duration",
		"Prompt",
	)
	sep := strings.Repeat("-", numW) + "  " +
		strings.Repeat("-", timeW) + "  " +
		strings.Repeat("-", durationW) + "  " +
		strings.Repeat("-", 47)

	fmt.Println(header)
	fmt.Println(sep)

	for _, r := range rows {
		fmt.Printf("%-*d  %-*s  %-*s  %s\n",
			numW, r.num,
			timeW, r.submittedAt,
			durationW, r.duration,
			r.promptText,
		)
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
