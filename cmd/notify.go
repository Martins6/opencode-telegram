package cmd

import (
	"fmt"

	"github.com/martins6/opencode-telegram/internal/database"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var notifyCmd = &cobra.Command{
	Use:   "notify \"message\"",
	Short: "Send a notification to the Telegram user",
	Long:  "Send a notification that will be delivered to the Telegram user via the bot.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		message := args[0]

		userID := viper.GetInt64("notifications.user_id")
		if userID == 0 {
			userID = 1
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
