package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"wtfsiw/internal/ai"
	"wtfsiw/internal/cli"
	"wtfsiw/internal/config"
	"wtfsiw/internal/tmdb"
	"wtfsiw/internal/trakt"
	"wtfsiw/internal/tui"
)

var (
	numResults int
	plainMode  bool
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
	rootCmd.Flags().BoolVarP(&plainMode, "plain", "p", false, "disable animations and colors (for scripting)")
}

func initConfig() {
	if err := config.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}
}

func runMain(cmd *cobra.Command, args []string) error {
	// Initialize AI provider (required for both modes)
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

	// If query provided as argument, run non-interactive CLI mode
	if len(args) > 0 {
		return runNonInteractive(aiProvider, tmdbClient, args[0], plainMode)
	}

	// Otherwise launch interactive chat TUI
	return runChatMode(aiProvider, tmdbClient)
}

func runChatMode(aiProvider ai.Provider, tmdbClient *tmdb.Client) error {
	// Initialize chat provider
	chatProvider, err := ai.NewChatProvider()
	if err != nil {
		return fmt.Errorf("failed to initialize chat provider: %w", err)
	}

	// Initialize Trakt client (optional - if not configured, some features unavailable)
	traktClient, err := trakt.NewClient()
	if err != nil {
		// Trakt not configured
		traktClient = nil
	}

	// Launch chat TUI
	return tui.RunChat(chatProvider, tmdbClient, traktClient, aiProvider)
}

func runNonInteractive(aiProvider ai.Provider, tmdbClient *tmdb.Client, query string, plain bool) error {
	ctx := context.Background()

	// Validate and clamp numResults to 1-10
	if numResults < 1 {
		numResults = 1
	} else if numResults > 10 {
		numResults = 10
	}

	// Print header
	if plain {
		fmt.Printf("Searching for: %s\n\n", query)
	} else {
		cli.PrintHeader(query)
	}

	var recommendations []ai.Recommendation
	var summary string

	// Helper to run with optional spinner
	runWithSpinner := func(msg string, fn func() error) error {
		if plain {
			fmt.Println(msg + "...")
			return fn()
		}
		spinner := cli.NewSpinner(msg + "...")
		spinner.Start()
		err := fn()
		if err != nil {
			spinner.Stop()
			cli.PrintError(err)
			return err
		}
		spinner.StopWithMessage(msg + " done")
		return nil
	}

	if tmdbClient == nil {
		// AI-only mode
		var resp *ai.RecommendationResponse
		err := runWithSpinner("Asking AI for recommendations", func() error {
			var err error
			resp, err = aiProvider.GetRecommendations(ctx, query, numResults)
			return err
		})
		if err != nil {
			return nil
		}
		recommendations = resp.Recommendations
		summary = resp.Summary
	} else {
		// TMDb mode
		var params *ai.SearchParams
		err := runWithSpinner("Analyzing with AI", func() error {
			var err error
			params, err = aiProvider.ExtractSearchParams(ctx, query)
			return err
		})
		if err != nil {
			return nil
		}

		var resp *tmdb.SearchResponse
		err = runWithSpinner("Searching TMDb", func() error {
			var err error
			resp, err = tmdbClient.Discover(params)
			return err
		})
		if err != nil {
			return nil
		}

		_ = runWithSpinner("Fetching providers", func() error {
			tmdbClient.EnrichWithProviders(resp.Results)
			return nil
		})

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
		summary = fmt.Sprintf("Found %d matches", len(recommendations))
	}

	fmt.Println()

	if len(recommendations) == 0 {
		if plain {
			fmt.Println("No results found.")
		} else {
			cli.PrintNoResults()
		}
		return nil
	}

	// Print results
	if plain {
		fmt.Printf("%s\n\n", summary)
		for i, rec := range recommendations {
			mediaType := "MOVIE"
			if rec.MediaType == "tv" {
				mediaType = "TV"
			}
			fmt.Printf("%d. [%s] %s (%s) - %.1f/10\n", i+1, mediaType, rec.Title, rec.Year, rec.Rating)
			if len(rec.Providers) > 0 {
				fmt.Printf("   Watch on: %s\n", joinStrings(rec.Providers, ", "))
			}
			if rec.WhyWatch != "" {
				fmt.Printf("   Why: %s\n", rec.WhyWatch)
			}
			fmt.Println()
		}
	} else {
		cli.PrintSummary(summary)
		cli.PrintDivider()
		fmt.Println()
		cli.PrintResults(recommendations, true)
	}

	return nil
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
