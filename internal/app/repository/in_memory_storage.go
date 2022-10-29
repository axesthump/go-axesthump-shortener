package repository

import (
	"fmt"
	"strconv"
	"sync"
)

type InMemoryStorage struct {
	mx     *sync.RWMutex
	urls   map[int64]string
	lastID int64
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		mx:     &sync.RWMutex{},
		urls:   make(map[int64]string, 0),
		lastID: int64(0),
	}
}

func (s *InMemoryStorage) CreateShortURL(beginURL string, url string) (string, error) {

	s.mx.Lock()
	defer s.mx.Unlock()
	s.urls[s.lastID] = url
	shortEndpoint := strconv.FormatInt(s.lastID, 10)
	shortURL := beginURL + shortEndpoint
	s.lastID++
	return shortURL, nil
}

func (s *InMemoryStorage) GetFullURL(shortURL int64) (string, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	if longURL, ok := s.urls[shortURL]; ok {
		return longURL, nil
	} else {
		return "", fmt.Errorf("url dont exist")
	}
}

func (s *InMemoryStorage) GetAllURLs(beginURL string) []URLInfo {
	s.mx.RLock()
	defer s.mx.RUnlock()

	urls := make([]URLInfo, 0, len(s.urls))
	for shortURL, originalURL := range s.urls {
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
