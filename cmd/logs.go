package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var logsCmd = &cobra.Command{
	Use:   "logs [today|YYYY-MM-DD]",
	Short: "View logs from specific date",
	Long:  "View logs from a specific date. Use 'today' for today's logs or specify a date in YYYY-MM-DD format.",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var dateStr string
		if len(args) == 0 {
			dateStr = "today"
		} else {
			dateStr = args[0]
		}

		if dateStr == "today" {
			dateStr = time.Now().Format("2006-01-02")
		}

		workspacePath := viper.GetString("workspace.path")
		if workspacePath == "" {
			homeDir, _ := os.UserHomeDir()
			workspacePath = homeDir + "/.opencode-telegram"
		}

		logFile := filepath.Join(workspacePath, ".logs", dateStr+".log")

		content, err := os.ReadFile(logFile)
		if err != nil {
			return fmt.Errorf("failed to read log file: %w", err)
		}

		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			if line != "" {
				fmt.Println(line)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(logsCmd)
}
