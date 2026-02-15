package main

import (
	"fmt"
	"os"

	"github.com/dansimau/agentstats/internal/cli"
	"github.com/dansimau/agentstats/internal/hook"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "agentstats",
		Short: "Track AI coding agent working time and prompt history",
		Long: `agentstats records prompt timing, session data, and git state for AI coding
agents. Use 'stats' and 'history' to inspect recorded data.`,
		SilenceUsage: true,
	}

	rootCmd.AddCommand(
		hook.NewHookCmd(),
		cli.NewStatsCmd(),
		cli.NewHistoryCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
