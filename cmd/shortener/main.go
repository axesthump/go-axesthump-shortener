package main

import (
	"github.com/caarlos0/env/v6"
	"go-axesthump-shortener/internal/app/handlers"
	"go-axesthump-shortener/internal/app/repository"
	"log"
	"net/http"
)

type serverConfig struct {
	ServerAddr string `env:"SERVER_ADDRESS,required"`
	BaseURL    string `env:"BASE_URL,required"`
}

func main() {
	var serverConfig serverConfig
	if err := env.Parse(&serverConfig); err != nil {
		panic(err)
	}

	appHandler := handlers.NewAppHandler(serverConfig.BaseURL+"/", repository.NewStorage())
	log.Fatal(http.ListenAndServe(serverConfig.ServerAddr, appHandler.Router))
}
