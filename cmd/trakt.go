package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"wtfsiw/internal/config"
	"wtfsiw/internal/trakt"
)

var traktCmd = &cobra.Command{
	Use:   "trakt",
	Short: "Manage Trakt integration",
	Long: `Manage Trakt integration for personalized recommendations.

Trakt tracks your watch history, ratings, and watchlist to help
personalize movie and TV show recommendations.

Get started:
  1. Create a Trakt API app at https://trakt.tv/oauth/applications
  2. Set your client credentials:
     wtfsiw config set trakt.client_id YOUR_CLIENT_ID
     wtfsiw config set trakt.client_secret YOUR_CLIENT_SECRET
  3. Authenticate:
     wtfsiw trakt auth`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.Get()
		fmt.Println("Trakt Integration Status")
		fmt.Println()
		fmt.Printf("  Client ID: %s\n", maskKey(cfg.Trakt.ClientID))
		fmt.Printf("  Client Secret: %s\n", maskKey(cfg.Trakt.ClientSecret))
		fmt.Printf("  Access Token: %s\n", maskKey(cfg.Trakt.AccessToken))
		fmt.Println()
		if cfg.Trakt.AccessToken == "" {
			fmt.Println("Not authenticated. Run 'wtfsiw trakt auth' to connect your account.")
		} else {
			fmt.Println("Authenticated. Run 'wtfsiw trakt watchlist' to view your watchlist.")
		}
	},
}

var traktAuthCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with Trakt",
	Long: `Authenticate with Trakt using the Device OAuth flow.

This will display a code that you enter at https://trakt.tv/activate
to authorize wtfsiw to access your Trakt account.

Prerequisites:
  - Client ID must be configured (wtfsiw config set trakt.client_id YOUR_ID)
  - Client Secret must be configured (wtfsiw config set trakt.client_secret YOUR_SECRET)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Get()

		if cfg.Trakt.ClientID == "" {
			return fmt.Errorf("Trakt client ID not configured. Run: wtfsiw config set trakt.client_id YOUR_CLIENT_ID")
		}
		if cfg.Trakt.ClientSecret == "" {
			return fmt.Errorf("Trakt client secret not configured. Run: wtfsiw config set trakt.client_secret YOUR_CLIENT_SECRET")
		}

		fmt.Println("Requesting device code...")
		deviceCode, err := trakt.GetDeviceCode(cfg.Trakt.ClientID)
		if err != nil {
			return fmt.Errorf("failed to get device code: %w", err)
		}

		fmt.Println()
		fmt.Printf("Go to: %s\n", deviceCode.VerificationURL)
		fmt.Printf("Enter code: %s\n", deviceCode.UserCode)
		fmt.Println()
		fmt.Println("Waiting for authorization...")

		token, err := trakt.PollForToken(
			cfg.Trakt.ClientID,
			cfg.Trakt.ClientSecret,
			deviceCode.DeviceCode,
			deviceCode.Interval,
		)
		if err != nil {
			return fmt.Errorf("authorization failed: %w", err)
		}

		// Save the access token
		if err := config.Set("trakt.access_token", token.AccessToken); err != nil {
			return fmt.Errorf("failed to save access token: %w", err)
		}

		fmt.Println()
		fmt.Println("Success! Your Trakt account is now connected.")
		return nil
	},
}

var traktWatchlistCmd = &cobra.Command{
	Use:   "watchlist [movies|shows]",
	Short: "View your Trakt watchlist",
	Long: `View items in your Trakt watchlist.

Usage:
  wtfsiw trakt watchlist          # Show all items
  wtfsiw trakt watchlist movies   # Show only movies
  wtfsiw trakt watchlist shows    # Show only TV shows`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := trakt.NewClient()
		if err != nil {
			return err
		}

		mediaType := ""
		if len(args) > 0 {
			mediaType = args[0]
		}

		items, err := client.GetWatchlist(mediaType)
		if err != nil {
			return fmt.Errorf("failed to get watchlist: %w", err)
		}

		if len(items) == 0 {
			fmt.Println("Your watchlist is empty.")
			return nil
		}

		fmt.Printf("Found %d items in your watchlist:\n\n", len(items))
		for i, item := range items {
			title := item.GetDisplayTitle()
			year := item.GetDisplayYear()
			rating := item.GetRating()
			genres := item.GetGenres()
			runtime := item.GetRuntime()
			overview := item.GetOverview()

			typeLabel := item.Type
			if typeLabel == "movie" {
				typeLabel = "Movie"
			} else if typeLabel == "show" {
				typeLabel = "TV Show"
			}

			// Title line
			fmt.Printf("%d. %s (%d) [%s]\n", i+1, title, year, typeLabel)

			// Rating and runtime
			if rating > 0 || runtime > 0 {
				fmt.Printf("   ")
				if rating > 0 {
					fmt.Printf("Rating: %.1f/10", rating)
				}
				if rating > 0 && runtime > 0 {
					fmt.Printf(" | ")
				}
				if runtime > 0 {
					if typeLabel == "TV Show" {
						fmt.Printf("Runtime: %dm/ep", runtime)
					} else {
						fmt.Printf("Runtime: %dm", runtime)
					}
				}
				fmt.Println()
			}

			// Genres
			if len(genres) > 0 {
				fmt.Printf("   Genres: %s\n", joinStrings(genres, ", "))
			}

			// Overview (truncated)
			if overview != "" {
				if len(overview) > 150 {
					overview = overview[:147] + "..."
				}
				fmt.Printf("   %s\n", overview)
			}

			// IDs for reference
			if item.Movie != nil {
				fmt.Printf("   IDs: IMDB=%s TMDB=%d\n", item.Movie.IDs.IMDB, item.Movie.IDs.TMDB)
			} else if item.Show != nil {
				fmt.Printf("   IDs: IMDB=%s TMDB=%d TVDB=%d\n", item.Show.IDs.IMDB, item.Show.IDs.TMDB, item.Show.IDs.TVDB)
			}

			fmt.Println()
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(traktCmd)
	traktCmd.AddCommand(traktAuthCmd)
	traktCmd.AddCommand(traktWatchlistCmd)
}
