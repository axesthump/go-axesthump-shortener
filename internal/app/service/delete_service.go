package service

import (
	"go-axesthump-shortener/internal/app/repository"
	"strings"
)

type DeleteService struct {
	urlsForDelete chan []repository.DeleteURL
	repository    repository.Repository
	baseURL       string
}

func NewDeleteService(r repository.Repository, baseURL string) *DeleteService {
	ds := &DeleteService{
		urlsForDelete: make(chan []repository.DeleteURL),
		repository:    r,
		baseURL:       baseURL,
	}
	for i := 0; i < 5; i++ {
		go func(ds *DeleteService) {
			for {
				data, ok := <-ds.urlsForDelete
				if !ok {
					return
				}
				err := r.DeleteURLs(data)
				if err != nil {
					ds.reAddURLs(data)
				}
			}

		}(ds)
	}
	return ds
}

func (ds *DeleteService) AddURLs(data string, userID uint32) {
	go func() {
		ds.urlsForDelete <- getURLsFromArr(data, userID, ds.baseURL)
	}()
}

func (ds *DeleteService) reAddURLs(urls []repository.DeleteURL) {
	go func() {
		ds.urlsForDelete <- urls
	}()
}

func (ds *DeleteService) Close() {
	close(ds.urlsForDelete)
}

func getURLsFromArr(data string, userID uint32, baseURL string) []repository.DeleteURL {
	data = data[1 : len(data)-1]
	data = strings.ReplaceAll(data, "\"", "")
	splitData := strings.Split(data, ",")
	urls := make([]repository.DeleteURL, len(splitData))
	for i, url := range splitData {
		url = strings.TrimPrefix(url, baseURL+"/")
		urls[i] = repository.DeleteURL{URL: url, UserID: userID}
	}
	return urls
}
