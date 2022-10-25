package repository

type Repository interface {
	CreateShortURL(beginURL string, url string) (string, error)
	GetFullURL(shortURL int64) (string, error)
	Close() error
}
