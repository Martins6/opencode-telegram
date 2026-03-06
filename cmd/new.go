package cmd

import (
	"fmt"
	"os"

	"github.com/martins6/opencode-telegram/internal/config"
	"github.com/martins6/opencode-telegram/internal/workspace"
	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new",
	Short: "Initialize new workspace from template",
	Long: `Creates a new workspace directory with all required template files.

This will create:
- AGENTS.md (agent definitions)
- SOUL.md (system operator behavior)
- USER.md (user information)
- IDENTITY.md (model identity)
- BOOTSTRAP.md (first-time setup)
- TOOLS.md (tool definitions)
- downloads/ directory structure`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path := ""
		if len(args) > 0 {
			path = args[0]
		} else {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			path = homeDir + "/.opencode-telegram"
		}

		if err := workspace.CreateTemplate(path); err != nil {
			return fmt.Errorf("failed to create workspace: %w", err)
		}

		if _, err := config.Load(""); err != nil {
			return fmt.Errorf("failed to create config: %w", err)
		}

		fmt.Printf("Workspace created at: %s\n", path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
}
