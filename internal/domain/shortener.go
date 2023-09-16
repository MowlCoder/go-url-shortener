package domain

type SaveShortURLDto struct {
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
	UserID      string `json:"user_id"`
}

type DeleteURLsTask struct {
	ShortURLs []string
	UserID    string
}
