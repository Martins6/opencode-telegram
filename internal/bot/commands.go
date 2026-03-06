package bot

import (
	"context"
	"log"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/martins6/opencode-telegram/internal/session"
)

func RegisterHandlers(b *bot.Bot) {
	b.RegisterHandler(bot.HandlerTypeMessageText, "/start", bot.MatchTypeExact, handleStart)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/help", bot.MatchTypeExact, handleHelp)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/set-agent", bot.MatchTypeExact, handleSetAgent)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/set-model", bot.MatchTypeExact, handleSetModel)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/set-provider", bot.MatchTypeExact, handleSetProvider)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/workspace", bot.MatchTypeExact, handleWorkspace)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/reset", bot.MatchTypeExact, handleReset)
	b.RegisterHandler(bot.HandlerTypeMessageText, "/new-session", bot.MatchTypeExact, handleNewSession)
}

func handleStart(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	if !isUserAllowed(update.Message.From.ID, update.Message.From.Username) {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "You are not authorized to use this bot.",
		})
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Welcome to OpenCode Telegram Agent! Use /help to see available commands.",
	})
}

func handleHelp(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	if !isUserAllowed(update.Message.From.ID, update.Message.From.Username) {
		return
	}

	helpText := `
Available commands:
/set-agent <name> - Set active agent (e.g., coder, planner)
/set-model <model> - Set LLM model (e.g., claude-sonnet-4-5)
/set-provider <provider> - Set LLM provider (e.g., anthropic, openai)
/workspace <path> - Set workspace path
/reset - Reset conversation history
/new-session - Start a fresh conversation
/help - Show this help message
`

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   helpText,
	})
}

func handleSetAgent(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	userID := update.Message.From.ID
	username := update.Message.From.Username
	if !isUserAllowed(userID, username) {
		return
	}

	args := update.Message.Text
	if len(args) < 12 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Usage: /set-agent <agent-name>",
		})
		return
	}

	agent := args[12:]
	if userSettings[userID] == nil {
		userSettings[userID] = &UserSettings{}
	}
	userSettings[userID].Agent = agent

	log.Printf("User %d set agent to: %s", userID, agent)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Agent set to: " + agent,
	})
}

func handleSetModel(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	userID := update.Message.From.ID
	username := update.Message.From.Username
	if !isUserAllowed(userID, username) {
		return
	}

	args := update.Message.Text
	if len(args) < 11 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Usage: /set-model <model-name>",
		})
		return
	}

	model := args[11:]
	if userSettings[userID] == nil {
		userSettings[userID] = &UserSettings{}
	}
	userSettings[userID].Model = model

	log.Printf("User %d set model to: %s", userID, model)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Model set to: " + model,
	})
}

func handleSetProvider(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	userID := update.Message.From.ID
	username := update.Message.From.Username
	if !isUserAllowed(userID, username) {
		return
	}

	args := update.Message.Text
	if len(args) < 14 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Usage: /set-provider <provider-name>",
		})
		return
	}

	provider := args[14:]
	if userSettings[userID] == nil {
		userSettings[userID] = &UserSettings{}
	}
	userSettings[userID].Provider = provider

	log.Printf("User %d set provider to: %s", userID, provider)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Provider set to: " + provider,
	})
}

func handleWorkspace(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	userID := update.Message.From.ID
	username := update.Message.From.Username
	if !isUserAllowed(userID, username) {
		return
	}

	args := update.Message.Text
	if len(args) < 11 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Usage: /workspace <path>",
		})
		return
	}

	workspace := args[11:]
	if userSettings[userID] == nil {
		userSettings[userID] = &UserSettings{}
	}
	userSettings[userID].Workspace = workspace

	log.Printf("User %d set workspace to: %s", userID, workspace)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Workspace set to: " + workspace,
	})
}

func handleReset(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	userID := update.Message.From.ID
	username := update.Message.From.Username
	if !isUserAllowed(userID, username) {
		return
	}

	delete(userSettings, userID)

	log.Printf("User %d reset conversation", userID)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Conversation reset. Your settings have been cleared.",
	})
}

func handleNewSession(ctx context.Context, b *bot.Bot, update *models.Update) {
	if update.Message == nil {
		return
	}

	userID := update.Message.From.ID
	username := update.Message.From.Username
	if !isUserAllowed(userID, username) {
		return
	}

	sessionMgr := session.GetManager()
	workspace := cfg.Workspace.Path

	userSession, err := sessionMgr.GetSession(userID, workspace)
	if err != nil {
		log.Printf("Error getting session for user %d: %v", userID, err)
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Error: Could not reset session. Please try again.",
		})
		return
	}

	sessionMgr.ResetConversation(userSession)

	log.Printf("User %d started new session", userID)

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "New session started. Your next message will start a fresh conversation.",
	})
}
