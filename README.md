# wtfsiw

**What The Fuck Should I Watch?** - An AI-powered CLI tool that helps you find something to watch based on natural language queries.

Describe what you're in the mood for, and get personalized movie and TV show recommendations with ratings, streaming providers, and AI-generated explanations of why each title matches your request.

## Features

- **Natural language search** - "something dark and psychological like Breaking Bad"
- **Dual AI backend** - Supports both Claude (Anthropic) and OpenAI
- **Rich TUI** - Interactive terminal interface with Bubble Tea
- **Streaming providers** - Shows where to watch (Netflix, HBO, etc.)
- **Star ratings** - Visual ratings with ★★★★☆ display
- **AI-only mode** - Works without TMDb for quick recommendations
- **CLI mode** - Non-interactive output for scripting

## Installation

```bash
# Clone and build
git clone https://github.com/yourusername/wtfsiw.git
cd wtfsiw
go build -o wtfsiw .

# Or install directly
go install github.com/yourusername/wtfsiw@latest
```

## Quick Start

```bash
# Set up your AI provider (choose one)
./wtfsiw config set ai.provider openai
./wtfsiw config set ai.openai_api_key YOUR_OPENAI_KEY

# Or use Claude
./wtfsiw config set ai.provider claude
./wtfsiw config set ai.claude_api_key YOUR_ANTHROPIC_KEY

# Optional: Add TMDb for real ratings and streaming info
./wtfsiw config set tmdb.api_key YOUR_TMDB_KEY

# Run it!
./wtfsiw
```

## Usage

### Interactive Mode (TUI)

```bash
./wtfsiw
```

Launches a beautiful terminal UI where you can type queries and browse results.

### CLI Mode

```bash
# Get recommendations directly
./wtfsiw "feel-good comedy from the 90s"

# Limit number of results
./wtfsiw "Korean thriller" -n 5

# Sci-fi with specific count
./wtfsiw "mind-bending sci-fi like Inception" --number 3
```

### Example Output

```
🎬 Searching for: dark psychological thriller like breaking bad

📋 Recommendations for dark psychological thrillers similar to Breaking Bad.

1. 📺 Better Call Saul (2015-2022)
   ★★★★☆ 9.3/10
   📍 Netflix
   💡 It shares the same universe as Breaking Bad, exploring complex characters and moral dilemmas.

2. 📺 Ozark (2017-2022)
   ★★★★☆ 8.5/10
   📍 Netflix
   💡 Similar narrative of an ordinary person descending into the criminal underworld.

3. 🎬 Se7en (1995)
   ★★★★☆ 8.6/10
   📍 HBO Max
   💡 A classic psychological thriller that expertly builds tension.
```

## Configuration

Config file location: `~/.config/wtfsiw/config.yaml`

```yaml
ai:
  provider: claude  # or "openai"
  claude_api_key: sk-ant-...
  openai_api_key: sk-...

tmdb:
  api_key: your-tmdb-key

preferences:
  region: US
  language: en
```

### Environment Variables

You can also use environment variables:
- `ANTHROPIC_API_KEY` - Claude API key
- `OPENAI_API_KEY` - OpenAI API key
- `TMDB_API_KEY` - TMDb API key

### Commands

```bash
./wtfsiw config              # Show current configuration
./wtfsiw config set KEY VAL  # Set a config value
./wtfsiw --help              # Show help
```

## API Keys

| Provider | Required | Get it at |
|----------|----------|-----------|
| OpenAI or Claude | Yes (one) | [platform.openai.com](https://platform.openai.com) or [console.anthropic.com](https://console.anthropic.com) |
| TMDb | Optional | [developer.themoviedb.org](https://developer.themoviedb.org) (free) |

**Without TMDb**: Works in AI-only mode with estimated ratings and provider guesses.

**With TMDb**: Real ratings, vote counts, and accurate streaming provider data.

## Tech Stack

- [Cobra](https://github.com/spf13/cobra) + [Viper](https://github.com/spf13/viper) - CLI framework
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Lip Gloss](https://github.com/charmbracelet/lipgloss) - TUI
- [Anthropic SDK](https://github.com/anthropics/anthropic-sdk-go) / [OpenAI SDK](https://github.com/sashabaranov/go-openai) - AI
- [TMDb API](https://developer.themoviedb.org) - Movie/TV data

## License

MIT
