package repository

import (
	"context"
)

type URLInfo struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type Repository interface {
	CreateShortURL(ctx context.Context, beginURL string, originalURL string, userID uint32) (string, error)
	GetFullURL(ctx context.Context, shortURL int64) (string, error)
	GetAllURLs(ctx context.Context, beginURL string, userID uint32) []URLInfo
	Close() error
}
