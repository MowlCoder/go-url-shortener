package models

type ShortenedURL struct {
	ID          int    `json:"id"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id"`
	IsDeleted   bool   `json:"is_deleted"`
}
