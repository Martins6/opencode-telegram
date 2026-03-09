package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/martins6/opencode-telegram/internal/database"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	mailSender  string
	mailSubject string
	mailContent string
	mailUrgency string
	mailJSON    bool
)

var mailCmd = &cobra.Command{
	Use:   "mail",
	Short: "Manage mails",
	Long:  "Send, view, or list mails.",
}

var mailSendCmd = &cobra.Command{
	Use:   "send",
	Short: "Send a mail",
	Long:  "Send a mail to the user. The mail will be delivered based on its urgency.",
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

		if mailUrgency == "" {
			mailUrgency = "low"
		}
		if mailUrgency != "high" && mailUrgency != "medium" && mailUrgency != "low" {
			return fmt.Errorf("urgency must be high, medium, or low")
		}

		userID := viper.GetInt64("notifications.user_id")
		if userID == 0 {
			userID = 1
		}

		id := generateUUID()
		err := database.InsertMail(id, userID, mailSender, mailSubject, mailContent, mailUrgency)
		if err != nil {
			return fmt.Errorf("failed to insert mail: %w", err)
		}

		fmt.Printf("Mail queued with ID: %s\n", id)
		return nil
	},
}

var mailOpenCmd = &cobra.Command{
	Use:   "open [id]",
	Short: "Open a mail",
	Long:  "View the content of a mail by its ID.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mailID := args[0]

		mail, err := database.GetMail(mailID)
		if err != nil {
			return fmt.Errorf("failed to get mail: %w", err)
		}

		if mailJSON {
			jsonBytes, err := json.MarshalIndent(mail, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal mail: %w", err)
			}
			fmt.Println(string(jsonBytes))
		} else {
			fmt.Printf("Sender: %s\n", mail.Sender)
			fmt.Printf("Subject: %s\n", mail.Subject)
			fmt.Printf("Content: %s\n", mail.Content)
		}

		return nil
	},
}

var mailListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all mails",
	Long:  "List all mails received by the user.",
	RunE: func(cmd *cobra.Command, args []string) error {
		userID := viper.GetInt64("notifications.user_id")
		if userID == 0 {
			userID = 1
		}

		mails, err := database.ListMails(userID)
		if err != nil {
			return fmt.Errorf("failed to list mails: %w", err)
		}

		if len(mails) == 0 {
			fmt.Println("No mails found.")
			return nil
		}

		fmt.Println("ID\t\t\t\t Sender\t\t Subject\t\t Urgency\t Sent")
		fmt.Println("---------------------------------------------------------------------------------------------------")
		for _, m := range mails {
			sentStr := "No"
			if m.MailSent {
				sentStr = "Yes"
			}
			fmt.Printf("%s\t %s\t %s\t %s\t %s\n", m.ID[:8], m.Sender, m.Subject, m.Urgency, sentStr)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(mailCmd)
	mailCmd.AddCommand(mailSendCmd)
	mailCmd.AddCommand(mailOpenCmd)
	mailCmd.AddCommand(mailListCmd)

	mailSendCmd.Flags().StringVar(&mailSender, "sender", "", "Sender of the mail (required)")
	mailSendCmd.Flags().StringVar(&mailSubject, "subject", "", "Subject of the mail (required)")
	mailSendCmd.Flags().StringVar(&mailContent, "content", "", "Content of the mail (required)")
	mailSendCmd.Flags().StringVar(&mailUrgency, "urgency", "low", "Urgency level: high, medium, or low")

	mailOpenCmd.Flags().BoolVar(&mailJSON, "json", false, "Output mail in JSON format")
}

func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%s-%s-%s-%s-%s", hex.EncodeToString(b[0:4]), hex.EncodeToString(b[4:6]), hex.EncodeToString(b[6:8]), hex.EncodeToString(b[8:10]), hex.EncodeToString(b[10:16]))
}
