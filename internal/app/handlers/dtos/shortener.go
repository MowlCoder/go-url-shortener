package dtos

type ShortURLDto struct {
	URL string `json:"url"`
}

type ShortURLResponse struct {
	Result string `json:"result"`
}
