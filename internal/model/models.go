package model

// ShortenRequest represents the JSON request body for /api/shorten
type ShortenRequest struct {
	URL string `json:"url"`
}

// ShortenResponse represents the JSON response body for /api/shorten
type ShortenResponse struct {
	Result string `json:"result"`
}

// BatchShortenRequestItem represents a single item in the batch request
type BatchShortenRequestItem struct {
	CorrelationID string `json:"correlation_id"` // Идентификатор для сопоставления результата
	OriginalURL   string `json:"original_url"`   // URL для сокращения
}

// BatchShortenResponseItem represents a single item in the batch response
type BatchShortenResponseItem struct {
	CorrelationID string `json:"correlation_id"` // Соответствует запросу
	ShortURL      string `json:"short_url"`      // Результат сокращения
}

type UserUrls struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// URLRecord represents a single record in the repository
type URLRecord struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      uint32
}
