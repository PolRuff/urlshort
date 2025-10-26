package model

// ShortenRequest represents the JSON request body for /api/shorten
type ShortenRequest struct {
	URL string `json:"url"`
}

// ShortenResponse represents the JSON response body for /api/shorten
type ShortenResponse struct {
	Result string `json:"result"`
}

// URLPair represents a mapping between a short ID and an original URL
type URLPair struct {
	ShortID string
	URL     string
}
