package repository

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
	"github.com/lib/pq"
	"log"
	"strconv"
)

// LongURLConflictError an error that occurs when the original urls conflict.
type LongURLConflictError struct {
}

// Error return error description.
func (e *LongURLConflictError) Error() string {
	return "LongURL conflict"
}

// DBStorage contains data for db.
type DBStorage struct {
	conn *pgx.Conn
	ctx  context.Context
}

// NewDBStorage returns new DBStorage.
func NewDBStorage(ctx context.Context, conn *pgx.Conn) *DBStorage {
	db := &DBStorage{
		conn: conn,
		ctx:  ctx,
	}
	return db
}

// GetLastUserID return last created user id.
func (db *DBStorage) GetLastUserID() int64 {
	query := "SELECT MAX(user_id) FROM shortener;"
	row := db.conn.QueryRow(db.ctx, query)
	lastID := 0
	err := row.Scan(&lastID)
	if err != nil {
		lastID = 0
	} else {
		lastID++
	}
	return int64(lastID)
}

// CreateShortURL create short url. Returns short url if operations success or error.
func (db *DBStorage) CreateShortURL(
	ctx context.Context,
	beginURL string,
	originalURL string,
	userID uint32,
) (string, error) {
	var shortEndpoint int64
	query := "INSERT INTO shortener (long_url, user_id) VALUES ($1, $2) ON CONFLICT (long_url) DO NOTHING RETURNING shortener_id;"
	row := db.conn.QueryRow(ctx, query, originalURL, userID)
	err := row.Scan(&shortEndpoint)
	shortURL := ""
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			var short int64
			short, err = db.GetShortURLByFullURL(ctx, originalURL)
			if err != nil {
				return "", err
			}
			strShort := strconv.FormatInt(short, 10)
			shortURL = beginURL + strShort
			err = &LongURLConflictError{}
		} else {
			return "", err
		}
	} else {
		strShort := strconv.FormatInt(shortEndpoint, 10)
		shortURL = beginURL + strShort
	}
	return shortURL, err
}

// GetShortURLByFullURL return short url from full url.
func (db *DBStorage) GetShortURLByFullURL(ctx context.Context, fullURL string) (int64, error) {
	query := "SELECT shortener_id FROM shortener WHERE long_url = $1"
	row := db.conn.QueryRow(ctx, query, fullURL)
	var shortURL = new(int64)
	err := row.Scan(shortURL)
	if err != nil {
		return 0, err
	}
	return *shortURL, nil
}

// GetFullURL returns full url by short url.
func (db *DBStorage) GetFullURL(ctx context.Context, shortURL int64) (string, error) {
	query := "SELECT long_url, is_deleted FROM shortener WHERE shortener_id = $1"
	row := db.conn.QueryRow(ctx, query, shortURL)
	var longURL = new(string)
	var isDeleted = new(bool)
	err := row.Scan(longURL, isDeleted)
	if err != nil {
		return "", err
	}
	if *isDeleted {
		return "", &DeletedURLError{}
	}
	return *longURL, nil
}

// GetAllURLs returns all urls owned specific user.
func (db *DBStorage) GetAllURLs(ctx context.Context, beginURL string, userID uint32) []URLInfo {
	query := "SELECT shortener_id, long_url FROM shortener WHERE user_id = $1"
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

// CreateShortURLs create short urls. Returns slice short urls if operations success or error.
func (db *DBStorage) CreateShortURLs(
	ctx context.Context,
	beginURL string,
	urls []URLWithID,
	userID uint32,
) ([]URLWithID, error) {
	tx, err := db.conn.Begin(ctx)
	if err != nil {
		return nil, err
	}

	if _, err = tx.Prepare(
		ctx, "insert", "INSERT INTO shortener (long_url, user_id) VALUES ($1, $2) RETURNING shortener_id;",
	); err != nil {
		return nil, err
	}

	res := make([]URLWithID, 0, len(urls))
	for _, url := range urls {
		var shortEndpoint int64
		row := tx.QueryRow(ctx, "insert", url.URL, userID)
		err = row.Scan(&shortEndpoint)
		if err != nil {
			if err = tx.Rollback(ctx); err != nil {
				log.Printf("update drivers: unable to rollback: %v\n", err)
				return nil, err
			}
			return nil, err
		}
		shortURL := beginURL + strconv.FormatInt(shortEndpoint, 10)
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

// DeleteURLs delete url from urlsForDelete.
func (db *DBStorage) DeleteURLs(urlsForDelete []DeleteURL) error {
	tx, err := db.conn.Begin(db.ctx)
	if err != nil {
		log.Printf("tx error - %s", err)
		return err
	}
	q := "UPDATE shortener SET is_deleted = true WHERE shortener_id = ANY ($1) AND user_id = $2;"
	shortIDs := convertShortIDs(urlsForDelete)

	_, err = tx.Exec(db.ctx, q, shortIDs, urlsForDelete[0].UserID)
	if err != nil {
		log.Printf("Exec error - %s", err)
		e := tx.Rollback(db.ctx)
		if e != nil {
			return e
		}
		return nil
	}
	err = tx.Commit(db.ctx)
	return err
}

// Close closes everything that should be closed in the context of the repository.
func (db *DBStorage) Close() error {
	return db.conn.Close(db.ctx)
}

// convertShortIDs create array with short ids.
func convertShortIDs(urlsForDelete []DeleteURL) pq.Int64Array {
	shortIDs := pq.Int64Array{}
	for _, url := range urlsForDelete {
		shortID, err := strconv.ParseInt(url.URL, 10, 64)
		if err != nil {
			continue
		}
		shortIDs = append(shortIDs, shortID)
	}
	return shortIDs
}
