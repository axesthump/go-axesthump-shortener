package config

import (
	"go-axesthump-shortener/internal/app/repository"
	"go-axesthump-shortener/internal/app/util"
	"os"
)

type AppConfig struct {
	ServerAddr string
	BaseURL    string
	Repo       repository.Repository
}

func CreateAppConfig() (*AppConfig, error) {
	serverAddr := util.GetEnvOrDefault("SERVER_ADDRESS", "localhost:8080")
	baseURL := util.GetEnvOrDefault("BASE_URL", "http://localhost:8080")
	storagePath := os.Getenv("FILE_STORAGE_PATH")

	var repo repository.Repository
	var err error
	if len(storagePath) == 0 {
		repo = repository.NewInMemoryStorage()
	} else {
		repo, err = repository.NewLocalStorage(storagePath)
	}
	if err != nil {
		return nil, err
	}
	return &AppConfig{
		ServerAddr: serverAddr,
		BaseURL:    baseURL,
		Repo:       repo,
	}, nil
}
