#!/usr/bin/env python3
"""
Stremio to Trakt Importer

Extract media details from Stremio HTML library export and convert to Trakt import format.

Usage:
    python stremio_to_trakt.py input.html > output.json
    python stremio_to_trakt.py input.html -o output.json
    cat input.html | python stremio_to_trakt.py > output.json

Supports movies, TV shows, and individual episodes. Extracts IMDB IDs, watch progress,
and watched status for seamless import into Trakt.
"""

from __future__ import annotations

import argparse
import json
import re
import sys
from datetime import datetime
from html.parser import HTMLParser
from typing import TextIO
from urllib.parse import unquote

__version__ = "1.0.0"


class StremioHTMLParser(HTMLParser):
    def __init__(self):
        super().__init__()
        self.items = []
        self.current_item = None
        self.in_meta_item = False
        self.in_title_label = False
        self.capture_text = False
        self.current_text_field = None

    def handle_starttag(self, tag, attrs):
        attrs_dict = dict(attrs)

        # Look for meta item containers (the <a> tags with media info)
        if tag == 'a' and 'class' in attrs_dict:
            classes = attrs_dict.get('class', '')
            if 'meta-item-container' in classes:
                self.in_meta_item = True
                href = attrs_dict.get('href', '')
                title = attrs_dict.get('title', '')

                # Extract IMDB ID and type from href
                # Format: #/detail/series/tt5875444 or #/detail/movie/tt12820516
                imdb_match = re.search(r'#/detail/(movie|series)/(tt\d+)', href)

                if imdb_match:
                    media_type = imdb_match.group(1)
                    imdb_id = imdb_match.group(2)

                    # Check for episode info in href
                    # Format: tt5875444:5:3 means season 5, episode 3
                    episode_match = re.search(r'(tt\d+):(\d+):(\d+)', href)

                    self.current_item = {
                        'imdb_id': imdb_id,
                        'title': title,
                        'type': 'show' if media_type == 'series' else 'movie',
                        'href': href,
                        'is_watched': False,
                        'progress': 0,
                        'season': None,
                        'episode': None,
                        'poster_url': None,
                        'poster_shape': None,
                        'year': None,
                        'duration': None,
                        'episode_title': None,
                        'release_info': None,
                    }

                    if episode_match:
                        self.current_item['season'] = int(episode_match.group(2))
                        self.current_item['episode'] = int(episode_match.group(3))

                    # Store all data attributes
                    for key, value in attrs_dict.items():
                        if key.startswith('data-'):
                            self.current_item[key] = value

        # Check for watched icon
        if self.current_item and tag == 'div' and 'class' in attrs_dict:
            if 'watched-icon-layer' in attrs_dict.get('class', ''):
                self.current_item['is_watched'] = True

        # Check for progress bar (multiple possible class name patterns)
        if self.current_item and tag == 'div' and 'class' in attrs_dict:
            classes = attrs_dict.get('class', '')
            if 'progress-bar' in classes or 'progressBar' in classes.lower():
                style = attrs_dict.get('style', '')
                progress_match = re.search(r'width:\s*([\d.]+)%', style)
                if progress_match:
                    self.current_item['progress'] = float(progress_match.group(1))

        # Extract poster image from background-image style
        if self.current_item and tag == 'div' and 'class' in attrs_dict:
            classes = attrs_dict.get('class', '')
            style = attrs_dict.get('style', '')
            if 'poster' in classes.lower() or 'thumbnail' in classes.lower() or 'image' in classes.lower():
                # Extract URL from background-image: url("...")
                url_match = re.search(r'background-image:\s*url\(["\']?([^"\')\s]+)["\']?\)', style)
                if url_match:
                    self.current_item['poster_url'] = unquote(url_match.group(1))
                # Check for poster shape class
                if 'poster-shape' in classes:
                    shape_match = re.search(r'poster-shape-(\w+)', classes)
                    if shape_match:
                        self.current_item['poster_shape'] = shape_match.group(1)

        # Extract poster from img tag
        if self.current_item and tag == 'img':
            src = attrs_dict.get('src', '')
            if src and not self.current_item['poster_url']:
                self.current_item['poster_url'] = unquote(src)
            alt = attrs_dict.get('alt', '')
            if alt and not self.current_item['title']:
                self.current_item['title'] = alt

        # Look for title/name labels
        if self.current_item and tag in ('div', 'span', 'p') and 'class' in attrs_dict:
            classes = attrs_dict.get('class', '')
            if 'title' in classes.lower() or 'name' in classes.lower() or 'label' in classes.lower():
                self.capture_text = True
                self.current_text_field = 'title_text'

        # Look for year/release info
        if self.current_item and tag in ('div', 'span', 'p') and 'class' in attrs_dict:
            classes = attrs_dict.get('class', '')
            if 'year' in classes.lower() or 'release' in classes.lower() or 'date' in classes.lower():
                self.capture_text = True
                self.current_text_field = 'release_info'

        # Look for duration info
        if self.current_item and tag in ('div', 'span', 'p') and 'class' in attrs_dict:
            classes = attrs_dict.get('class', '')
            if 'duration' in classes.lower() or 'runtime' in classes.lower() or 'time' in classes.lower():
                self.capture_text = True
                self.current_text_field = 'duration'

        # Look for episode title
        if self.current_item and tag in ('div', 'span', 'p') and 'class' in attrs_dict:
            classes = attrs_dict.get('class', '')
            if 'episode' in classes.lower() and 'title' in classes.lower():
                self.capture_text = True
                self.current_text_field = 'episode_title'

        # Extract any inline styles that might contain useful info
        if self.current_item and 'style' in attrs_dict:
            style = attrs_dict.get('style', '')
            # Look for any background images we might have missed
            if 'background-image' in style and not self.current_item['poster_url']:
                url_match = re.search(r'background-image:\s*url\(["\']?([^"\')\s]+)["\']?\)', style)
                if url_match:
                    self.current_item['poster_url'] = unquote(url_match.group(1))

    def handle_data(self, data):
        if self.current_item and self.capture_text and data.strip():
            text = data.strip()
            if self.current_text_field == 'release_info':
                self.current_item['release_info'] = text
                # Try to extract year
                year_match = re.search(r'\b(19|20)\d{2}\b', text)
                if year_match:
                    self.current_item['year'] = int(year_match.group(0))
            elif self.current_text_field == 'duration':
                self.current_item['duration'] = text
            elif self.current_text_field == 'episode_title':
                self.current_item['episode_title'] = text
            elif self.current_text_field == 'title_text':
                # Only update if we don't have a title yet
                if not self.current_item['title']:
                    self.current_item['title'] = text

    def handle_endtag(self, tag):
        if tag in ('div', 'span', 'p'):
            self.capture_text = False
            self.current_text_field = None

        if tag == 'a' and self.in_meta_item:
            if self.current_item:
                # Clean up None values for cleaner output
                self.current_item = {k: v for k, v in self.current_item.items() if v is not None}
                self.items.append(self.current_item)
                self.current_item = None
            self.in_meta_item = False


