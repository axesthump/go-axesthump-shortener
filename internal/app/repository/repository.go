package repository

type URLInfo struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Repository interface {
	CreateShortURL(beginURL string, url string, userID uint32) (string, error)
	GetFullURL(shortURL int64, userID uint32) (string, error)
	GetAllURLs(beginURL string, userID uint32) []URLInfo
	Close() error
}
