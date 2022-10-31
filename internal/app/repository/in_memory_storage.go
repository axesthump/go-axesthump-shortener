package repository

import (
	"fmt"
	"strconv"
	"sync"
)

type urlInfo struct {
	urls map[int64]string
}

type InMemoryStorage struct {
	mx       *sync.RWMutex
	userURLs map[uint32]urlInfo
	lastID   int64
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		mx:       &sync.RWMutex{},
		userURLs: make(map[uint32]urlInfo),
		lastID:   int64(0),
	}
}

func (s *InMemoryStorage) CreateShortURL(beginURL string, url string, userID uint32) (string, error) {
	s.mx.Lock()
	defer s.mx.Unlock()
	if _, ok := s.userURLs[userID]; !ok {
		s.userURLs[userID] = urlInfo{
			make(map[int64]string),
		}
	}
	s.userURLs[userID].urls[s.lastID] = url
	shortEndpoint := strconv.FormatInt(s.lastID, 10)
	shortURL := beginURL + shortEndpoint
	s.lastID++
	return shortURL, nil
}

func (s *InMemoryStorage) GetFullURL(shortURL int64, userID uint32) (string, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	if longURL, ok := s.userURLs[userID].urls[shortURL]; ok {
		return longURL, nil
	} else {
		return "", fmt.Errorf("url dont exist")
	}
}

func (s *InMemoryStorage) GetAllURLs(beginURL string, userID uint32) []URLInfo {
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
