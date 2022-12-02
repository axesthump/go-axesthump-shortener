package config

import (
	"github.com/stretchr/testify/assert"
	"go-axesthump-shortener/internal/app/repository"
	"os"
	"testing"
)

func TestNewAppConfig(t *testing.T) {

}

func Test_getConsoleArgs_with_flags(t *testing.T) {
	type testData struct {
		serverAddr  string
		baseURL     string
		storagePath string
		dbConnURL   string
		needEnv     bool
	}
	tests := []struct {
		name string
		td   testData
	}{
		{
			name: "Test getConsoleArgs from env",
			td: testData{
				serverAddr:  "localhost:8080",
				baseURL:     "http://localhost:8080",
				storagePath: "./storage",
				dbConnURL:   "mysql/some/conn",
				needEnv:     true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.td.needEnv {
				os.Setenv("SERVER_ADDRESS", tt.td.serverAddr)
				os.Setenv("BASE_URL", tt.td.baseURL)
				os.Setenv("FILE_STORAGE_PATH", tt.td.storagePath)
				os.Setenv("DATABASE_DSN", tt.td.dbConnURL)
			}
			conf := getConsoleArgs()
			assert.Equal(t, tt.td.serverAddr, conf.ServerAddr)
			assert.Equal(t, tt.td.baseURL, conf.BaseURL)
			assert.Equal(t, tt.td.storagePath, conf.storagePath)
			assert.Equal(t, tt.td.dbConnURL, conf.dbConnURL)
		})
	}
}

func Test_setStorage(t *testing.T) {
	type testData struct {
		conf              *AppConfig
		needRemoveFile    bool
		needError         bool
		isLocalStorage    bool
		isInMemoryStorage bool
		isDBStorage       bool
	}
	tests := []struct {
		name string
		td   testData
	}{
		{
			name: "Test setStorage with localStorage",
			td: testData{
				conf: &AppConfig{
					ServerAddr:  "localhost:8080",
					BaseURL:     "http://localhost:8080",
					storagePath: "./storage",
					Conn:        nil,
				},
				needRemoveFile:    true,
				needError:         false,
				isLocalStorage:    true,
				isDBStorage:       false,
				isInMemoryStorage: false,
			},
		},
		{
			name: "Test setStorage with inMemoryStorage",
			td: testData{
				conf: &AppConfig{
					ServerAddr:  "localhost:8080",
					BaseURL:     "http://localhost:8080",
					storagePath: "",
					Conn:        nil,
				},
				needRemoveFile:    false,
				needError:         false,
				isLocalStorage:    false,
				isDBStorage:       false,
				isInMemoryStorage: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := setStorage(tt.td.conf)
			if tt.td.needError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			switch {
			case tt.td.isLocalStorage:
				_, ok := tt.td.conf.Repo.(*repository.LocalStorage)
				assert.True(t, ok)
			case tt.td.isDBStorage:
				_, ok := tt.td.conf.Repo.(*repository.DBStorage)
				assert.True(t, ok)
			case tt.td.isInMemoryStorage:
				_, ok := tt.td.conf.Repo.(*repository.InMemoryStorage)
				assert.True(t, ok)
			}
			if tt.td.needRemoveFile {
				err = os.Remove(tt.td.conf.storagePath[2:])
				assert.NoError(t, err)
			}
		})
	}

}
