# wtfsiw

**What The Fuck Should I Watch?** - An AI-powered CLI tool that helps you find something to watch based on natural language queries.

Describe what you're in the mood for, and get personalized movie and TV show recommendations with ratings, streaming providers, and AI-generated explanations of why each title matches your request.

## Features

- **Natural language search** - "something dark and psychological like Breaking Bad"
- **Dual AI backend** - Supports both Claude (Anthropic) and OpenAI
- **Rich TUI** - Interactive terminal interface with Bubble Tea
- **Pretty CLI** - Animated spinners, colors, styled output (or plain mode for scripting)
- **Streaming providers** - Shows where to watch (Netflix, HBO, etc.)
- **Star ratings** - Visual ratings with ★★★★☆ display
- **AI-only mode** - Works without TMDb for quick recommendations

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
# Get recommendations with animated output
./wtfsiw "feel-good comedy from the 90s"

# Limit number of results
./wtfsiw "Korean thriller" -n 5

# Plain output for scripting (no colors/animations)
./wtfsiw "mind-bending sci-fi like Inception" -n 3 --plain
```

CLI mode features animated spinners, colored output, and styled results. Use `--plain` or `-p` to disable all formatting for piping to other commands.

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
| Trakt | Optional | [trakt.tv/oauth/applications](https://trakt.tv/oauth/applications) (free) |

**Without TMDb**: Works in AI-only mode with estimated ratings and provider guesses.

**With TMDb**: Real ratings, vote counts, and accurate streaming provider data.

**With Trakt**: Access your watchlist, watch history, and ratings for personalized recommendations.

## Trakt Integration

Trakt integration allows wtfsiw to access your personal watch data for better recommendations.

### Step 1: Create a Trakt API Application

1. Go to [trakt.tv/oauth/applications](https://trakt.tv/oauth/applications)
2. Click **"New Application"**
3. Fill in the form:
   - **Name**: `wtfsiw` (or any name you prefer)
   - **Description**: `CLI tool for movie/TV recommendations`
   - **Redirect URI**: `urn:ietf:wg:oauth:2.0:oob` (required for CLI apps)
   - **Permissions**: Leave unchecked (read-only access is sufficient)
4. Click **"Save App"**
5. You'll receive a **Client ID** and **Client Secret**

### Step 2: Configure wtfsiw

```bash
# Set your Trakt credentials
./wtfsiw config set trakt.client_id YOUR_CLIENT_ID
./wtfsiw config set trakt.client_secret YOUR_CLIENT_SECRET
```

Or use environment variables:
```bash
export TRAKT_CLIENT_ID=your_client_id
export TRAKT_CLIENT_SECRET=your_client_secret
```

### Step 3: Authenticate

```bash
./wtfsiw trakt auth
```

This starts the Device OAuth flow:
1. You'll see a URL and a code (e.g., `Go to: https://trakt.tv/activate` and `Enter code: A1B2C3D4`)
2. Open the URL in your browser
3. Enter the code when prompted
4. Authorize the application
5. Return to the terminal - authentication completes automatically

Your access token is saved to the config file. You only need to do this once.

### Step 4: Use Trakt Features

```bash
# Check Trakt connection status
./wtfsiw trakt

# View your watchlist
./wtfsiw trakt watchlist

# View only movies or shows
./wtfsiw trakt watchlist movies
./wtfsiw trakt watchlist shows
```

### Trakt Commands

| Command | Description |
|---------|-------------|
| `wtfsiw trakt` | Show Trakt connection status |
| `wtfsiw trakt auth` | Authenticate with Trakt |
| `wtfsiw trakt watchlist` | View all watchlist items |
| `wtfsiw trakt watchlist movies` | View only movies |
| `wtfsiw trakt watchlist shows` | View only TV shows |

### Environment Variables

| Variable | Description |
|----------|-------------|
| `TRAKT_CLIENT_ID` | Your Trakt application's Client ID |
| `TRAKT_CLIENT_SECRET` | Your Trakt application's Client Secret |
| `TRAKT_ACCESS_TOKEN` | OAuth access token (auto-saved after auth) |

## Tech Stack

- [Cobra](https://github.com/spf13/cobra) + [Viper](https://github.com/spf13/viper) - CLI framework
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) + [Lip Gloss](https://github.com/charmbracelet/lipgloss) - TUI
- [Anthropic SDK](https://github.com/anthropics/anthropic-sdk-go) / [OpenAI SDK](https://github.com/sashabaranov/go-openai) - AI
- [TMDb API](https://developer.themoviedb.org) - Movie/TV data

## License

MIT
