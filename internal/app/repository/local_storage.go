package repository

import (
	"encoding/gob"
	"errors"
	"io"
	"os"
	"strconv"
	"sync"
)

type urlData struct {
	shortURL int64
	longURL  string
}

type LocalStorage struct {
	mx      *sync.RWMutex
	file    *os.File
	encoder *gob.Encoder
	lastID  int64
}

func NewLocalStorage(filename string) (*LocalStorage, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return nil, err
	}
	return &LocalStorage{
		mx:      &sync.RWMutex{},
		file:    file,
		encoder: gob.NewEncoder(file),
		lastID:  int64(0),
	}, nil
}

func (ls *LocalStorage) CreateShortURL(beginURL string, url string) (string, error) {
	ls.mx.Lock()
	defer ls.mx.Unlock()
	shortEndpoint := strconv.FormatInt(ls.lastID, 10)
	shortURL := beginURL + shortEndpoint

	urlData := urlData{
		ls.lastID,
		url,
	}

	err := ls.encoder.Encode(&urlData)
	if err != nil {
		return "", err
	}

	ls.lastID++
	return shortURL, nil
}

func (ls *LocalStorage) GetFullURL(shortURL int64) (string, error) {
	decoder := gob.NewDecoder(ls.file)
	for {
		var urlData urlData
		err := decoder.Decode(&urlData)
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		if urlData.shortURL == shortURL {
			return urlData.longURL, nil
		}
	}
	return "", errors.New("url nor found")
}

func (ls *LocalStorage) Close() error {
	return ls.file.Close()
}
