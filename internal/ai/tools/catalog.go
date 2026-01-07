package tools

// Catalog contains all available tools for the chat assistant
var Catalog = []ToolDefinition{
	{
		Name:        "search_media",
		Description: "Search for movies or TV shows with various filters. Use this when the user wants to discover content based on preferences like genre, year, rating, language, or streaming service.",
		Parameters: []ToolParameter{
			{
				Name:        "keywords",
				Type:        "array",
				Items:       &ToolParameter{Type: "string"},
				Description: "Search keywords or terms",
			},
			{
				Name:        "genres",
				Type:        "array",
				Items:       &ToolParameter{Type: "string"},
				Description: "Genre filters: action, comedy, drama, horror, thriller, sci-fi, romance, documentary, animation, fantasy, mystery, crime, war, western, family, history",
			},
			{
				Name:        "media_type",
				Type:        "string",
				Enum:        []string{"movie", "tv", "all"},
				Description: "Type of media to search for",
			},
			{
				Name:        "year_from",
				Type:        "integer",
				Description: "Start year for release date filter",
			},
			{
				Name:        "year_to",
				Type:        "integer",
				Description: "End year for release date filter",
			},
			{
				Name:        "min_rating",
				Type:        "number",
				Description: "Minimum rating (0-10 scale)",
			},
			{
				Name:        "language",
				Type:        "string",
				Description: "Original language ISO code (e.g., 'en', 'ko', 'ja', 'fr', 'es')",
			},
			{
				Name:        "providers",
				Type:        "array",
				Items:       &ToolParameter{Type: "string"},
				Description: "Streaming providers to filter by: Netflix, Disney Plus, HBO Max, Amazon Prime Video, Hulu, Apple TV Plus, etc.",
			},
			{
				Name:        "actors",
				Type:        "array",
				Items:       &ToolParameter{Type: "string"},
				Description: "Actor names to filter by",
			},
			{
				Name:        "studios",
				Type:        "array",
				Items:       &ToolParameter{Type: "string"},
				Description: "Production studios: Pixar, A24, Marvel, Studio Ghibli, etc.",
			},
		},
	},
	{
		Name:        "get_media_details",
		Description: "Get detailed information about a specific movie or TV show by its TMDb ID. Use this when you need more information about a specific title.",
		Parameters: []ToolParameter{
			{
				Name:        "id",
				Type:        "integer",
				Required:    true,
				Description: "The TMDb ID of the movie or TV show",
			},
			{
				Name:        "media_type",
				Type:        "string",
				Required:    true,
				Enum:        []string{"movie", "tv"},
				Description: "Whether it's a movie or TV show",
			},
		},
	},
	{
		Name:        "get_streaming_providers",
		Description: "Get streaming availability for a specific movie or TV show. Shows where it can be watched, rented, or purchased.",
		Parameters: []ToolParameter{
			{
				Name:        "id",
				Type:        "integer",
				Required:    true,
				Description: "The TMDb ID of the movie or TV show",
			},
			{
				Name:        "media_type",
				Type:        "string",
				Required:    true,
				Enum:        []string{"movie", "tv"},
				Description: "Whether it's a movie or TV show",
			},
		},
	},
	{
		Name:        "get_similar",
		Description: "Find movies or TV shows similar to a given title. Use this when the user likes a specific title and wants similar recommendations.",
		Parameters: []ToolParameter{
			{
				Name:        "id",
				Type:        "integer",
				Required:    true,
				Description: "The TMDb ID of the movie or TV show",
			},
			{
				Name:        "media_type",
				Type:        "string",
				Required:    true,
				Enum:        []string{"movie", "tv"},
				Description: "Whether it's a movie or TV show",
			},
		},
	},
	{
		Name:        "search_by_title",
		Description: "Search for a movie or TV show by its title. Use this to find the TMDb ID of a specific title the user mentions.",
		Parameters: []ToolParameter{
			{
				Name:        "title",
				Type:        "string",
				Required:    true,
				Description: "The title to search for",
			},
		},
	},
	{
		Name:        "get_trakt_watchlist",
		Description: "Get items from the user's Trakt watchlist. Only works if the user has connected their Trakt account.",
		Parameters: []ToolParameter{
			{
				Name:        "media_type",
				Type:        "string",
				Enum:        []string{"movies", "shows", ""},
				Description: "Filter by media type, or leave empty for all",
			},
		},
	},
	{
		Name:        "get_trakt_history",
		Description: "Get the user's recently watched items from Trakt. Only works if the user has connected their Trakt account.",
		Parameters: []ToolParameter{
			{
				Name:        "media_type",
				Type:        "string",
				Enum:        []string{"movies", "shows", ""},
				Description: "Filter by media type, or leave empty for all",
			},
			{
				Name:        "limit",
				Type:        "integer",
				Description: "Maximum number of items to return (default 20)",
			},
		},
	},
	{
		Name:        "generate_recommendations",
		Description: "Generate AI recommendations directly based on a description. Use this when TMDb search filters aren't sufficient or for subjective/mood-based requests.",
		Parameters: []ToolParameter{
			{
				Name:        "description",
				Type:        "string",
				Required:    true,
				Description: "Description of what the user is looking for",
			},
			{
				Name:        "count",
				Type:        "integer",
				Description: "Number of recommendations to generate (default 5)",
			},
		},
	},
}