def parse_stremio_html(html_content: str) -> list[dict]:
    """Parse Stremio HTML and extract media items."""
    parser = StremioHTMLParser()
    parser.feed(html_content)
    return parser.items


def convert_to_trakt_format(
    items: list[dict],
    add_watchlist: bool = True,
    mark_unknown_dates: bool = True,
) -> list[dict]:
    """
    Convert parsed items to Trakt import format.

    Args:
        items: List of parsed media items
        add_watchlist: Add watchlisted_at to all items (default True)
        mark_unknown_dates: Use 'unknown' for watched_at if no date available

    Returns:
        List of items in Trakt import format
    """
    trakt_items = []

    for item in items:
        trakt_item = {
            'imdb_id': item.get('imdb_id'),
            'type': item.get('type'),
        }

        # Add title
        if item.get('title'):
            trakt_item['title'] = item['title']

        # Add episode info if present
        if item.get('season') is not None and item.get('episode') is not None:
            trakt_item['type'] = 'episode'
            trakt_item['season'] = item['season']
            trakt_item['episode'] = item['episode']

        # Add watchlisted_at for all items (this is the most important field)
        if add_watchlist:
            trakt_item['watchlisted_at'] = datetime.now().isoformat() + 'Z'

        # Determine watched status
        # Consider watched if progress > 90% or has watched icon
        is_watched = item.get('is_watched', False) or item.get('progress', 0) > 90

        if is_watched:
            if mark_unknown_dates:
                trakt_item['watched_at'] = 'unknown'
            else:
                trakt_item['watched_at'] = datetime.now().isoformat() + 'Z'

        # Add extra fields from HTML
        if item.get('href'):
            trakt_item['_href'] = item['href']
        if item.get('poster_url'):
            trakt_item['_poster_url'] = item['poster_url']
        if item.get('progress', 0) > 0:
            trakt_item['_progress'] = item['progress']
        if item.get('year'):
            trakt_item['_year'] = item['year']
        if item.get('duration'):
            trakt_item['_duration'] = item['duration']
        if item.get('episode_title'):
            trakt_item['_episode_title'] = item['episode_title']

        trakt_items.append(trakt_item)

    return trakt_items


