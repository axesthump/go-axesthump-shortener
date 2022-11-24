package repository

import (
	"context"
	"fmt"
	"go-axesthump-shortener/internal/app/generator"
	"strconv"
	"sync"
)

// StorageURL url info.
type StorageURL struct {
	url       string
	userID    uint32
	isDeleted bool
}

// InMemoryStorage contains data for in memory storage.
type InMemoryStorage struct {
	sync.RWMutex
	userURLs    map[int64]*StorageURL
	idGenerator *generator.IDGenerator
}

// NewInMemoryStorage returns new InMemoryStorage.
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		userURLs:    make(map[int64]*StorageURL),
		idGenerator: generator.NewIDGenerator(0),
	}
}

// CreateShortURL create short url. Returns short url if operations success or error.
func (s *InMemoryStorage) CreateShortURL(
	ctx context.Context,
	beginURL string,
	originalURL string,
	userID uint32,
) (string, error) {
	newShortURL := s.idGenerator.GetID()
	shortEndpoint := strconv.FormatInt(newShortURL, 10)
	shortURL := beginURL + shortEndpoint
	s.Lock()
	s.userURLs[newShortURL] = &StorageURL{
		url:    originalURL,
		userID: userID,
	}
	s.Unlock()
	return shortURL, nil
}

// GetFullURL returns full url by short url.
func (s *InMemoryStorage) GetFullURL(ctx context.Context, shortURL int64) (string, error) {
	s.RLock()
	defer s.RUnlock()
	if url, ok := s.userURLs[shortURL]; ok {
		if url.isDeleted {
			return "", &DeletedURLError{}
		}
		return url.url, nil
	}
	return "", fmt.Errorf("URL dont exist")
}

// GetAllURLs returns all urls owned specific user.
func (s *InMemoryStorage) GetAllURLs(ctx context.Context, beginURL string, userID uint32) []URLInfo {
	s.RLock()
	defer s.RUnlock()

	urls := make([]URLInfo, 0, len(s.userURLs))
	for shortURL, urlInfo := range s.userURLs {
		if urlInfo.userID != userID {
			continue
		}
		short := strconv.FormatInt(shortURL, 10)
		url := URLInfo{
			ShortURL:    beginURL + short,
			OriginalURL: urlInfo.url,
		}
		urls = append(urls, url)
	}
	return urls
}

// CreateShortURLs create short urls. Returns slice short urls if operations success or error.
func (s *InMemoryStorage) CreateShortURLs(
	ctx context.Context,
	beginURL string,
	urls []URLWithID,
	userID uint32,
) ([]URLWithID, error) {
	res := make([]URLWithID, 0, len(urls))
	for _, url := range urls {
		shortURL, err := s.CreateShortURL(ctx, beginURL, url.URL, userID)
		if err != nil {
			return nil, err
		}
		res = append(res, URLWithID{
			CorrelationID: url.CorrelationID,
			URL:           shortURL,
		})
	}
	return res, nil
}

// DeleteURLs delete url from urlsForDelete.
func (s *InMemoryStorage) DeleteURLs(urlsForDelete []DeleteURL) error {
	s.Lock()
	for _, urlForDelete := range urlsForDelete {
		shortURL, err := strconv.ParseInt(urlForDelete.URL, 10, 64)
		if err != nil {
			return err
		}
		if savedURL, ok := s.userURLs[shortURL]; ok {
			if savedURL.userID == urlForDelete.UserID {
				savedURL.isDeleted = true
			}
		}
	}
	s.Unlock()
	return nil
}

// Close closes everything that should be closed in the context of the repository.
func (s *InMemoryStorage) Close() error {
	s.idGenerator.Cancel()
	return nil
}
