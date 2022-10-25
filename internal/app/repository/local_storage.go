package repository

import (
	"bufio"
	"errors"
	"os"
	"strconv"
	"strings"
	"sync"
)

type LocalStorage struct {
	mx     *sync.RWMutex
	file   *os.File
	lastID int64
}

func NewLocalStorage(filename string) (*LocalStorage, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return nil, err
	}
	lastID := getLastId(file)
	return &LocalStorage{
		mx:     &sync.RWMutex{},
		file:   file,
		lastID: lastID,
	}, nil
}

func getLastId(file *os.File) int64 {
	var lastID int64
	var err error
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		data := string(scanner.Bytes())
		urlData := strings.Split(data, "~")
		if len(urlData) != 2 {
			panic(errors.New("bad data in file"))
		}
		lastID, err = strconv.ParseInt(urlData[0], 10, 64)
		if err != nil {
			panic(errors.New("bad data in file"))
		}
	}
	return lastID + 1
}

func (ls *LocalStorage) CreateShortURL(beginURL string, url string) (string, error) {
	ls.mx.Lock()
	defer ls.mx.Unlock()
	shortEndpoint := strconv.FormatInt(ls.lastID, 10)
	shortURL := beginURL + shortEndpoint
	data := strconv.FormatInt(ls.lastID, 10) + "~" + url

	wr := bufio.NewWriter(ls.file)
	_, err := wr.Write([]byte(data))
	if err != nil {
		return "", err
	}

	err = wr.WriteByte('\n')
	if err != nil {
		return "", err
	}
	err = wr.Flush()
	if err != nil {
		return "", err
	}

	ls.lastID++
	return shortURL, nil
}

func (ls *LocalStorage) GetFullURL(shortURL int64) (string, error) {
	ls.mx.RLock()
	defer ls.mx.RUnlock()
	fileForRead, err := os.OpenFile(ls.file.Name(), os.O_RDONLY, 0777)
	if err != nil {
		return "", err
	}
	defer fileForRead.Close()
	scanner := bufio.NewScanner(fileForRead)
	for scanner.Scan() {
		data := string(scanner.Bytes())
		urlData := strings.Split(data, "~")
		if len(urlData) != 2 {
			panic(errors.New("bad data in file"))
		}
		elementShortURL, err := strconv.ParseInt(urlData[0], 10, 64)
		if err != nil {
			return "", err
		}
		if elementShortURL == shortURL {
			return urlData[1], nil
		}
	}
	return "", errors.New("url nor found")
}

func (ls *LocalStorage) Close() error {
	return ls.file.Close()
}
