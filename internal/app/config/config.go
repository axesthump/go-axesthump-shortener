package config

import (
	"flag"
	"go-axesthump-shortener/internal/app/repository"
	"go-axesthump-shortener/internal/app/util"
	"os"
)

type AppConfig struct {
	ServerAddr  string
	BaseURL     string
	Repo        repository.Repository
	storagePath string
}

func CreateAppConfig() (*AppConfig, error) {
	appConfig := getConsoleArgs()
	if appConfig.BaseURL == "" {
		appConfig.BaseURL = util.GetEnvOrDefault("BASE_URL", "http://localhost:8080")
	}
	if appConfig.ServerAddr == "" {
		appConfig.ServerAddr = util.GetEnvOrDefault("SERVER_ADDRESS", "localhost:8080")
	}
	if appConfig.storagePath == "" {
		os.Getenv("FILE_STORAGE_PATH")
	}

	var err error
	if len(appConfig.storagePath) == 0 {
		appConfig.Repo = repository.NewInMemoryStorage()
	} else {
		appConfig.Repo, err = repository.NewLocalStorage(appConfig.storagePath)
	}
	if err != nil {
		return nil, err
	}
	return appConfig, nil
}

func getConsoleArgs() *AppConfig {
	serverAddr := flag.String("a", "", "server address")
	baseURL := flag.String("b", "", "base url")
	storagePath := flag.String("f", "", "storage path")

	return &AppConfig{
		ServerAddr:  *serverAddr,
		BaseURL:     *baseURL,
		storagePath: *storagePath,
	}
}
