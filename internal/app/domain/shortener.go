package domain

type SaveShortURLDto struct {
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
}
