package newsapi

// Article is a single item from the top-headlines or everything responses.
// Field set follows https://newsapi.org/docs/ — unknown fields are ignored.
type Article struct {
	Source   *NameRef `json:"source,omitempty"`
	Author   string   `json:"author,omitempty"`
	Title    string   `json:"title,omitempty"`
	Desc     string   `json:"description,omitempty"`
	URL      string   `json:"url,omitempty"`
	URLToImg string   `json:"urlToImage,omitempty"`
	PubAt    string   `json:"publishedAt,omitempty"`
	Content  string   `json:"content,omitempty"`
}

// NameRef is a minimal { id, name } source (used inside Article).
type NameRef struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// Source is an entry in the v2 (or v1) sources list.
type Source struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	URL         string `json:"url,omitempty"`
	Category    string `json:"category,omitempty"`
	Language    string `json:"language,omitempty"`
	Country     string `json:"country,omitempty"`
}

// TopHeadlinesResult is the v2 /v2/top-headlines response body.
type TopHeadlinesResult struct {
	Status       string    `json:"status"`
	TotalResults int       `json:"totalResults,omitempty"`
	Articles     []Article `json:"articles"`
}

// EverythingResult is the v2 /v2/everything response body.
type EverythingResult struct {
	Status       string    `json:"status"`
	TotalResults int       `json:"totalResults,omitempty"`
	Articles     []Article `json:"articles"`
}

// SourcesV2Result is the v2 /v2/sources response body.
type SourcesV2Result struct {
	Status  string   `json:"status"`
	Sources []Source `json:"sources"`
}

// ArticlesV1Result is the v1 /v1/articles response body.
type ArticlesV1Result struct {
	Status   string    `json:"status"`
	Source   string    `json:"source"`
	SortBy   string    `json:"sortBy"`
	Articles []Article `json:"articles"`
}

// SourcesV1Result is the v1 /v1/sources response body.
type SourcesV1Result struct {
	Status  string   `json:"status"`
	Sources []Source `json:"sources"`
}
