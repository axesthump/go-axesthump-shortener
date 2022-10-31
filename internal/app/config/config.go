package config

import (
	"context"
	"flag"
	"github.com/jackc/pgx/v5"
	"go-axesthump-shortener/internal/app/repository"
	"go-axesthump-shortener/internal/app/util"
	"os"
)

type AppConfig struct {
	ServerAddr  string
	BaseURL     string
	Repo        repository.Repository
	storagePath string
	dbConnURL   string
	DBContext   context.Context
	Conn        *pgx.Conn
}

func CreateAppConfig() (*AppConfig, error) {
	appConfig := getConsoleArgs()
	setParametersFromEnv(appConfig)
	if err := setStorage(appConfig); err != nil {
		return nil, err
	}
	setDBConn(appConfig)
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
}

func setStorage(config *AppConfig) error {
	if len(config.storagePath) == 0 {
		config.Repo = repository.NewInMemoryStorage()
	} else {
		var err error
		config.Repo, err = repository.NewLocalStorage(config.storagePath)
		if err != nil {
			return err
		}
	}
	return nil
}

func setParametersFromEnv(config *AppConfig) {
	if config.BaseURL == "" {
		config.BaseURL = util.GetEnvOrDefault("BASE_URL", "http://localhost:8080")
	}
	if config.ServerAddr == "" {
		config.ServerAddr = util.GetEnvOrDefault("SERVER_ADDRESS", "localhost:8080")
	}
	if config.storagePath == "" {
		config.storagePath = os.Getenv("FILE_STORAGE_PATH")
	}
	if config.dbConnURL == "" {
		config.storagePath = os.Getenv("DATABASE_DSN")
	}
}

func getConsoleArgs() *AppConfig {
	serverAddr := flag.String("a", "", "server address")
	baseURL := flag.String("b", "", "base url")
	storagePath := flag.String("f", "", "storage path")
	dbConnect := flag.String("d", "", "db conn")
	flag.Parse()

	return &AppConfig{
		ServerAddr:  *serverAddr,
		BaseURL:     *baseURL,
		storagePath: *storagePath,
		dbConnURL:   *dbConnect,
	}
}
