# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Run Commands

```bash
# Build
go build -o wtfsiw .

# Run TUI (interactive mode)
./wtfsiw

# Run CLI (non-interactive mode)
./wtfsiw "dark psychological thriller" -n 5

# Test configuration
./wtfsiw config
./wtfsiw config set ai.provider openai
./wtfsiw config set ai.openai_api_key YOUR_KEY
```

## Architecture

**wtfsiw** (What The Fuck Should I Watch?) is a Go CLI tool that uses AI to recommend movies/TV shows based on natural language queries.

### Core Flow
1. User provides a natural language query
2. AI provider extracts search parameters OR generates recommendations directly
3. If TMDb is configured: searches TMDb API with extracted params, enriches with streaming providers
4. If TMDb not configured: uses AI-only mode with direct recommendations
5. Results displayed in TUI (interactive) or stdout (non-interactive)

### Key Components

- **`cmd/`** - Cobra CLI commands
  - `root.go`: Main entry, handles both TUI and non-interactive modes
  - `config.go`: Configuration management subcommand

- **`internal/ai/`** - AI provider abstraction
  - `provider.go`: Interface + `Recommendation` struct (unified format for both TMDb and AI results)
  - `claude.go`: Anthropic Claude implementation
  - `openai.go`: OpenAI implementation (uses JSON response format)

- **`internal/tmdb/`** - TMDb API client
  - `client.go`: HTTP client, `Media` struct, genre mappings
  - `search.go`: Discover/search logic, converts AI `SearchParams` to TMDb queries
  - `providers.go`: Streaming provider enrichment (where to watch)

- **`internal/tui/`** - Bubble Tea TUI
  - `app.go`: Main model with states (Input → Loading → Results → Detail)
  - `styles.go`: Lip Gloss styles, star rating rendering

- **`internal/config/`** - Viper configuration
  - Reads from `~/.config/wtfsiw/config.yaml` and environment variables

### AI Provider Interface

Both providers implement:
```go
type Provider interface {
    ExtractSearchParams(ctx, query) (*SearchParams, error)  // For TMDb mode
    GetRecommendations(ctx, query, count) (*RecommendationResponse, error)  // For AI-only mode
}
```

### Dual Mode Operation

- **TMDb mode**: AI extracts keywords/genres → TMDb search → real ratings/providers
- **AI-only mode**: AI generates recommendations directly (when TMDb API key not set)

Config priority: config file → environment variables (`ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, `TMDB_API_KEY`)
