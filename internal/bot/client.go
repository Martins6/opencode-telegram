package bot

import (
	"context"
	"log"

	"github.com/go-telegram/bot"
	"github.com/martins6/opencode-telegram/internal/config"
)

var Bot *bot.Bot

func Initialize(cfg *config.Config) (*bot.Bot, error) {
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
