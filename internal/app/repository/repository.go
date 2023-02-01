package repository

import (
	"context"
)

// DeletedURLError delete url error.
type DeletedURLError struct {
}

// Error return DeletedURLError description.
func (e *DeletedURLError) Error() string {
	return "URL deleted"
}

// DeleteURL contains info about url for delete.
type DeleteURL struct {
	// URL - url for delete.
	URL string
	// UserID - id user, who want delete this url.
	UserID uint32
}

// URLInfo contains url info.
type URLInfo struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// URLWithID contains url (short/original) and correlation id.
type URLWithID struct {
	CorrelationID string
	URL           string
}

// Repository define api for work with storage.
type Repository interface {
	// CreateShortURL creates short url. Returns short url if operations success or error.
	CreateShortURL(ctx context.Context, beginURL string, originalURL string, userID uint32) (string, error)

	// CreateShortURLs creates short urls. Returns slice short urls if operations success or error.
	CreateShortURLs(ctx context.Context, beginURL string, urls []URLWithID, userID uint32) ([]URLWithID, error)

	// GetFullURL returns full url by short url.
	GetFullURL(ctx context.Context, shortURL int64) (string, error)

	// GetAllURLs returns all urls owned specific user.
	GetAllURLs(ctx context.Context, beginURL string, userID uint32) []URLInfo

	// DeleteURLs delete url from urlsForDelete.
	DeleteURLs(urlsForDelete []DeleteURL) error

	// GetStats returns count of users and shortURLs
	GetStats() (map[string]int, error)

	// Close closes everything that should be closed in the context of the repository.
	Close() error
}
