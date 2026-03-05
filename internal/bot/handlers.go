package bot

import (
	"context"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/martins6/opencode-telegram/internal/config"
)

var userSettings map[int64]*UserSettings
var cfg *config.Config

type UserSettings struct {
	Agent     string
	Model     string
	Provider  string
	Workspace string
}

func SetConfig(c *config.Config) {
	cfg = c
	if userSettings == nil {
		userSettings = make(map[int64]*UserSettings)
	}
}

func DefaultHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	userID := update.Message.From.ID

	if !isUserAllowed(userID) {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "You are not authorized to use this bot.",
		})
		return
	}

	log.Printf("Received message from user %d: %s", userID, update.Message.Text)
}

func isUserAllowed(userID int64) bool {
	if cfg == nil || len(cfg.Bot.AllowedUsers) == 0 {
		return true
	}

	for _, allowedID := range cfg.Bot.AllowedUsers {
		if allowedID == userID {
			return true
		}
	}
	return false
}
