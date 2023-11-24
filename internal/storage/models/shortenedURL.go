package models

// ShortenedURL is model of shortened url. Use model to store data in storages.
type ShortenedURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id"`
	ID          int    `json:"id"`
	IsDeleted   bool   `json:"is_deleted"`
}
