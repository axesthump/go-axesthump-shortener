package service

import (
	"fmt"
	"sync"
)

type Storage struct {
	mx     *sync.RWMutex
	urls   map[int64]string
	lastId int64
}

func NewStorage() *Storage {
	return &Storage{
		mx:     &sync.RWMutex{},
		urls:   make(map[int64]string, 0),
		lastId: int64(0),
	}
}

func (s *Storage) CreateShortUrl(url string) (shortUrl int64) {
	s.mx.Lock()
	s.urls[s.lastId] = url
	s.mx.Unlock()

	shortUrl = s.lastId
	s.lastId++
	return shortUrl
}

func (s *Storage) GetFullUrl(shortUrl int64) (string, error) {
	s.mx.RLock()
	defer s.mx.RUnlock()
	if longUrl, ok := s.urls[shortUrl]; ok {
		return longUrl, nil
	} else {
		return "", fmt.Errorf("url dont exist")
	}
}
