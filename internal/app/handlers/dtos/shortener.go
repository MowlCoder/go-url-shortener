package dtos

type ShortURLDto struct {
	URL string `json:"url"`
}

type ShortURLResponse struct {
	Result string `json:"result"`
}

type ShortBatchURLDto struct {
	OriginalURL   string `json:"original_url"`
	CorrelationID string `json:"correlation_id"`
}

type ShortBatchURLResponse struct {
	ShortURL      string `json:"short_url"`
	CorrelationID string `json:"correlation_id"`
}
