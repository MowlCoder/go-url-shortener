package domain

// ShortenedURL is model of shortened url. Use model to store data in storages.
type ShortenedURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id"`
	ID          int    `json:"id"`
	IsDeleted   bool   `json:"is_deleted"`
}

// SaveShortURLDto contains info about short url saving to pass around layers.
type SaveShortURLDto struct {
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
	UserID      string `json:"user_id"`
}

// InternalStats contains internal stats about system state
type InternalStats struct {
	URLs  int `json:"urls"`
	Users int `json:"users"`
}

type ShortBatchURL struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
	ShortURL      string `json:"short_url"`
}
