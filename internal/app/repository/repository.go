package repository

import (
	"context"
)

type URLInfo struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type DeleteURL struct {
	URL    string
	UserID uint32
}

type URLWithID struct {
	CorrelationID string
	URL           string
}

type Repository interface {
	CreateShortURL(ctx context.Context, beginURL string, originalURL string, userID uint32) (string, error)
	CreateShortURLs(ctx context.Context, beginURL string, urls []URLWithID, userID uint32) ([]URLWithID, error)
	GetFullURL(ctx context.Context, shortURL int64) (string, error)
	GetAllURLs(ctx context.Context, beginURL string, userID uint32) []URLInfo
	DeleteURLs(urlsForDelete []DeleteURL) error
	Close() error
}
