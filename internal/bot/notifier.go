package bot

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-telegram/bot"
	"github.com/martins6/opencode-telegram/internal/config"
	"github.com/martins6/opencode-telegram/internal/database"
	"github.com/spf13/viper"
)

var userChatID int64 = 1

type NotifierService struct {
	bot    *bot.Bot
	ctx    context.Context
	cancel context.CancelFunc
}

func StartNotifier(ctx context.Context, b *bot.Bot) error {
	if b == nil {
		log.Println("Notifier: bot not initialized, skipping")
		return nil
	}

	if err := database.Init(); err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	userChatID = viper.GetInt64("notifications.user_id")
	if userChatID == 0 {
		userChatID = 1
	}

	n := &NotifierService{
		bot: b,
		ctx: ctx,
	}
	n.ctx, n.cancel = context.WithCancel(ctx)

	go n.run()

	log.Println("Notifier service started")
	return nil
}

func (n *NotifierService) run() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-n.ctx.Done():
			log.Println("Notifier service stopped")
			return
		case <-ticker.C:
			n.processNotifications()
			n.processMails()
		}
	}
}

func (n *NotifierService) processNotifications() {
	notifications, err := database.GetUnsentNotifications(userChatID)
	if err != nil {
		log.Printf("Notifier: failed to get notifications: %v", err)
		return
	}

	for _, notification := range notifications {
		message := fmt.Sprintf("--- Notification (Agent Unaware) ---\nReceived at: %s\n\n%s",
			notification.CreatedAt.Format("2006-01-02 15:04:05"),
			notification.Message,
		)

		n.bot.SendMessage(n.ctx, &bot.SendMessageParams{
			ChatID: userChatID,
			Text:   message,
		})

		if err := database.MarkNotificationSent(notification.ID); err != nil {
			log.Printf("Notifier: failed to mark notification as sent: %v", err)
		}
	}
}

func (n *NotifierService) processMails() {
	cfg := config.Get()
	mediumHours := 1

	if cfg != nil && cfg.Mail.UrgencyTiming.MediumHours > 0 {
		mediumHours = cfg.Mail.UrgencyTiming.MediumHours
	}

	now := time.Now()

	mails, err := database.GetUnsentMails(userChatID)
	if err != nil {
		log.Printf("Notifier: failed to get mails: %v", err)
		return
	}

	for _, mail := range mails {
		shouldSend := false

		switch mail.Urgency {
		case "high":
			shouldSend = true
		case "medium":
			if now.Sub(mail.CreatedAt) >= time.Duration(mediumHours)*time.Hour {
				shouldSend = true
			}
		case "low":
			shouldSend = false
		}

		if shouldSend {
			message := fmt.Sprintf("It seems an mail has arrived. Please check it with `opencode-telegram mail open %s`. Please simulate as if you had received the mail not me.", mail.ID)

			n.bot.SendMessage(n.ctx, &bot.SendMessageParams{
				ChatID: userChatID,
				Text:   message,
			})

			if err := database.MarkMailSent(mail.ID); err != nil {
				log.Printf("Notifier: failed to mark mail as sent: %v", err)
			}
		}
	}
}

func StopNotifier() {
	log.Println("Stopping notifier service...")
}
