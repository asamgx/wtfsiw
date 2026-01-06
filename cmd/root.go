package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"wtfsiw/internal/ai"
	"wtfsiw/internal/config"
	"wtfsiw/internal/tmdb"
	"wtfsiw/internal/tui"
)

var (
	numResults int
)

var rootCmd = &cobra.Command{
	Use:   "wtfsiw [query]",
	Short: "What The Fuck Should I Watch? - AI-powered movie/TV recommendations",
	Long: `wtfsiw helps you find something to watch using AI-powered natural language search.

Describe what you're in the mood for in plain English, and wtfsiw will
analyze your request and find matching movies and TV shows.

Examples:
  wtfsiw "something dark and psychological like Breaking Bad"
  wtfsiw "feel-good comedy from the 90s"
  wtfsiw "Korean thriller, recent, highly rated" -n 5
  wtfsiw  # launches interactive mode`,
	Args: cobra.MaximumNArgs(1),
	RunE: runMain,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.Flags().IntVarP(&numResults, "number", "n", 10, "number of recommendations (1-10)")
}

func initConfig() {
	if err := config.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}
}

func runMain(cmd *cobra.Command, args []string) error {
	// Initialize AI provider (required)
	aiProvider, err := ai.NewProvider()
	if err != nil {
		return fmt.Errorf("failed to initialize AI: %w\n\nRun 'wtfsiw config' for setup instructions", err)
	}

	// Initialize TMDb client (optional - if not configured, use AI-only mode)
	tmdbClient, err := tmdb.NewClient()
	if err != nil {
		// TMDb not configured, will use AI-only mode
		tmdbClient = nil
	}

	// If query provided as argument, run non-interactive mode
	if len(args) > 0 {
		return runNonInteractive(aiProvider, tmdbClient, args[0])
	}

	// Otherwise launch TUI
	return tui.Run(aiProvider, tmdbClient)
}

func runNonInteractive(aiProvider ai.Provider, tmdbClient *tmdb.Client, query string) error {
	ctx := context.Background()

	// Validate and clamp numResults to 1-10
	if numResults < 1 {
		numResults = 1
	} else if numResults > 10 {
		numResults = 10
	}

	fmt.Printf("🎬 Searching for: %s\n\n", query)

	var recommendations []ai.Recommendation
	var summary string

	if tmdbClient == nil {
		// AI-only mode
		fmt.Println("Using AI-only mode (TMDb not configured)...")
		resp, err := aiProvider.GetRecommendations(ctx, query, numResults)
		if err != nil {
			return fmt.Errorf("AI recommendation failed: %w", err)
		}
		recommendations = resp.Recommendations
		summary = resp.Summary
	} else {
		// TMDb mode
		fmt.Println("Analyzing with AI...")
		params, err := aiProvider.ExtractSearchParams(ctx, query)
		if err != nil {
			return fmt.Errorf("AI analysis failed: %w", err)
		}

		fmt.Println("Searching TMDb...")
		resp, err := tmdbClient.Discover(params)
		if err != nil {
			return fmt.Errorf("search failed: %w", err)
		}

		tmdbClient.EnrichWithProviders(resp.Results)

		// Limit to requested number
		results := resp.Results
		if len(results) > numResults {
			results = results[:numResults]
		}

		for _, media := range results {
			providers := make([]string, len(media.Providers))
			for j, p := range media.Providers {
				providers[j] = p.Name
			}
			recommendations = append(recommendations, ai.Recommendation{
				Title:     media.GetDisplayTitle(),
				Year:      media.GetDisplayYear(),
				MediaType: media.MediaType,
				Rating:    media.VoteAverage,
				Overview:  media.Overview,
				Providers: providers,
				VoteCount: media.VoteCount,
			})
		}
		summary = fmt.Sprintf("Keywords: %s", strings.Join(params.Keywords, ", "))
	}

	if len(recommendations) == 0 {
		fmt.Println("No results found.")
		return nil
	}

	fmt.Printf("📋 %s\n\n", summary)

	for i, rec := range recommendations {
		mediaType := "🎬"
		if rec.MediaType == "tv" {
			mediaType = "📺"
		}

		stars := renderStarsText(rec.Rating)

		fmt.Printf("%d. %s %s (%s)\n", i+1, mediaType, rec.Title, rec.Year)
		fmt.Printf("   %s %.1f/10\n", stars, rec.Rating)

		if len(rec.Providers) > 0 {
			fmt.Printf("   📍 %s\n", strings.Join(rec.Providers, ", "))
		}

		if rec.WhyWatch != "" {
			fmt.Printf("   💡 %s\n", rec.WhyWatch)
		}

		if rec.Overview != "" {
			overview := rec.Overview
			if len(overview) > 150 {
				overview = overview[:147] + "..."
			}
			fmt.Printf("   %s\n", overview)
		}
		fmt.Println()
	}

	return nil
}

func renderStarsText(rating float64) string {
	stars := int(rating / 2)
	result := ""
	for i := 0; i < 5; i++ {
		if i < stars {
			result += "★"
		} else {
			result += "☆"
		}
	}
	return result
}
