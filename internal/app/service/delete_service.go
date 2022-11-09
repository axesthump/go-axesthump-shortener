package service

import (
	"bytes"
	"context"
	"github.com/jackc/pgx/v5"
	"log"
	"strings"
)

type deleteURL struct {
	url    string
	userID uint32
}

type DeleteService struct {
	urlsForDelete chan []deleteURL
	conn          *pgx.Conn
	baseURL       string
	ctx           context.Context
}

func NewDeleteService(ctx context.Context, conn *pgx.Conn, baseURL string) *DeleteService {
	ds := &DeleteService{
		urlsForDelete: make(chan []deleteURL),
		conn:          conn,
		baseURL:       baseURL,
		ctx:           ctx,
	}
	for i := 0; i < 3; i++ {
		go func(ds *DeleteService) {
			for {
				data, ok := <-ds.urlsForDelete
				if !ok {
					return
				}
				err := ds.deleteURLs(data)
				if err != nil {
					log.Printf("Found err %s", err)
					ds.reAddURLs(data)
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

func (ds *DeleteService) reAddURLs(urls []deleteURL) {
	go func() {
		ds.urlsForDelete <- urls
	}()
}

func (ds *DeleteService) deleteURLs(urlsForDelete []deleteURL) error {
	if ds.conn == nil {
		return nil
	}
	tx, err := ds.conn.Begin(ds.ctx)
	if err != nil {
		log.Printf("tx error - %s", err)
		return err
	}
	q, err := createQueryForDelete(urlsForDelete)

	if err != nil {
		log.Printf("createQueryForDelete error - %s", err)
		return nil
	}

	_, err = tx.Exec(ds.ctx, q, urlsForDelete[0].userID)
	if err != nil {
		log.Printf("Exec error - %s", err)
		e := tx.Rollback(ds.ctx)
		if e != nil {
			return e
		}
		return nil
	}
	err = tx.Commit(ds.ctx)
	return err
}

func createQueryForDelete(urlsForDelete []deleteURL) (string, error) {
	buff := bytes.Buffer{}
	_, err := buff.WriteString("UPDATE shortener SET is_deleted = true WHERE shortener_id in (")
	if err != nil {
		return "", err
	}
	sep := ""
	for _, url := range urlsForDelete {
		buff.WriteString(sep)
		buff.WriteString(url.url)
		sep = ","
	}
	buff.WriteString(") AND user_id = $1;")
	return buff.String(), nil
}

func (ds *DeleteService) Close() {
	close(ds.urlsForDelete)
}

func getURLsFromArr(data string, userID uint32, baseURL string) []deleteURL {
	data = data[1 : len(data)-1]
	data = strings.ReplaceAll(data, "\"", "")
	splitData := strings.Split(data, ",")
	urls := make([]deleteURL, len(splitData))
	for i, url := range splitData {
		url = strings.TrimSpace(url)
		url = strings.TrimPrefix(url, baseURL+"/")
		urls[i] = deleteURL{url: url, userID: userID}
	}
	return urls
}
