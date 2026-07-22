package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/martins6/acolyte/internal/config"
	"github.com/martins6/acolyte/internal/logger"
	"github.com/spf13/cobra"
)

var logsDate string

var logsCmd = &cobra.Command{
	Use:   "logs [N]",
	Short: "View the most recent log entries",
	Long: `View log entries written by Acolyte.

By default, prints the 10 most recent entries across all retained log files,
oldest first within the slice. Pass a different N as the positional argument.

Use --date to restrict output to a single day:
  acolyte logs --date today
  acolyte logs 25 --date 2026-07-20`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return fmt.Errorf("accepts at most one argument: [N]")
		}
		if len(args) == 1 {
			n, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid integer %q: %w", args[0], err)
			}
			if n <= 0 {
				return fmt.Errorf("N must be greater than 0 (got %d)", n)
			}
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		n := 10
		if len(args) == 1 {
			parsed, err := strconv.Atoi(args[0])
			if err != nil {
				return fmt.Errorf("invalid integer %q: %w", args[0], err)
			}
			n = parsed
		}

		workspacePath, err := resolveLogsWorkspace()
		if err != nil {
			return err
		}

		var onlyDate time.Time
		if logsDate != "" {
			switch logsDate {
			case "today":
				onlyDate = time.Now()
			default:
				parsed, parseErr := time.ParseInLocation("2006-01-02", logsDate, time.Local)
				if parseErr != nil {
					return fmt.Errorf("invalid --date %q (expected 'today' or YYYY-MM-DD): %w", logsDate, parseErr)
				}
				onlyDate = parsed
			}
		}

		entries, err := logger.TailLastN(workspacePath, n, onlyDate)
		if err != nil {
			return fmt.Errorf("failed to read logs: %w", err)
		}
		if len(entries) == 0 {
			if logsDate != "" {
				fmt.Printf("no log entries found for %s\n", logsDate)
			} else {
				fmt.Println("no log entries found")
			}
			return nil
		}
		for _, e := range entries {
			fmt.Println(e.Raw)
		}
		return nil
	},
}

func resolveLogsWorkspace() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("locate home directory: %w", err)
	}
	defaultWorkspace := filepath.Join(homeDir, ".acolyte")
	configPath := filepath.Join(defaultWorkspace, "config.toml")

	if cfg, err := config.LoadIfExists(configPath); err == nil && cfg != nil && cfg.Workspace.Path != "" {
		return cfg.Workspace.Path, nil
	}
	return defaultWorkspace, nil
}

func init() {
	logsCmd.Flags().StringVar(&logsDate, "date", "", "Restrict to a single day: 'today' or YYYY-MM-DD")
	rootCmd.AddCommand(logsCmd)
}