def format_for_trakt_import(items: list[dict]) -> list[dict]:
    """
    Format items for Trakt import using their exact schema.

    Returns the properly formatted JSON structure for Trakt.
    """
    formatted = []

    for item in items:
        entry = {
            'imdb_id': item['imdb_id'],
            'type': item['type'],
        }

        # Add title
        if item.get('title'):
            entry['title'] = item['title']

        # Add watched_at if present
        if 'watched_at' in item:
            entry['watched_at'] = item['watched_at']

        # Add watchlisted_at if present
        if 'watchlisted_at' in item:
            entry['watchlisted_at'] = item['watchlisted_at']

        # Add rating if present (1-10)
        if 'rating' in item:
            entry['rating'] = item['rating']
            if 'rated_at' in item:
                entry['rated_at'] = item['rated_at']

        # Add episode info for episodes
        if item['type'] == 'episode':
            entry['season'] = item.get('season')
            entry['episode'] = item.get('episode')

        # Add all extra fields (prefixed with _)
        for key, value in item.items():
            if key.startswith('_') and value is not None:
                entry[key] = value

        formatted.append(entry)

    return formatted


def main() -> None:
    parser = argparse.ArgumentParser(
        description='Extract media from Stremio HTML and convert to Trakt import format',
        epilog='For more info: https://gist.github.com/asamgx/622f920a7a22f562b89b9b2f1ee3326e',
    )
    parser.add_argument(
        '-v', '--version',
        action='version',
        version=f'%(prog)s {__version__}',
    )
    parser.add_argument(
        'input',
        nargs='?',
        type=argparse.FileType('r'),
        default=sys.stdin,
        help='Input HTML file (or stdin)',
    )
    parser.add_argument(
        '-o', '--output',
        type=argparse.FileType('w'),
        default=sys.stdout,
        help='Output JSON file (default: stdout)',
    )
    parser.add_argument(
        '--no-watchlist',
        action='store_true',
        help='Do not add watchlisted_at field (by default all items get watchlisted_at)',
    )
    parser.add_argument(
        '--use-current-date',
        action='store_true',
        help='Use current date instead of "unknown" for watched_at',
    )
    parser.add_argument(
        '--watched-only',
        action='store_true',
        help='Only include items that are marked as watched',
    )
    parser.add_argument(
        '--min-progress',
        type=float,
        default=0,
        help='Minimum progress percentage to include (default: 0)',
    )

    args = parser.parse_args()

    # Read HTML content
    html_content = args.input.read()

    if not html_content.strip():
        print("Error: Input is empty. Please provide a valid Stremio HTML export.", file=sys.stderr)
        sys.exit(1)

    # Parse HTML
    items = parse_stremio_html(html_content)

    if not items:
        print("Error: No media items found in the HTML.", file=sys.stderr)
        print("", file=sys.stderr)
        print("Make sure you're exporting from Stremio's Library page.", file=sys.stderr)
        print("The HTML should contain elements with 'meta-item-container' class.", file=sys.stderr)
        sys.exit(1)

    # Count by type
    movies = sum(1 for i in items if i.get('type') == 'movie')
    shows = sum(1 for i in items if i.get('type') == 'show' and not i.get('season'))
    episodes = sum(1 for i in items if i.get('season') is not None)
    watched = sum(1 for i in items if i.get('is_watched') or i.get('progress', 0) > 90)

    print(f"Found {len(items)} items: {movies} movies, {shows} shows, {episodes} episodes ({watched} watched)", file=sys.stderr)

    # Filter by progress if specified (before conversion)
    if args.min_progress > 0:
        items = [
            item for item in items
            if item.get('progress', 0) >= args.min_progress
        ]

    # Filter to watched only if specified (before conversion)
    if args.watched_only:
        items = [
            item for item in items
            if item.get('is_watched', False) or item.get('progress', 0) > 90
        ]

    # Convert to Trakt format
    trakt_items = convert_to_trakt_format(
        items,
        add_watchlist=not args.no_watchlist,
        mark_unknown_dates=not args.use_current_date
    )

    # Format for Trakt import
    output = format_for_trakt_import(trakt_items)

    # Write output
    json.dump(output, args.output, indent=2)
    args.output.write('\n')

    print(f"Exported {len(output)} items to Trakt format", file=sys.stderr)
    if args.output != sys.stdout:
        print(f"Output written to: {args.output.name}", file=sys.stderr)


if __name__ == '__main__':
    main()
