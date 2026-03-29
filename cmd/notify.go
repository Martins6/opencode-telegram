package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/martins6/opencode-telegram/internal/config"
	"github.com/martins6/opencode-telegram/internal/database"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var notifyCmd = &cobra.Command{
	Use:   "notify \"message\"",
	Short: "Send a notification to the Telegram user",
	Long:  "Send a notification that will be delivered to the Telegram user via the bot.",
	Args:  cobra.ExactArgs(1),
	PreRun: func(cmd *cobra.Command, args []string) {
		if _, err := config.Load(""); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to load config: %v\n", err)
		}

		workspacePath := viper.GetString("workspace.path")
		if workspacePath == "" {
			homeDir, _ := os.UserHomeDir()
			workspacePath = filepath.Join(homeDir, ".opencode-telegram")
		}
		if err := database.Init(workspacePath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to initialize database: %v\n", err)
		}
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		message := args[0]

		userID := config.GetAllowedUserChatID()
		if userID == 0 {
			return fmt.Errorf("please send a message to the bot first to register your chat ID")
		}

		id, err := database.InsertNotification(userID, message)
		if err != nil {
			return fmt.Errorf("failed to insert notification: %w", err)
		}

		fmt.Printf("Notification queued with ID: %d\n", id)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(notifyCmd)
}
