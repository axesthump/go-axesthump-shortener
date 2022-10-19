package service

import (
	"fmt"
	"strconv"
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

func (s *Storage) CreateShortURL(beginURL string, url string) (string, error) {

	s.mx.Lock()
	s.urls[s.lastID] = url
	s.mx.Unlock()

	shortEndpoint := strconv.FormatInt(s.lastID, 10)
	shortURL := beginURL + shortEndpoint
	s.lastID++
	return shortURL, nil
}

func (s *Storage) GetFullURL(shortURL int64) (string, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	if longURL, ok := s.urls[shortURL]; ok {
		return longURL, nil
	} else {
		return "", fmt.Errorf("url dont exist")
	}
}
