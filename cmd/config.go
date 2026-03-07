package cmd

import (
	"fmt"
	"os"

	"github.com/martins6/opencode-telegram/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management",
	Long:  "Manage configuration settings for the Telegram bot",
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		homeDir, _ := os.UserHomeDir()
		configPath := fmt.Sprintf("%s/.opencode-telegram/config.toml", homeDir)

		viper.SetConfigType("toml")
		viper.SetConfigFile(configPath)
		if err := viper.ReadInConfig(); err != nil {
			if !os.IsNotExist(err) {
				return fmt.Errorf("failed to read config: %w", err)
			}
		}

		viper.Set(key, value)

		if err := viper.WriteConfigAs(configPath); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("Set %s = %s\n", key, value)
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := viper.Get(key)
		if value == nil {
			return fmt.Errorf("key %s not found", key)
		}
		fmt.Printf("%s = %v\n", key, value)
		return nil
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration values",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load("")
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}
		fmt.Printf("Bot Token: %s\n", cfg.Bot.Token)
		fmt.Printf("Allowed Users: %v\n", cfg.Bot.AllowedUsers)
		fmt.Printf("Workspace Path: %s\n", cfg.Workspace.Path)
		fmt.Printf("OpenCode Port: %s\n", cfg.OpenCode.Port)
		fmt.Printf("Default Agent: %s\n", cfg.Defaults.Agent)
		fmt.Printf("Default Model: %s\n", cfg.Defaults.Model)
		fmt.Printf("Default Provider: %s\n", cfg.Defaults.Provider)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configListCmd)
	rootCmd.AddCommand(configCmd)
}
