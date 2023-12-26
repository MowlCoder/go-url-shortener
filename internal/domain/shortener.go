package domain

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
