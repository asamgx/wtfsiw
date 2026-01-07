package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	AI          AIConfig          `mapstructure:"ai"`
	TMDB        TMDBConfig        `mapstructure:"tmdb"`
	Trakt       TraktConfig       `mapstructure:"trakt"`
	Preferences PreferencesConfig `mapstructure:"preferences"`
}

type AIConfig struct {
	Provider     string `mapstructure:"provider"`
	ClaudeAPIKey string `mapstructure:"claude_api_key"`
	OpenAIAPIKey string `mapstructure:"openai_api_key"`
}

type TMDBConfig struct {
	APIKey string `mapstructure:"api_key"`
}

type TraktConfig struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	AccessToken  string `mapstructure:"access_token"`
}

type PreferencesConfig struct {
	DefaultType string  `mapstructure:"default_type"`
	Region      string  `mapstructure:"region"`
	Language    string  `mapstructure:"language"`
	MinRating   float64 `mapstructure:"min_rating"`
	MaxResults  int     `mapstructure:"max_results"`
}

var cfg *Config

func Init() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(home, ".config", "wtfsiw")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDir)
	viper.AddConfigPath(".")

	// Set defaults
	viper.SetDefault("ai.provider", "claude")
	viper.SetDefault("preferences.default_type", "all")
	viper.SetDefault("preferences.region", "US")
	viper.SetDefault("preferences.language", "en")
	viper.SetDefault("preferences.min_rating", 0.0)
	viper.SetDefault("preferences.max_results", 10)

	// Bind environment variables
	viper.BindEnv("ai.claude_api_key", "ANTHROPIC_API_KEY")
	viper.BindEnv("ai.openai_api_key", "OPENAI_API_KEY")
	viper.BindEnv("tmdb.api_key", "TMDB_API_KEY")
	viper.BindEnv("trakt.client_id", "TRAKT_CLIENT_ID")
	viper.BindEnv("trakt.client_secret", "TRAKT_CLIENT_SECRET")
	viper.BindEnv("trakt.access_token", "TRAKT_ACCESS_TOKEN")

	// Read config file if exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("failed to read config: %w", err)
		}
	}

	cfg = &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

func Get() *Config {
	if cfg == nil {
		cfg = &Config{}
	}
	return cfg
}

func GetConfigPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "wtfsiw", "config.yaml")
}

func Save() error {
	return viper.WriteConfigAs(GetConfigPath())
}

func Set(key, value string) error {
	viper.Set(key, value)
	return Save()
}

// GetSessionsDir returns the path to the sessions directory
func GetSessionsDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "wtfsiw", "sessions")
}
