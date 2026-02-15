package cli

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/dansimau/agentstats/internal/db"
	"github.com/dansimau/agentstats/internal/project"
	"github.com/spf13/cobra"
)

// NewStatsCmd returns the 'stats' subcommand.
func NewStatsCmd() *cobra.Command {
	var projectDir string
	var dbPath string

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show AI working time statistics for a project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStats(dbPath, projectDir)
		},
	}

	cmd.Flags().StringVarP(&projectDir, "project", "p", "", "Project directory (default: current directory)")
	cmd.Flags().StringVar(&dbPath, "db", "", "Path to database (default: XDG data dir)")
	return cmd
}

type statsResult struct {
	totalPrompts    int
	completedPrompts int
	totalSeconds    float64
	firstSubmit     string
	lastSubmit      string
}

func runStats(dbPath, projectDir string) error {
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

	stats, err := queryStats(database, proj.ID)
	if err != nil {
		return fmt.Errorf("query stats: %w", err)
	}

	fmt.Printf("Project:               %s", proj.ShortName())
	if proj.DisplayOrigin() != "" {
		fmt.Printf(" (%s)", proj.DisplayOrigin())
	}
	fmt.Println()

	if proj.GitOrigin != "" {
		fmt.Printf("Git origin:            %s\n", proj.GitOrigin)
	}

	fmt.Printf("Total prompts:         %d\n", stats.totalPrompts)
	fmt.Printf("Total AI working time: %s\n", formatDuration(stats.totalSeconds))

	if stats.completedPrompts > 0 {
		avg := stats.totalSeconds / float64(stats.completedPrompts)
		fmt.Printf("Average per prompt:    %s\n", formatDuration(avg))
	}

	if stats.firstSubmit != "" && stats.lastSubmit != "" {
		period := stats.firstSubmit
		if stats.firstSubmit != stats.lastSubmit {
			period = stats.firstSubmit + " to " + stats.lastSubmit
		}
		fmt.Printf("Time period:           %s\n", period)
	}

	return nil
}

func queryStats(database *sql.DB, projectID string) (*statsResult, error) {
	row := database.QueryRow(`
		SELECT
			COUNT(*),
			COUNT(completed_at),
			COALESCE(SUM(
				CASE WHEN completed_at IS NOT NULL
				THEN (julianday(completed_at) - julianday(submitted_at)) * 86400.0
				ELSE 0 END
			), 0),
			COALESCE(MIN(DATE(submitted_at)), ''),
			COALESCE(MAX(DATE(submitted_at)), '')
		FROM prompts
		WHERE project_id = ?
	`, projectID)

	var r statsResult
	if err := row.Scan(
		&r.totalPrompts,
		&r.completedPrompts,
		&r.totalSeconds,
		&r.firstSubmit,
		&r.lastSubmit,
	); err != nil {
		return nil, err
	}
	return &r, nil
}

func formatDuration(seconds float64) string {
	if seconds <= 0 {
		return "0s"
	}
	total := int(seconds)
	h := total / 3600
	m := (total % 3600) / 60
	s := total % 60

	if h > 0 {
		return fmt.Sprintf("%dh %dm %ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}
