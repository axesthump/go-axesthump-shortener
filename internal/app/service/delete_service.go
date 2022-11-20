// Package service define service for delete.
package service

import (
	"go-axesthump-shortener/internal/app/repository"
	"log"
	"strings"
	"time"
)

// DeleteService contains data for delete service.
type DeleteService struct {
	// urlsForDelete - take urls for delete.
	urlsForDelete chan []repository.DeleteURL
	repo          repository.Repository
	baseURL       string
}

// NewDeleteService return new DeleteService and start deleteService logic.
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

// AddURLs add new urls for delete in chan.
func (ds *DeleteService) AddURLs(data string, userID uint32) {
	go func() {
		ds.urlsForDelete <- getURLsFromArr(data, userID, ds.baseURL)
	}()
}

func (ds *DeleteService) Close() {
	close(ds.urlsForDelete)
}

// reAddURLs if db connection is unstable urls for delete add in chan again.
func (ds *DeleteService) reAddURLs(urls []repository.DeleteURL) {
	time.Sleep(time.Millisecond * 100)
	go func() {
		ds.urlsForDelete <- urls
	}()
}

// getURLsFromArr convert data to slice DeleteURL.
func getURLsFromArr(data string, userID uint32, baseURL string) []repository.DeleteURL {
	data = data[1 : len(data)-1]
	splitData := strings.Split(data, ",")
	baseURL = baseURL + "/"
	urls := make([]repository.DeleteURL, len(splitData))
	for i, url := range splitData {
		url = strings.TrimSpace(url)
		url = url[1 : len(url)-1]
		url = strings.TrimPrefix(url, baseURL)
		urls[i] = repository.DeleteURL{URL: url, UserID: userID}
	}
	return urls
}
