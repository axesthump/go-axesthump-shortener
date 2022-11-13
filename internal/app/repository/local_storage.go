package repository

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
)

const splitSeq = "~s~e~c~"

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
	lastID := getLastID(file)
	return &LocalStorage{
		mx:     &sync.RWMutex{},
		file:   file,
		lastID: lastID,
	}, nil
}

func getLastID(file *os.File) int64 {
	var lastID int64
	var err error
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		data := string(scanner.Bytes())
		urlData := strings.Split(data, splitSeq)
		if len(urlData) != 4 {
			panic(errors.New("bad data in file"))
		}
		lastID, err = strconv.ParseInt(urlData[1], 10, 64)
		if err != nil {
			panic(errors.New("bad data in file"))
		}
	}
	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return 0
	}
	return lastID + 1
}

func (ls *LocalStorage) GetUserLastID() uint32 {
	var lastID int64
	var err error
	var max int64 = 0
	scanner := bufio.NewScanner(ls.file)
	for scanner.Scan() {
		data := string(scanner.Bytes())
		urlData := strings.Split(data, splitSeq)
		if len(urlData) != 4 {
			panic(errors.New("bad data in file"))
		}
		lastID, err = strconv.ParseInt(urlData[0], 10, 64)
		if lastID > max {
			max = lastID
		}
		if err != nil {
			panic(errors.New("bad data in file"))
		}
	}
	return uint32(max + 1)
}

func (ls *LocalStorage) CreateShortURL(
	ctx context.Context,
	beginURL string,
	originalURL string,
	userID uint32,
) (string, error) {
	ls.mx.Lock()
	defer ls.mx.Unlock()
	shortEndpoint := strconv.FormatInt(ls.lastID, 10)
	shortURL := beginURL + shortEndpoint
	data := strconv.FormatInt(int64(userID), 10) +
		splitSeq + strconv.FormatInt(ls.lastID, 10) +
		splitSeq + originalURL +
		splitSeq + "false"

	wr := bufio.NewWriter(ls.file)
	if _, err := wr.WriteString(data + "\n"); err != nil {
		return "", err
	}
	if err := wr.Flush(); err != nil {
		return "", err
	}
	ls.lastID++
	return shortURL, nil
}

func (ls *LocalStorage) CreateShortURLs(
	ctx context.Context,
	beginURL string,
	urls []URLWithID,
	userID uint32,
) ([]URLWithID, error) {
	res := make([]URLWithID, 0, len(urls))
	for _, url := range urls {
		shortURL, err := ls.CreateShortURL(ctx, beginURL, url.URL, userID)
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

func (ls *LocalStorage) GetFullURL(ctx context.Context, shortURL int64) (string, error) {
	ls.mx.RLock()
	defer ls.mx.RUnlock()
	fileForRead, err := os.OpenFile(ls.file.Name(), os.O_RDONLY, 0777)
	if err != nil {
		return "", err
	}
	defer fileForRead.Close()
	scanner := bufio.NewScanner(fileForRead)
	var fullURL string
	var finalErr error
	for scanner.Scan() {
		data := scanner.Text()
		urlData := strings.Split(data, splitSeq)
		if len(urlData) != 4 {
			panic(errors.New("bad data in file"))
		}
		var elementShortURL int64
		elementShortURL, err = strconv.ParseInt(urlData[1], 10, 64)
		if err != nil {
			return "", err
		}
		if elementShortURL == shortURL {
			if urlData[3] == "true" {
				finalErr = &DeletedURLError{}
			} else {
				finalErr = nil
				fullURL = urlData[2]
			}
		}
	}
	if len(fullURL) == 0 {
		return "", errors.New("URL nor found")
	} else {
		return fullURL, finalErr
	}
}

func (ls *LocalStorage) DeleteURLs(urlsForDelete []DeleteURL) error {
	ls.mx.Lock()
	defer ls.mx.Unlock()

	fileForRead, err := os.OpenFile(ls.file.Name(), os.O_RDONLY, 0777)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(fileForRead)
	urlsForDeleteData := make([]url, 0, len(urlsForDelete))
	for scanner.Scan() {
		data := scanner.Text()
		urlData := strings.Split(data, splitSeq)
		if len(urlData) != 4 {
			continue
		}
		storageUserID, err := strconv.ParseInt(urlData[0], 10, 64)
		if err != nil || uint32(storageUserID) != urlsForDelete[0].UserID {
			continue
		}
		shortURL := urlData[1]
		url := url{
			userID:    urlsForDelete[0].UserID,
			url:       shortURL,
			fullURL:   urlData[2],
			isDeleted: true,
		}
		urlsForDeleteData = append(urlsForDeleteData, url)
	}

	fileForRead.Close()
	wr := bufio.NewWriter(ls.file)
	for _, url := range urlsForDeleteData {
		data := strconv.FormatInt(int64(url.userID), 10) +
			splitSeq + url.url +
			splitSeq + url.fullURL +
			splitSeq + "true"
		if _, err := wr.WriteString(data + "\n"); err != nil {
			return err
		}
		if err := wr.Flush(); err != nil {
			return err
		}
	}
	if err := wr.Flush(); err != nil {
		return err
	}
	return nil
}

func (ls *LocalStorage) GetAllURLs(ctx context.Context, beginURL string, userID uint32) []URLInfo {
	ls.mx.RLock()
	defer ls.mx.RUnlock()
	fileForRead, err := os.OpenFile(ls.file.Name(), os.O_RDONLY, 0777)
	urls := make([]URLInfo, 0)
	if err != nil {
		return urls
	}
	defer fileForRead.Close()
	scanner := bufio.NewScanner(fileForRead)
	for scanner.Scan() {
		data := scanner.Text()
		urlData := strings.Split(data, splitSeq)
		if len(urlData) != 4 {
			continue
		}
		storageUserID, err := strconv.ParseInt(urlData[0], 10, 64)
		if err != nil || uint32(storageUserID) != userID {
			continue
		}
		shortURL, err := strconv.ParseInt(urlData[1], 10, 64)
		if err != nil {
			continue
		}
		shortEndpoint := strconv.FormatInt(shortURL, 10)
		short := beginURL + shortEndpoint
		url := URLInfo{
			ShortURL:    short,
			OriginalURL: urlData[1],
		}
		urls = append(urls, url)
	}
	return urls
}

func (ls *LocalStorage) Close() error {
	return ls.file.Close()
}
