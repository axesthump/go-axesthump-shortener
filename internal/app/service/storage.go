package service

import (
	"fmt"
	"strings"
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

func (s *Storage) CreateShortURL(url string) (int64, error) {
	urlElements := strings.Split(url, "/")
	if len(urlElements) < 4 {
		return 0, fmt.Errorf("bad url")
	}

	endpoint := ""
	beginURL := ""

	for i, el := range urlElements {
		if i >= 3 {
			endpoint += el + "/"
		} else {
			beginURL += el + "/"
		}
	}

	endpoint = strings.TrimSuffix(endpoint, "/")
	s.mx.Lock()
	s.urls[s.lastID] = beginURL + endpoint
	s.mx.Unlock()

	shortURL := s.lastID
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
