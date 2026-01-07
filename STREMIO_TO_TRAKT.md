# The Naive/Easy/Working Way to Extract Stremio Library (Watchlist)

Convert your Stremio library to Trakt import format.

A Python script to convert your Stremio library HTML export into Trakt-compatible JSON format for easy import of your watchlist and watch history.

## Why This Tool?

Stremio doesn't have a direct export feature, but you can save your library page as HTML. This script parses that HTML and extracts:

- Movies and TV shows with IMDB IDs
- Individual episodes with season/episode numbers
- Watch progress percentages
- Watched status

The output is formatted for Trakt's JSON import, letting you migrate your library in minutes.

## Prerequisites

- Python 3.6 or higher
- No external dependencies (uses only standard library)

## How to Export from Stremio

### Step 1: Open Your Library (Critical!)

1. Open **Stremio Web** in your browser: [web.stremio.com](https://web.stremio.com)
2. Log in to your account
3. Navigate to your Library: [web.stremio.com/#/library?sort=lastwatched](https://web.stremio.com/#/library?sort=lastwatched)

### Step 2: Load All Items (Important!)

> **⚠️ Critical Step:** Stremio uses infinite scroll - items are loaded dynamically as you scroll down. The HTML only contains what's currently loaded in the browser.

**You MUST scroll through your entire library before saving:**

1. Start scrolling down slowly
2. Wait for new items to load as you scroll
3. Keep scrolling until you reach the very bottom
4. If you have a large library, this may take a minute or two
5. You'll know you're done when scrolling no longer loads new items

**Tip:** You can hold `Page Down` or use `End` key repeatedly, but make sure to pause and let items load between key presses.

### Step 3: Save the HTML

Use either method:

#### Method A: Save Page As (Recommended)

1. Press `Ctrl+S` (or `Cmd+S` on Mac)
2. Choose "Webpage, Complete" or "HTML Only"
3. Save as `library.html`

#### Method B: Browser Dev Tools

1. Open Developer Tools (`F12` or `Ctrl+Shift+I`)
2. In the Elements tab, right-click on the `<html>` tag
3. Select **Copy** → **Copy outerHTML**
4. Paste into a new text file
5. Save as `library.html`

## Usage

```bash
# Basic usage - outputs to stdout
python3 stremio_to_trakt.py library.html

# Save to file
python3 stremio_to_trakt.py library.html -o trakt_import.json

# From stdin
cat library.html | python3 stremio_to_trakt.py > trakt_import.json

# Check version
python3 stremio_to_trakt.py --version
```

## Options

| Option | Description |
|--------|-------------|
| `-o, --output FILE` | Write output to file instead of stdout |
| `-v, --version` | Show version number |
| `--no-watchlist` | Don't add `watchlisted_at` field (by default all items get it) |
| `--use-current-date` | Use current ISO date instead of "unknown" for `watched_at` |
| `--watched-only` | Only export items marked as watched (>90% progress or watched icon) |
| `--min-progress N` | Only include items with N% or more watch progress |

### Examples

```bash
# Export only items you've watched
python3 stremio_to_trakt.py library.html --watched-only -o watched.json

# Export items with at least 50% progress
python3 stremio_to_trakt.py library.html --min-progress 50 -o in_progress.json

# Export watchlist only (no watched_at dates)
python3 stremio_to_trakt.py library.html --no-watchlist -o watchlist.json
```

## Output Format

The output follows Trakt's import JSON format:

```json
[
  {
    "imdb_id": "tt5875444",
    "type": "episode",
    "title": "Better Call Saul",
    "watchlisted_at": "2024-01-07T12:00:00Z",
    "watched_at": "unknown",
    "season": 5,
    "episode": 3
  },
  {
    "imdb_id": "tt12820516",
    "type": "movie",
    "title": "Prey",
    "watchlisted_at": "2024-01-07T12:00:00Z"
  }
]
```

### Fields

**Trakt standard fields:**
- `imdb_id` - IMDB ID (e.g., tt5875444)
- `type` - `movie`, `show`, or `episode`
- `title` - Media title
- `watchlisted_at` - ISO 8601 timestamp
- `watched_at` - ISO 8601 timestamp or `"unknown"`
- `season` / `episode` - Episode info (for episodes only)

**Extra fields (prefixed with `_`):**
These are included for reference but ignored by Trakt during import:
- `_href` - Stremio URL path
- `_poster_url` - Poster image URL
- `_progress` - Watch progress percentage
- `_year` - Release year

## Import to Trakt

1. Go to [trakt.tv/apps/import](https://trakt.tv/apps/import)
2. Select **JSON** as the file type
3. Upload your `trakt_import.json` file
4. Review the items and select what to import (watchlist, history, etc.)
5. Confirm the import

**Note:** Trakt safely ignores fields it doesn't recognize (like the `_` prefixed fields), so they won't cause any issues.

## Troubleshooting

### "No media items found in the HTML"

- Make sure you saved the page from the **Library** section, not the home page
- Ensure you scrolled down to load all items before saving
- The HTML needs elements with `meta-item-container` class

### Items are missing

- Stremio lazy-loads content. Scroll all the way down before saving.
- Some items may not have IMDB IDs and will be skipped.

### Watch progress not detected

- Progress bars may use different class names in different Stremio versions
- The script looks for elements with `progress-bar` or `progressBar` classes

### Episodes showing as "show" type

- Episode info comes from the URL pattern `tt1234567:2:5` (show:season:episode)
- If this pattern isn't in the URL, it's treated as a show entry

## License

MIT
