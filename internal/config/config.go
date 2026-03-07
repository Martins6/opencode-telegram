package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	Bot       BotConfig       `mapstructure:"bot"`
	Workspace WorkspaceConfig `mapstructure:"workspace"`
	OpenCode  OpenCodeConfig  `mapstructure:"opencode"`
	Defaults  DefaultsConfig  `mapstructure:"defaults"`
}

type BotConfig struct {
	Token        string   `mapstructure:"token"`
	AllowedUsers []string `mapstructure:"allowed_users"`
}

type WorkspaceConfig struct {
	Path string `mapstructure:"path"`
}

type OpenCodeConfig struct {
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
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
	viper.SetDefault("bot.allowed_users", []string{})
	viper.SetDefault("workspace.path", filepath.Join(homeDir, ".opencode-telegram"))
	viper.SetDefault("opencode.port", "4096")
	viper.SetDefault("opencode.password", "")
	viper.SetDefault("defaults.agent", "telegram-agent")
	viper.SetDefault("defaults.model", "MiniMax-M2.5")
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
