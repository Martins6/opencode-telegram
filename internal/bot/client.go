package bot

import (
	"context"
	"log"

	"github.com/go-telegram/bot"
	"github.com/martins6/acolyte/internal/config"
)

var Bot *bot.Bot

func Initialize(cfg *config.Config) (*bot.Bot, error) {
	// Note: bot token and allowed_user_id are read once at startup and
	// require a daemon restart to change. Hot-reloading the token would
	// require tearing down and re-establishing the Telegram client
	// connection, which is intentionally not supported. Runtime-mutable
	// configs (defaults.agent/model/provider, bot.timezone) are reloaded
	// on every message handler / notifier tick via config.Load("").
	if cfg.Bot.Token == "" {
		return nil, nil
	}

	opts := []bot.Option{
		bot.WithDefaultHandler(DefaultHandler),
	}

	b, err := bot.New(cfg.Bot.Token, opts...)
	if err != nil {
		return nil, err
	}

	Bot = b
	return b, nil
}

func Start(ctx context.Context, b *bot.Bot) error {
	if b == nil {
		log.Println("Telegram bot not initialized (no token configured)")
		return nil
	}

	go func() {
		b.Start(ctx)
	}()

	return nil
}
