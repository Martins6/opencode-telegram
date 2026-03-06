package bot

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/martins6/opencode-telegram/internal/config"
	"github.com/martins6/opencode-telegram/internal/logger"
	"github.com/martins6/opencode-telegram/internal/opencode"
	"github.com/martins6/opencode-telegram/internal/session"
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
	username := update.Message.From.Username

	if !isUserAllowed(userID, username) {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "You are not authorized to use this bot.",
		})
		return
	}

	log.Printf("Received message from user %d: %s", userID, update.Message.Text)

	if strings.TrimSpace(update.Message.Text) == "" {
		logger.Log(logger.DEBUG, userID, "Received empty or whitespace-only message, skipping")
		return
	}

	logger.Log(logger.INPUT, userID, fmt.Sprintf("Message: %s", truncate(update.Message.Text, 100)))

	chatID := update.Message.Chat.ID
	workspace := cfg.Workspace.Path

	sessionMgr := session.GetManager()
	userSession, err := sessionMgr.GetSession(userID, workspace)
	if err != nil {
		log.Printf("Error getting session for user %d: %v", userID, err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   "Error: Could not initialize session. Please try again.",
		})
		return
	}

	settings := userSettings[userID]
	agent := cfg.Defaults.Agent
	model := cfg.Defaults.Model
	provider := cfg.Defaults.Provider

	if agent == "" {
		agent = "build"
	}

	if settings != nil {
		if settings.Agent != "" {
			agent = settings.Agent
		}
		if settings.Model != "" {
			model = settings.Model
		}
		if settings.Provider != "" {
			provider = settings.Provider
		}
	}

	if model == "" {
		model = "MiniMax-M2.5"
	}
	if provider == "" {
		provider = "minimax-coding-plan"
	}

	runner := opencode.NewRunner(workspace, agent, model, provider)

	var sessionID string
	if userSession.OpenCodeID != "" && !userSession.IsNewSession {
		sessionID = userSession.OpenCodeID
		log.Printf("Continuing session %s for user %d", sessionID, userID)
	} else {
		log.Printf("Starting new session for user %d", userID)
	}

	log.Printf("Sending message to OpenCode for user %d (session: %s, agent: %s, model: %s, provider: %s)",
		userID, sessionID, agent, model, provider)

	result, err := runner.Execute(sessionID, update.Message.Text)
	if err != nil {
		log.Printf("Error running opencode for user %d: %v", userID, err)

		errorMsg := "Error: Could not get response from OpenCode."
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "invalid") {
			errorMsg = fmt.Sprintf("Error: %v. Please check your model/provider settings with /set-model or /set-provider command.", err)
		} else if strings.Contains(err.Error(), "no valid output") {
			errorMsg = "Error: OpenCode returned no output. Please check your model/provider settings."
		} else {
			errorMsg = fmt.Sprintf("Error: Could not get response from OpenCode. %v", err)
		}

		logger.Log(logger.ERROR, userID, fmt.Sprintf("Failed to run opencode: %v", err))
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   errorMsg,
		})
		return
	}

	if result.SessionID != "" {
		userSession.OpenCodeID = result.SessionID
		userSession.IsNewSession = false
		sessionMgr.UpdateSession(userSession)
		log.Printf("Updated session ID to %s for user %d", result.SessionID, userID)
	}

	responseText := result.ResponseText
	if responseText == "" {
		responseText = "No response from OpenCode."
	}

	logger.Log(logger.OUTPUT, userID, fmt.Sprintf("Response: %s", truncate(responseText, 100)))

	const maxMessageLength = 4000
	if len(responseText) > maxMessageLength {
		logger.Log(logger.DEBUG, userID, fmt.Sprintf("Response too long (%d chars), splitting into chunks", len(responseText)))
		sendLongMessage(ctx, b, chatID, responseText, maxMessageLength)
	} else {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   responseText,
		})
	}
}

func sendLongMessage(ctx context.Context, b *bot.Bot, chatID int64, text string, maxLen int) {
	for len(text) > 0 {
		chunkLen := maxLen
		if len(text) < chunkLen {
			chunkLen = len(text)
		} else {
			lastSpace := strings.LastIndex(text[:chunkLen], " ")
			if lastSpace > 0 {
				chunkLen = lastSpace
			}
		}

		chunk := text[:chunkLen]
		_, err := b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: chatID,
			Text:   chunk,
		})
		if err != nil {
			log.Printf("Error sending message chunk: %v", err)
			return
		}

		text = text[chunkLen:]
		if len(text) > 0 && text[0] == ' ' {
			text = text[1:]
		}
	}
}

func extractResponseText(resp *opencode.MessageResponse) string {
	if resp == nil {
		return ""
	}

	var text string
	for _, part := range resp.Parts {
		if part.Type == "text" {
			text += part.Text
		}
	}
	return text
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

func isUserAllowed(userID int64, username string) bool {
	if cfg == nil || len(cfg.Bot.AllowedUsers) == 0 {
		return true
	}

	for _, allowed := range cfg.Bot.AllowedUsers {
		if parsedID, err := strconv.ParseInt(allowed, 10, 64); err == nil {
			if parsedID == userID {
				return true
			}
		}
		if username != "" && allowed == username {
			return true
		}
	}
	return false
}
