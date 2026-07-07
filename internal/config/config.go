package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/martins6/acolyte/internal/database"
	"github.com/spf13/viper"
)

var ErrTimezoneNotConfigured = errors.New("timezone not configured")

type Config struct {
	Bot       BotConfig       `mapstructure:"bot"`
	Workspace WorkspaceConfig `mapstructure:"workspace"`
	Defaults  DefaultsConfig  `mapstructure:"defaults"`
}

type BotConfig struct {
	Token         string `mapstructure:"token"`
	AllowedUserID string `mapstructure:"allowed_user_id"`
	Timezone      string `mapstructure:"timezone"`
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

var (
	loadMu          sync.Mutex
	setDefaultsOnce sync.Once
)

func Load(cfgFile string) (*Config, error) {
	loadMu.Lock()
	defer loadMu.Unlock()

	viper.SetConfigType("toml")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	defaultConfigPath := filepath.Join(homeDir, ".acolyte")
	setDefaultsOnce.Do(func() {
		viper.SetDefault("bot.token", "")
		viper.SetDefault("bot.allowed_user_id", "")
		viper.SetDefault("bot.timezone", "")
		viper.SetDefault("workspace.path", defaultConfigPath)
		viper.SetDefault("defaults.agent", "acolyte")
		viper.SetDefault("defaults.model", "MiniMax-M3")
		viper.SetDefault("defaults.provider", "minimax-coding-plan")
	})

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

func GetLocation() (*time.Location, error) {
	if globalConfig == nil || globalConfig.Bot.Timezone == "" {
		return nil, ErrTimezoneNotConfigured
	}
	loc, err := time.LoadLocation(globalConfig.Bot.Timezone)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone %q: %w", globalConfig.Bot.Timezone, err)
	}
	return loc, nil
}

func SetForTest(c *Config) { globalConfig = c }
