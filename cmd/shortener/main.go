package main

import (
	"go-axesthump-shortener/internal/app/config"
	"go-axesthump-shortener/internal/app/handlers"
	"go-axesthump-shortener/internal/app/repository"
	"log"
	"net/http"
)

func main() {
	conf, err := config.CreateAppConfig()
	if err != nil {
		panic(err)
	}
	appHandler := handlers.NewAppHandler(conf.BaseURL+"/", repository.NewInMemoryStorage())
	log.Fatal(http.ListenAndServe(conf.ServerAddr, appHandler.Router))
}
