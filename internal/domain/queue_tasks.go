package domain

// DeleteURLsTask is containing short urls to delete and id of user who request deletion
type DeleteURLsTask struct {
	UserID    string
	ShortURLs []string
}
