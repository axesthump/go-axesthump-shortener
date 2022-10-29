package repository

type URLInfo struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Repository interface {
	CreateShortURL(beginURL string, url string) (string, error)
	GetFullURL(shortURL int64) (string, error)
	GetAllURLs(beginURL string) []URLInfo
	Close() error
}
