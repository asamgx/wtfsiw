package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"wtfsiw/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage wtfsiw configuration",
	Long: `Manage wtfsiw configuration including API keys and preferences.

Configuration file location: ~/.config/wtfsiw/config.yaml

Required API keys:
  - TMDb API key (free): https://developer.themoviedb.org/
  - Claude API key: https://console.anthropic.com/
  - OpenAI API key (optional): https://platform.openai.com/

You can also set these via environment variables:
  - TMDB_API_KEY
  - ANTHROPIC_API_KEY
  - OPENAI_API_KEY`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Configuration file:", config.GetConfigPath())
		fmt.Println()
		fmt.Println("Current settings:")
		cfg := config.Get()
		fmt.Printf("  AI Provider: %s\n", cfg.AI.Provider)
		fmt.Printf("  Claude API Key: %s\n", maskKey(cfg.AI.ClaudeAPIKey))
		fmt.Printf("  OpenAI API Key: %s\n", maskKey(cfg.AI.OpenAIAPIKey))
		fmt.Printf("  TMDb API Key: %s\n", maskKey(cfg.TMDB.APIKey))
		fmt.Printf("  Trakt Client ID: %s\n", maskKey(cfg.Trakt.ClientID))
		fmt.Printf("  Trakt Access Token: %s\n", maskKey(cfg.Trakt.AccessToken))
		fmt.Printf("  Region: %s\n", cfg.Preferences.Region)
		fmt.Printf("  Language: %s\n", cfg.Preferences.Language)
		fmt.Println()
		fmt.Println("Use 'wtfsiw config set <key> <value>' to update settings")
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a configuration value.

Available keys:
  ai.provider          - AI provider to use (claude or openai)
  ai.claude_api_key    - Anthropic Claude API key
  ai.openai_api_key    - OpenAI API key
  tmdb.api_key         - TMDb API key
  trakt.client_id      - Trakt API client ID
  trakt.client_secret  - Trakt API client secret
  trakt.access_token   - Trakt access token (use 'wtfsiw trakt auth' instead)
  preferences.region   - Region for streaming providers (e.g., US, GB)
  preferences.language - Language code (e.g., en, es)
  preferences.min_rating - Minimum rating filter (0-10)
  preferences.max_results - Maximum results to show

Examples:
  wtfsiw config set tmdb.api_key abc123
  wtfsiw config set ai.provider openai
  wtfsiw config set trakt.client_id YOUR_CLIENT_ID`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		if err := config.Set(key, value); err != nil {
			return fmt.Errorf("failed to set config: %w", err)
		}

		fmt.Printf("Set %s = %s\n", key, maskKey(value))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configSetCmd)
}

func maskKey(key string) string {
	if key == "" {
		return "(not set)"
	}
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "..." + key[len(key)-4:]
}
