package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/martins6/acolyte/internal/database"
	"github.com/pelletier/go-toml/v2"
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

var ErrConfigNotFound = errors.New("singleton config not found")

func SingletonConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".acolyte", "config.toml"), nil
}

func LoadIfExists(cfgFile string) (*Config, error) {
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

	target := cfgFile
	if target == "" {
		target = filepath.Join(defaultConfigPath, "config.toml")
	}
	if _, err := os.Stat(target); err != nil {
		if os.IsNotExist(err) {
			return nil, ErrConfigNotFound
		}
		return nil, err
	}

	viper.SetConfigFile(target)
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}
	globalConfig = &config
	return &config, nil
}

func LoadSingleton() (*Config, error) {
	path, err := SingletonConfigPath()
	if err != nil {
		return nil, err
	}
	return LoadIfExists(path)
}

func WriteWorkspacePath(workspace string) error {
	path, err := SingletonConfigPath()
	if err != nil {
		return err
	}

	abs, err := filepath.Abs(workspace)
	if err != nil {
		return fmt.Errorf("resolve absolute workspace: %w", err)
	}

	loadMu.Lock()
	defer loadMu.Unlock()

	viper.SetConfigFile(path)
	viper.SetConfigType("toml")

	current := &Config{}
	if _, err := os.Stat(path); err == nil {
		if err := viper.ReadInConfig(); err != nil {
			return fmt.Errorf("read singleton config: %w", err)
		}
		if err := viper.Unmarshal(current); err != nil {
			return fmt.Errorf("parse singleton config: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("ensure singleton directory: %w", err)
	}

	if current.Bot.Token == "" {
		current.Bot.Token = viper.GetString("bot.token")
	}
	if current.Bot.AllowedUserID == "" {
		current.Bot.AllowedUserID = viper.GetString("bot.allowed_user_id")
	}
	if current.Defaults.Agent == "" {
		current.Defaults.Agent = viper.GetString("defaults.agent")
	}
	if current.Defaults.Model == "" {
		current.Defaults.Model = viper.GetString("defaults.model")
	}
	if current.Defaults.Provider == "" {
		current.Defaults.Provider = viper.GetString("defaults.provider")
	}

	current.Workspace.Path = abs

	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	enc.SetIndentTables(true)
	if err := enc.Encode(current); err != nil {
		return fmt.Errorf("encode singleton config: %w", err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write singleton config: %w", err)
	}

	viper.Set("workspace.path", abs)
	if globalConfig != nil {
		globalConfig.Workspace.Path = abs
	}
	return nil
}
