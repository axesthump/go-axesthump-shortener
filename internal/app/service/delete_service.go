package service

import (
	"go-axesthump-shortener/internal/app/repository"
	"log"
	"strings"
	"time"
)

type DeleteService struct {
	urlsForDelete chan []repository.DeleteURL
	repo          repository.Repository
	baseURL       string
}

func NewDeleteService(repo repository.Repository, baseURL string) *DeleteService {
	ds := &DeleteService{
		urlsForDelete: make(chan []repository.DeleteURL),
		repo:          repo,
		baseURL:       baseURL,
	}
	for i := 0; i < 3; i++ {
		go func(ds *DeleteService) {
			for urlsForDelete := range ds.urlsForDelete {
				err := ds.repo.DeleteURLs(urlsForDelete)
				if err != nil {
					log.Printf("Found err %s", err)
					ds.reAddURLs(urlsForDelete)
				} else {
					log.Printf("Delete success!")
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
	time.Sleep(time.Millisecond * 100)
	go func() {
		ds.urlsForDelete <- urls
	}()
}

func (ds *DeleteService) Close() {
	close(ds.urlsForDelete)
}

func getURLsFromArr(data string, userID uint32, baseURL string) []repository.DeleteURL {
	data = data[1 : len(data)-1]
	splitData := strings.Split(data, ",")
	baseURL = baseURL + "/"
	urls := make([]repository.DeleteURL, len(splitData))
	for i, url := range splitData {
		url = url[1 : len(url)-1]
		url = strings.TrimSpace(url)
		url = strings.TrimPrefix(url, baseURL)
		urls[i] = repository.DeleteURL{URL: url, UserID: userID}
	}
	return urls
}
