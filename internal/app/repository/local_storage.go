package repository

import (
	"bufio"
	"context"
	"errors"
	"go-axesthump-shortener/internal/app/generator"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
)

// Info about store data in file.
const (
	splitSeq       = "~s~e~c~" // separator for one row with data
	countDataInRow = 4         // count data from url in one row
)

// url delete url info.
type url struct {
	url       string
	fullURL   string
	userID    uint32
	isDeleted bool
}

// LocalStorage contains data for local storage.
type LocalStorage struct {
	sync.RWMutex
	file        *os.File
	idGenerator *generator.IDGenerator
}

// NewLocalStorage returns new LocalStorage.
func NewLocalStorage(filename string) (*LocalStorage, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return nil, err
	}
	lastID := getLastID(file)
	return &LocalStorage{
		RWMutex:     sync.RWMutex{},
		file:        file,
		idGenerator: generator.NewIDGenerator(lastID),
	}, nil
}

// GetUserLastID returns last user id contains in local storage.
func (ls *LocalStorage) GetUserLastID() uint32 {
	var lastID int64
	var err error
	var max int64 = 0
	scanner := bufio.NewScanner(ls.file)
	for scanner.Scan() {
		data := scanner.Text()
		urlData := strings.Split(data, splitSeq)
		if len(urlData) != countDataInRow {
			panic(errors.New("bad data in file"))
		}
		if lastID, err = strconv.ParseInt(urlData[0], 10, 64); err != nil {
			panic(errors.New("bad data in file"))
		}
		if lastID > max {
			max = lastID
		}
	}
	return uint32(max + 1)
}

// CreateShortURL creates short url. Returns short url if operations success or error.
func (ls *LocalStorage) CreateShortURL(
	ctx context.Context,
	beginURL string,
	originalURL string,
	userID uint32,
) (string, error) {
	newShortID := ls.idGenerator.GetID()
	shortEndpoint := strconv.FormatInt(newShortID, 10)
	shortURL := beginURL + shortEndpoint

	data := createRow(int64(userID), strconv.FormatInt(newShortID, 10), originalURL, "false")

	ls.Lock()
	wr := bufio.NewWriter(ls.file)
	if _, err := wr.WriteString(data + "\n"); err != nil {
		return "", err
	}
	if err := wr.Flush(); err != nil {
		return "", err
	}
	ls.Unlock()
	return shortURL, nil
}

// CreateShortURLs creates short urls. Returns slice short urls if operations success or error.
func (ls *LocalStorage) CreateShortURLs(
	ctx context.Context,
	beginURL string,
	urls []URLWithID,
	userID uint32,
) ([]URLWithID, error) {
	res := make([]URLWithID, len(urls))
	for i, url := range urls {
		shortURL, err := ls.CreateShortURL(ctx, beginURL, url.URL, userID)
		if err != nil {
			return nil, err
		}
		res[i].URL = shortURL
		res[i].CorrelationID = url.CorrelationID
	}
	return res, nil
}

// GetFullURL returns full url by short url.
func (ls *LocalStorage) GetFullURL(ctx context.Context, shortURL int64) (string, error) {
	ls.RLock()
	defer ls.RUnlock()
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
		if len(urlData) != countDataInRow {
			panic(errors.New("bad data in file"))
		}
		var elementShortURL int64
		if elementShortURL, err = strconv.ParseInt(urlData[1], 10, 64); err != nil {
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

// DeleteURLs deletes url from urlsForDelete.
func (ls *LocalStorage) DeleteURLs(urlsForDelete []DeleteURL) error {
	ls.Lock()
	defer ls.Unlock()

	fileForRead, err := os.OpenFile(ls.file.Name(), os.O_RDONLY, 0777)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(fileForRead)
	urlsForDeleteData := make([]url, 0, len(urlsForDelete))
	for scanner.Scan() {
		data := scanner.Text()
		urlData := strings.Split(data, splitSeq)
		if len(urlData) != countDataInRow {
			continue
		}
		storageUserID, err := strconv.ParseInt(urlData[0], 10, 64)
		if err != nil || uint32(storageUserID) != urlsForDelete[0].UserID {
			continue
		}
		shortURL := urlData[1]
		urlForDelete := url{
			userID:    urlsForDelete[0].UserID,
			url:       shortURL,
			fullURL:   urlData[2],
			isDeleted: true,
		}
		urlsForDeleteData = append(urlsForDeleteData, urlForDelete)
	}

	fileForRead.Close()
	wr := bufio.NewWriter(ls.file)
	for _, url := range urlsForDeleteData {
		data := createRow(int64(url.userID), url.url, url.fullURL, "true")
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

// GetAllURLs returns all urls owned specific user.
func (ls *LocalStorage) GetAllURLs(ctx context.Context, beginURL string, userID uint32) []URLInfo {
	ls.RLock()
	defer ls.RUnlock()
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
		if len(urlData) != countDataInRow {
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

// Close closes everything that should be closed in the context of the repository.
func (ls *LocalStorage) Close() error {
	return ls.file.Close()
}

// getLastID returns last short url in local storage.
func getLastID(file *os.File) int64 {
	var lastID int64
	var err error
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		data := scanner.Text()
		urlData := strings.Split(data, splitSeq)
		if len(urlData) != countDataInRow {
			panic(errors.New("bad data in file"))
		}
		if lastID, err = strconv.ParseInt(urlData[1], 10, 64); err != nil {
			panic(errors.New("bad data in file"))
		}
	}
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return 0
	}
	return lastID + 1
}

// createRow returns new row to append in local storage.
func createRow(userID int64, shortURL string, fullURL string, isDeleted string) string {
	return strconv.FormatInt(userID, 10) +
		splitSeq + shortURL +
		splitSeq + fullURL +
		splitSeq + isDeleted
}
