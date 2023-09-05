package domain

type SaveShortUrlDto struct {
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
}
