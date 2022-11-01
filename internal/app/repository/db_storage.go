package repository

import (
	"context"
	"github.com/jackc/pgx/v5"
	"log"
	"strconv"
	"sync"
)

type dbStorage struct {
	conn *pgx.Conn
	sync.RWMutex
	lastID int64
	ctx    context.Context
}

func NewDBStorage(ctx context.Context, conn *pgx.Conn) *dbStorage {
	db := &dbStorage{
		conn: conn,
		ctx:  ctx,
	}
	db.lastID = db.getLastID()
	return db
}

func (db *dbStorage) getLastID() int64 {
	query := "SELECT shortener_id FROM shortener ORDER BY shortener_id DESC LIMIT 1;"
	row := db.conn.QueryRow(context.Background(), query)
	lastID := 0
	err := row.Scan(&lastID)
	if err != nil {
		lastID = 0
	} else {
		lastID++
	}
	return int64(lastID)
}

func (db *dbStorage) GetLastUserID() int64 {
	query := "SELECT user_id FROM shortener ORDER BY user_id DESC LIMIT 1;"
	row := db.conn.QueryRow(context.Background(), query)
	lastID := 0
	err := row.Scan(&lastID)
	if err != nil {
		lastID = 0
	}
	lastID++
	return int64(lastID)
}

func (db *dbStorage) CreateShortURL(
	ctx context.Context,
	beginURL string,
	originalURL string,
	userID uint32,
) (string, error) {
	db.Lock()
	shortEndpoint := strconv.FormatInt(db.lastID, 10)
	shortURL := beginURL + shortEndpoint
	db.lastID++
	db.Unlock()
	query := "INSERT INTO shortener (short_url, long_url, user_id) VALUES ($1, $2, $3);"
	_, err := db.conn.Exec(ctx, query, shortEndpoint, originalURL, userID)
	if err != nil {
		return "", err
	}
	return shortURL, nil
}

func (db *dbStorage) GetFullURL(ctx context.Context, shortURL int64) (string, error) {
	query := "SELECT long_url FROM shortener WHERE short_url = $1"
	row := db.conn.QueryRow(ctx, query, shortURL)
	var longURL = new(string)
	err := row.Scan(longURL)
	if err != nil {
		return "", err
	}
	return *longURL, nil
}

func (db *dbStorage) GetAllURLs(ctx context.Context, beginURL string, userID uint32) []URLInfo {
	query := "SELECT short_url, long_url FROM shortener WHERE user_id = $1"
	row, err := db.conn.Query(ctx, query, userID)
	if err != nil {
		return []URLInfo{}
	}
	urls := make([]URLInfo, 0)
	for row.Next() {
		var shortURL = new(string)
		var longURL = new(string)
		err := row.Scan(shortURL, longURL)
		if err != nil {
			log.Printf("Cant scan!")
			return []URLInfo{}
		}
		urls = append(urls, URLInfo{
			ShortURL:    *shortURL,
			OriginalURL: *longURL,
		})
	}
	return urls
}

func (db *dbStorage) CreateShortURLs(
	ctx context.Context,
	beginURL string,
	urls []URLWithID,
	userID uint32,
) ([]URLWithID, error) {
	tx, err := db.conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err = tx.Prepare(
		ctx, "insert", "INSERT INTO shortener (short_url, long_url, user_id) VALUES ($1, $2, $3);",
	); err != nil {
		return nil, err
	}

	res := make([]URLWithID, 0, len(urls))
	db.Lock()
	defer db.Unlock()
	for _, url := range urls {
		shortEndpoint := strconv.FormatInt(db.lastID, 10)
		shortURL := beginURL + shortEndpoint
		db.lastID++
		_, err := tx.Exec(ctx, "insert", shortEndpoint, url.URL, userID)
		if err != nil {
			return nil, err
		}
		res = append(res, URLWithID{
			CorrelationID: url.CorrelationID,
			URL:           shortURL,
		})
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return res, nil
}

func (db *dbStorage) Close() error {
	return db.conn.Close(db.ctx)
}
