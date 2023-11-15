package dtos

// ShortURLDto request body for url shorting
type ShortURLDto struct {
	URL string `json:"url"`
}

// ShortURLResponse response body of url shorting
type ShortURLResponse struct {
	Result string `json:"result"`
}

// ShortBatchURLDto request body for batch url shorting
type ShortBatchURLDto struct {
	OriginalURL   string `json:"original_url"`
	CorrelationID string `json:"correlation_id"`
}

// ShortBatchURLResponse response body of batch url shorting
type ShortBatchURLResponse struct {
	ShortURL      string `json:"short_url"`
	CorrelationID string `json:"correlation_id"`
}

// UserURLsResponse response body of getting user urls
type UserURLsResponse struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// DeleteURLsRequest request body for deleting urls
type DeleteURLsRequest []string
