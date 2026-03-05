package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "opencode-telegram",
	Short: "Telegram bot gateway for OpenCode AI agent",
	Long: `A Telegram bot that acts as a gateway to an OpenCode server,
allowing users to interact with the OpenCode agent directly from Telegram.

The bot handles text, media (images, audio, files), and provides 
configuration via slash commands.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ~/.opencode-telegram/config.toml)")
}
