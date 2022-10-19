package service

import (
	"fmt"
	"sync"
)

type Storage struct {
	mx     *sync.RWMutex
	urls   map[int64]string
	lastID int64
}

func NewStorage() *Storage {
	return &Storage{
		mx:     &sync.RWMutex{},
		urls:   make(map[int64]string, 0),
		lastID: int64(0),
	}
}

func (s *Storage) CreateShortURL(url string) (shortURL int64) {
	s.mx.Lock()
	s.urls[s.lastID] = url
	s.mx.Unlock()

	shortURL = s.lastID
	s.lastID++
	return shortURL
}

func (s *Storage) GetFullURL(shortUrl int64) (string, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	if longURL, ok := s.urls[shortUrl]; ok {
		return longURL, nil
	} else {
		return "", fmt.Errorf("url dont exist")
	}
}
