// Package config define AppConfig for server configuration.
package config

import (
	"context"
	"encoding/json"
	"flag"
	"github.com/jackc/pgx/v5"
	"go-axesthump-shortener/internal/app/generator"
	"go-axesthump-shortener/internal/app/repository"
	"go-axesthump-shortener/internal/app/service"
	"go-axesthump-shortener/internal/app/util"
	"io"
	"log"
	"os"
	"strconv"
	"sync"
)

type ConfFile struct {
	ServerAddr      string `json:"server_addr"`
	BaseURL         string `json:"base_url"`
	FileStoragePass string `json:"file_storage_pass"`
	DBDsn           string `json:"database_dsn"`
	EnableHTTPS     bool   `json:"enable_https"`
}

// AppConfig contains data for configuration
type AppConfig struct {
	ServerAddr      string
	BaseURL         string
	Repo            repository.Repository
	DBContext       context.Context
	Conn            *pgx.Conn
	UserIDGenerator *generator.IDGenerator
	DeleteService   *service.DeleteService
	IsHTTPS         bool
	RequestWait     *sync.WaitGroup

	storagePath string
	dbConnURL   string
}

// NewAppConfig returns new AppConfig or error if it fails to create
// Creates and connects a repository based on the flags passed to the program.
func NewAppConfig() (*AppConfig, error) {
	appConfig := getServerConf()
	appConfig.RequestWait = &sync.WaitGroup{}
	setDBConn(appConfig)
	if err := setStorage(appConfig); err != nil {
		return nil, err
	}
	appConfig.DeleteService = service.NewDeleteService(appConfig.Repo, appConfig.BaseURL)
	return appConfig, nil
}

// setDBConn establishes a db connection
// If the database url was not passed to the program, then the installation will not occur and the execution will continue.
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

// createTable creates table shortener if not exists.
func createTable(config *AppConfig) {
	query := "CREATE TABLE IF NOT EXISTS shortener (shortener_id SERIAL PRIMARY KEY, long_url varchar(255) NOT NULL UNIQUE, user_id int NOT NULL, is_deleted BOOLEAN DEFAULT FALSE NOT NULL); CREATE INDEX IF NOT EXISTS idx_shortener_user_id ON shortener(user_id);"
	_, err := config.Conn.Exec(config.DBContext, query)
	if err != nil {
		panic(err)
	}
}

// setStorage a factory method that sets the required repository based on their configuration.
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

// getServerConf fills the configuration with the values from the set flags,
// if they are not present, then fills with the values from the environment changes,
// if they are also not present, then fills with empty strings.
func getServerConf() *AppConfig {

	serverAddr := flag.String(
		"a",
		"",
		"server address",
	)
	baseURL := flag.String(
		"b",
		"",
		"base url",
	)
	storagePath := flag.String(
		"f",
		"",
		"storage path",
	)
	dbConnect := flag.String(
		"d",
		"",
		"db conn",
	)
	isHTTPS := flag.String(
		"s",
		"",
		"https connection",
	)
	confFileShort := flag.String(
		"c",
		"",
		"config file",
	)

	confFileFull := flag.String(
		"config",
		"",
		"config file",
	)

	flag.Parse()

	appConfig := &AppConfig{}

	var confFilePath string
	if *confFileShort != "" {
		confFilePath = *confFileShort
	}
	if *confFileFull != "" {
		confFilePath = *confFileFull
	}

	confFile := &ConfFile{}
	if confFilePath != "" {
		confFileNew, err := getConfFile(confFilePath)
		if err == nil {
			confFile = confFileNew
		}
	}

	if *serverAddr == "" {
		servAddr := util.GetEnvOrDefault("SERVER_ADDRESS", "")
		if servAddr == "" {
			if confFile.ServerAddr == "" {
				appConfig.ServerAddr = "localhost:8080"
			} else {
				appConfig.ServerAddr = confFile.ServerAddr
			}
		}
	} else {
		appConfig.ServerAddr = *serverAddr
	}

	if *baseURL == "" {
		envBaseURL := util.GetEnvOrDefault("BASE_URL", "")
		if envBaseURL == "" {
			if confFile.BaseURL == "" {
				appConfig.BaseURL = "http://localhost:8080"
			} else {
				appConfig.BaseURL = confFile.BaseURL
			}
		}
	} else {
		appConfig.BaseURL = *baseURL
	}

	if *storagePath == "" {
		envStoragePath := os.Getenv("FILE_STORAGE_PATH")
		if envStoragePath == "" {
			if confFile.FileStoragePass != "" {
				envStoragePath = appConfig.storagePath
			}
			appConfig.storagePath = envStoragePath
		}
	} else {
		appConfig.storagePath = *storagePath
	}

	if *dbConnect == "" {
		envDBConnect := os.Getenv("DATABASE_DSN")
		if envDBConnect == "" {
			if confFile.DBDsn != "" {
				envDBConnect = appConfig.storagePath
			}
			appConfig.dbConnURL = envDBConnect
		}
	} else {
		appConfig.dbConnURL = *dbConnect
	}

	if *isHTTPS == "" {
		envIsHTTPS := os.Getenv("ENABLE_HTTPS")
		if envIsHTTPS != "" {
			b, err := strconv.ParseBool(envIsHTTPS)
			if err != nil {
				appConfig.IsHTTPS = false
			} else {
				appConfig.IsHTTPS = b
			}
		} else {
			appConfig.IsHTTPS = confFile.EnableHTTPS
		}
	}

	return appConfig
}

// getConfFile returns ConfFile
func getConfFile(path string) (*ConfFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	confFile := &ConfFile{}
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, confFile)
	if err != nil {
		return nil, err
	}

	err = f.Close()
	if err != nil {
		return nil, err
	}

	return confFile, nil
}
