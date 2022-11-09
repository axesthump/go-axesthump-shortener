package repository

import (
	"context"
	"fmt"
	"strconv"
	"sync"
)

type StorageURL struct {
	url       string
	userID    uint32
	isDeleted bool
}

type InMemoryStorage struct {
	sync.RWMutex
	userURLs map[int64]StorageURL
	lastID   int64
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		userURLs: make(map[int64]StorageURL),
		lastID:   int64(0),
	}
}

func (s *InMemoryStorage) CreateShortURL(
	ctx context.Context,
	beginURL string,
	originalURL string,
	userID uint32,
) (string, error) {
	s.Lock()
	defer s.Unlock()
	shortEndpoint := strconv.FormatInt(s.lastID, 10)
	shortURL := beginURL + shortEndpoint
	s.userURLs[s.lastID] = StorageURL{
		url:    originalURL,
		userID: userID,
	}
	s.lastID++
	return shortURL, nil
}

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

func (s *InMemoryStorage) DeleteURLs(urlsForDelete []DeleteURL) error {
	s.Lock()
	for _, url := range urlsForDelete {
		shortURL, err := strconv.ParseInt(url.URL, 10, 64)
		if err != nil {
			return err
		}
		savedURL, ok := s.userURLs[shortURL]
		if !ok {
			continue
		}
		if savedURL.userID == url.UserID {
			savedURL.isDeleted = true
		}
	}
	s.Unlock()
	return nil
}

func (s *InMemoryStorage) Close() error {
	return nil
}
