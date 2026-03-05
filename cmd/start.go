package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/martins6/opencode-telegram/internal/bot"
	"github.com/martins6/opencode-telegram/internal/config"
	"github.com/martins6/opencode-telegram/internal/logger"
	"github.com/martins6/opencode-telegram/internal/opencode"
	"github.com/martins6/opencode-telegram/internal/workspace"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start bot + OpenCode server in workspace",
	Long: `Starts the Telegram bot and OpenCode server in the configured workspace.

The bot will:
1. Start the OpenCode server on the configured port
2. Initialize the Telegram bot
3. Handle incoming messages and media

Press Ctrl+C to stop both the bot and server gracefully.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load("")
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		workspacePath := cfg.Workspace.Path
		if workspacePath == "" {
			homeDir, _ := os.UserHomeDir()
			workspacePath = homeDir + "/.opencode-telegram"
		}

		if err := workspace.ValidateWorkspace(workspacePath); err != nil {
			log.Printf("Workspace not found, creating: %v", err)
			if err := workspace.CreateTemplate(workspacePath); err != nil {
				return fmt.Errorf("failed to create workspace: %w", err)
			}
		}

		if err := logger.Initialize(workspacePath); err != nil {
			log.Printf("Warning: Failed to initialize logger: %v", err)
		}

		log.Println("Starting OpenCode server...")
		server := opencode.NewServer(cfg.OpenCode.Port, cfg.OpenCode.Password, workspacePath)
		if err := server.Start(workspacePath); err != nil {
			return fmt.Errorf("failed to start OpenCode server: %w", err)
		}

		log.Println("Initializing Telegram bot...")
		bot.SetConfig(cfg)
		telegramBot, err := bot.Initialize(cfg)
		if err != nil {
			server.Stop()
			return fmt.Errorf("failed to initialize bot: %w", err)
		}

		if telegramBot != nil {
			bot.RegisterHandlers(telegramBot)
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		if err := bot.Start(ctx, telegramBot); err != nil {
			server.Stop()
			return fmt.Errorf("failed to start bot: %w", err)
		}

		log.Println("Bot is running. Press Ctrl+C to stop.")

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down...")

		if telegramBot != nil {
			log.Println("Stopping Telegram bot...")
		}

		server.Stop()

		logger.Close()

		log.Println("Shutdown complete")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
