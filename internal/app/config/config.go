package config

import (
	"context"
	"flag"
	"github.com/jackc/pgx/v5"
	"go-axesthump-shortener/internal/app/generator"
	"go-axesthump-shortener/internal/app/repository"
	"go-axesthump-shortener/internal/app/service"
	"go-axesthump-shortener/internal/app/util"
	"log"
	"os"
)

type AppConfig struct {
	ServerAddr      string
	BaseURL         string
	Repo            repository.Repository
	DBContext       context.Context
	Conn            *pgx.Conn
	UserIDGenerator *generator.IDGenerator
	DeleteService   *service.DeleteService

	storagePath string
	dbConnURL   string
}

func NewAppConfig() (*AppConfig, error) {
	appConfig := getConsoleArgs()
	setDBConn(appConfig)
	if err := setStorage(appConfig); err != nil {
		return nil, err
	}
	appConfig.DeleteService = service.NewDeleteService(appConfig.Repo, appConfig.BaseURL)
	return appConfig, nil
}

func setDBConn(config *AppConfig) {
	if len(config.dbConnURL) == 0 {
		config.Conn = nil
		return
	}

	config.DBContext = context.Background()
	conn, err := pgx.Connect(config.DBContext, config.dbConnURL)
	if err != nil {
		config.Conn = nil
		return
	}
	config.Conn = conn
	createTable(config)
}

func createTable(config *AppConfig) {
	query := "CREATE TABLE IF NOT EXISTS shortener (shortener_id SERIAL PRIMARY KEY, long_url varchar(255) NOT NULL UNIQUE, user_id int NOT NULL, is_deleted BOOLEAN DEFAULT FALSE NOT NULL); CREATE INDEX IF NOT EXISTS idx_shortener_user_id ON shortener(user_id);"
	_, err := config.Conn.Exec(config.DBContext, query)
	if err != nil {
		panic(err)
	}
}

func setStorage(config *AppConfig) error {
	var lastUserID uint32
	switch {
	case config.Conn != nil:
		log.Printf("Use db repository!")
		db := repository.NewDBStorage(config.DBContext, config.Conn)
		config.Repo = db
		lastUserID = uint32(db.GetLastUserID())
	case len(config.storagePath) != 0:
		log.Printf("Use localStorage repository!")
		localStorage, err := repository.NewLocalStorage(config.storagePath)
		if err != nil {
			return err
		}
		config.Repo = localStorage
		lastUserID = localStorage.GetUserLastID()
	default:
		log.Printf("Use inMemory repository!")
		config.Repo = repository.NewInMemoryStorage()
		lastUserID = 0
	}
	config.UserIDGenerator = generator.NewIDGenerator(int64(lastUserID))
	return nil
}

func getConsoleArgs() *AppConfig {
	serverAddr := flag.String(
		"a",
		util.GetEnvOrDefault("SERVER_ADDRESS", "localhost:8080"),
		"server address",
	)
	baseURL := flag.String(
		"b",
		util.GetEnvOrDefault("BASE_URL", "http://localhost:8080"),
		"base url",
	)
	storagePath := flag.String(
		"f",
		os.Getenv("FILE_STORAGE_PATH"),
		"storage path",
	)
	dbConnect := flag.String(
		"d",
		os.Getenv("DATABASE_DSN"),
		"db conn",
	)
	flag.Parse()

	return &AppConfig{
		ServerAddr:  *serverAddr,
		BaseURL:     *baseURL,
		storagePath: *storagePath,
		dbConnURL:   *dbConnect,
	}
}
