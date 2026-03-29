package config

import (
	"os"
	"path/filepath"

	"github.com/martins6/opencode-telegram/internal/database"
	"github.com/spf13/viper"
)

type Config struct {
	Bot       BotConfig       `mapstructure:"bot"`
	Workspace WorkspaceConfig `mapstructure:"workspace"`
	Defaults  DefaultsConfig  `mapstructure:"defaults"`
}

type BotConfig struct {
	Token         string `mapstructure:"token"`
	AllowedUserID string `mapstructure:"allowed_user_id"`
}

type WorkspaceConfig struct {
	Path string `mapstructure:"path"`
}

type DefaultsConfig struct {
	Agent    string `mapstructure:"agent"`
	Model    string `mapstructure:"model"`
	Provider string `mapstructure:"provider"`
}

var globalConfig *Config

func Load(cfgFile string) (*Config, error) {
	viper.SetConfigType("toml")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	defaultConfigPath := filepath.Join(homeDir, ".opencode-telegram")
	viper.SetDefault("bot.token", "")
	viper.SetDefault("bot.allowed_user_id", "")
	viper.SetDefault("workspace.path", filepath.Join(homeDir, ".opencode-telegram"))
	viper.SetDefault("defaults.agent", "telegram-agent")
	viper.SetDefault("defaults.model", "MiniMax-M2.7")
	viper.SetDefault("defaults.provider", "minimax-coding-plan")

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(defaultConfigPath)
		viper.SetConfigName("config")
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if err := os.MkdirAll(defaultConfigPath, 0755); err != nil {
				return nil, err
			}
			configPath := filepath.Join(defaultConfigPath, "config.toml")
			if err := viper.SafeWriteConfigAs(configPath); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	globalConfig = &config
	return &config, nil
}

func Get() *Config {
	return globalConfig
}

func GetAllowedUserChatID() int64 {
	if globalConfig == nil {
		return 0
	}
	if globalConfig.Bot.AllowedUserID == "" {
		return 0
	}
	chatID, err := database.GetResolvedChatID(globalConfig.Bot.AllowedUserID)
	if err != nil {
		return 0
	}
	return chatID
}
