package repository

import (
	"context"
	"fmt"
	"strconv"
	"sync"
)

type StorageURL struct {
	urls map[int64]string
}

type InMemoryStorage struct {
	mx       *sync.RWMutex
	userURLs map[uint32]StorageURL
	lastID   int64
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		mx:       &sync.RWMutex{},
		userURLs: make(map[uint32]StorageURL),
		lastID:   int64(0),
	}
}

func (s *InMemoryStorage) CreateShortURL(
	ctx context.Context,
	beginURL string,
	originalURL string,
	userID uint32,
) (string, error) {
	s.mx.Lock()
	defer s.mx.Unlock()
	if _, ok := s.userURLs[userID]; !ok {
		s.userURLs[userID] = StorageURL{
			make(map[int64]string),
		}
	}
	s.userURLs[userID].urls[s.lastID] = originalURL
	shortEndpoint := strconv.FormatInt(s.lastID, 10)
	shortURL := beginURL + shortEndpoint
	s.lastID++
	return shortURL, nil
}

func (s *InMemoryStorage) GetFullURL(ctx context.Context, shortURL int64) (string, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	for _, storageURL := range s.userURLs {
		if longURL, ok := storageURL.urls[shortURL]; ok {
			return longURL, nil
		}
	}
	return "", fmt.Errorf("url dont exist")
}

func (s *InMemoryStorage) GetAllURLs(ctx context.Context, beginURL string, userID uint32) []URLInfo {
	s.mx.RLock()
	defer s.mx.RUnlock()

	urls := make([]URLInfo, 0, len(s.userURLs))
	for shortURL, originalURL := range s.userURLs[userID].urls {
		short := strconv.FormatInt(shortURL, 10)
		url := URLInfo{
			ShortURL:    beginURL + short,
			OriginalURL: originalURL,
		}
		urls = append(urls, url)
	}
	return urls
}

func (s *InMemoryStorage) Close() error {
	return nil
}
