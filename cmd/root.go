package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var cfgFile string

var versionValue = "dev"

func SetVersion(v string) {
	versionValue = v
}

var rootCmd = &cobra.Command{
	Use:   "acolyte",
	Short: "Telegram bot gateway for OpenCode AI agent",
	Long: `A Telegram bot that acts as a gateway to an OpenCode server,
allowing users to interact with the OpenCode agent directly from Telegram.

The bot handles text, media (images, audio, files), and provides 
configuration via slash commands.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err == nil {
		return
	}
	var exit ExitCode
	if errors.As(err, &exit) {
		os.Exit(int(exit))
	}
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ~/.acolyte/config.toml)")
}
