package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/martins6/opencode-telegram/internal/config"
	"github.com/martins6/opencode-telegram/internal/database"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	mailSender  string
	mailSubject string
	mailContent string
)

var mailCmd = &cobra.Command{
	Use:   "mail",
	Short: "Manage mails",
	Long:  "Send mails.",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		workspacePath := viper.GetString("workspace.path")
		if workspacePath == "" {
			homeDir, _ := os.UserHomeDir()
			workspacePath = filepath.Join(homeDir, ".opencode-telegram")
		}
		if err := database.Init(workspacePath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to initialize database: %v\n", err)
		}
	},
}

var mailSendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send a mail",
	Long:  "Send a mail to the user. The mail will be delivered immediately.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if mailSender == "" {
			return fmt.Errorf("sender is required")
		}
		if mailSubject == "" {
			return fmt.Errorf("subject is required")
		}
		if mailContent == "" {
			return fmt.Errorf("content is required")
		}

		userID := config.GetAllowedUserChatID()
		if userID == 0 {
			return fmt.Errorf("please send a message to the bot first to register your chat ID")
		}

		id := generateUUID()
		err := database.InsertMail(id, userID, mailSender, mailSubject, mailContent)
		if err != nil {
			return fmt.Errorf("failed to insert mail: %w", err)
		}

		fmt.Printf("Mail queued with ID: %s\n", id)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(mailCmd)
	mailCmd.AddCommand(mailSendCmd)

	mailSendCmd.Flags().StringVar(&mailSender, "sender", "", "Sender of the mail (required)")
	mailSendCmd.Flags().StringVar(&mailSubject, "subject", "", "Subject of the mail (required)")
	mailSendCmd.Flags().StringVar(&mailContent, "content", "", "Content of the mail (required)")
}

func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%s-%s-%s-%s-%s", hex.EncodeToString(b[0:4]), hex.EncodeToString(b[4:6]), hex.EncodeToString(b[6:8]), hex.EncodeToString(b[8:10]), hex.EncodeToString(b[10:16]))
}
