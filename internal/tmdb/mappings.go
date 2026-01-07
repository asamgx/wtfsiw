package tmdb

// WatchProviderMap maps common provider names to TMDb provider IDs
// Based on US region - IDs may vary by region
var WatchProviderMap = map[string]int{
	"netflix":            8,
	"amazon prime":       9,
	"amazon prime video": 9,
	"prime video":        9,
	"disney+":            337,
	"disney plus":        337,
	"hbo max":            384,
	"max":                1899, // HBO Max rebranded to Max
	"hulu":               15,
	"apple tv+":          350,
	"apple tv plus":      350,
	"paramount+":         531,
	"paramount plus":     531,
	"peacock":            386,
	"showtime":           37,
	"starz":              43,
	"criterion channel":  258,
	"mubi":               11,
	"shudder":            99,
	"tubi":               73,
	"pluto tv":           300,
	"crunchyroll":        283,
	"funimation":         269,
	"youtube":            192,
	"google play":        3,
	"vudu":               7,
	"fandango at home":   7, // Vudu rebranded
	"amazon video":       10,
	"apple tv":           2,
	"mgm+":               636,
	"mgm plus":           636,
	"amc+":               526,
	"amc plus":           526,
	"discovery+":         520,
	"discovery plus":     520,
	"bet+":               1759,
	"bet plus":           1759,
}

// StudioMap maps common studio names to TMDb company IDs
var StudioMap = map[string]int{
	// Major Studios
	"pixar":            3,
	"disney":           2,
	"walt disney":      2,
	"warner bros":      174,
	"warner brothers":  174,
	"universal":        33,
	"paramount":        4,
	"sony":             34,
	"sony pictures":    34,
	"columbia":         5,
	"20th century":     25,
	"20th century fox": 25,
	"fox":              25,
	"mgm":              8411,
	"lionsgate":        1632,
	"new line":         12,
	"new line cinema":  12,

	// Indie/Specialty
	"a24":          41077,
	"neon":         90733,
	"searchlight":  43,
	"fox searchlight": 43,
	"focus features": 10146,
	"annapurna":    130826,
	"blumhouse":    3172,
	"legendary":    923,

	// Animation
	"dreamworks":      521,
	"dreamworks animation": 521,
	"illumination":    6704,
	"laika":           11537,
	"blue sky":        9513,
	"studio ghibli":   10342,
	"ghibli":          10342,
	"toei":            5542,
	"toei animation":  5542,
	"madhouse":        3464,
	"bones":           2849,
	"mappa":           109939,
	"wit studio":      31673,
	"ufotable":        6140,
	"kyoto animation": 3518,

	// Superhero/Franchise
	"marvel":        420,
	"marvel studios": 420,
	"dc":            128064,
	"dc studios":   128064,
	"dc films":     128064,
	"lucasfilm":    1,

	// Horror
	"platinum dunes": 7220,
	"atomic monster": 76907,
}

// CertificationMap maps user-friendly names to TMDb certification values
var CertificationMap = map[string]string{
	// Movies (US)
	"g":      "G",
	"pg":     "PG",
	"pg-13":  "PG-13",
	"pg13":   "PG-13",
	"r":      "R",
	"nc-17":  "NC-17",
	"nc17":   "NC-17",

	// TV (US)
	"tv-y":   "TV-Y",
	"tvy":    "TV-Y",
	"tv-y7":  "TV-Y7",
	"tvy7":   "TV-Y7",
	"tv-g":   "TV-G",
	"tvg":    "TV-G",
	"tv-pg":  "TV-PG",
	"tvpg":   "TV-PG",
	"tv-14":  "TV-14",
	"tv14":   "TV-14",
	"tv-ma":  "TV-MA",
	"tvma":   "TV-MA",
}

// TVStatusMap maps user-friendly status names to TMDb status values
var TVStatusMap = map[string]int{
	"returning":        0, // Returning Series
	"returning series": 0,
	"ongoing":          0,
	"still airing":     0,
	"planned":          1, // Planned
	"in production":    2, // In Production
	"ended":            3, // Ended
	"completed":        3,
	"finished":         3,
	"canceled":         4, // Canceled
	"cancelled":        4,
	"pilot":            5, // Pilot
}

// SortByMap maps user-friendly sort names to TMDb sort values
var SortByMap = map[string]string{
	"popularity":    "popularity.desc",
	"popular":       "popularity.desc",
	"rating":        "vote_average.desc",
	"rated":         "vote_average.desc",
	"highest rated": "vote_average.desc",
	"release_date":  "primary_release_date.desc",
	"newest":        "primary_release_date.desc",
	"recent":        "primary_release_date.desc",
	"revenue":       "revenue.desc",
	"box office":    "revenue.desc",
	"title":         "title.asc",
	"alphabetical":  "title.asc",
}

// MonetizationTypeMap maps user-friendly names to TMDb values
var MonetizationTypeMap = map[string]string{
	"subscription": "flatrate",
	"flatrate":     "flatrate",
	"streaming":    "flatrate",
	"free":         "free",
	"rent":         "rent",
	"rental":       "rent",
	"buy":          "buy",
	"purchase":     "buy",
	"ads":          "ads",
}
