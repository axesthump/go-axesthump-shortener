package main

import (
	"go-axesthump-shortener/internal/app/handlers"
	"go-axesthump-shortener/internal/app/repository"
	"log"
	"net/http"
)

const (
	serverURL = "localhost:8080"
	protocol  = "http://"
)

func main() {
	appHandler := handlers.NewAppHandler(protocol+serverURL+"/", repository.NewStorage())
	log.Fatal(http.ListenAndServe(serverURL, appHandler.Router))
}
