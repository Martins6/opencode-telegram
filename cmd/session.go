package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/martins6/acolyte/internal/config"
	"github.com/martins6/acolyte/internal/opencode"
	"github.com/martins6/acolyte/internal/workspace"
	"github.com/spf13/cobra"
)

func Global() *config.Config {
	return config.Get()
}

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Inspect OpenCode sessions",
	Long:  "List and export OpenCode sessions that live inside the configured workspace.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		_, err := config.LoadIfExists(cfgFile)
		if err != nil {
			if !errors.Is(err, config.ErrConfigNotFound) {
				return fmt.Errorf("load config: %w", err)
			}
			if _, err := config.LoadSingleton(); err != nil {
				if !errors.Is(err, config.ErrConfigNotFound) {
					return fmt.Errorf("load singleton config: %w", err)
				}
			}
		}
		return nil
	},
}

var sessionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List OpenCode sessions in the configured workspace",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		workspacePath, err := resolveSessionWorkspace()
		if err != nil {
			return err
		}
		if err := workspace.StrictValidate(workspacePath); err != nil {
			return err
		}

		stdout, err := opencode.SessionList(cmd.Context(), workspacePath)
		if err != nil {
			if errors.Is(err, opencode.ErrOpenCodeMissing) {
				return fmt.Errorf("opencode CLI not installed; install it from https://opencode.ai and ensure it is on your PATH")
			}
			return err
		}

		_, err = os.Stdout.Write(stdout)
		return err
	},
}

var sessionExportCmd = &cobra.Command{
	Use:   "export <sessionID>",
	Short: "Export a single OpenCode session",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		workspacePath, err := resolveSessionWorkspace()
		if err != nil {
			return err
		}
		if err := workspace.StrictValidate(workspacePath); err != nil {
			return err
		}

		stdout, err := opencode.SessionExport(cmd.Context(), workspacePath, args[0])
		if err != nil {
			if errors.Is(err, opencode.ErrOpenCodeMissing) {
				return fmt.Errorf("opencode CLI not installed; install it from https://opencode.ai and ensure it is on your PATH")
			}
			if errors.Is(err, opencode.ErrInvalidSessionID) {
				return fmt.Errorf("session ID is required")
			}
			return err
		}

		_, err = os.Stdout.Write(stdout)
		return err
	},
}

func resolveSessionWorkspace() (string, error) {
	cfg := Global()
	if cfg != nil && cfg.Workspace.Path != "" {
		return filepath.Abs(cfg.Workspace.Path)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("locate home directory: %w", err)
	}
	return filepath.Join(homeDir, ".acolyte"), nil
}

func init() {
	sessionCmd.AddCommand(sessionListCmd)
	sessionCmd.AddCommand(sessionExportCmd)
	rootCmd.AddCommand(sessionCmd)
}
