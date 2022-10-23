package main

import (
	"go-axesthump-shortener/internal/app/handlers"
	"go-axesthump-shortener/internal/app/repository"
	"go-axesthump-shortener/internal/app/util"
	"log"
	"net/http"
)

func main() {
	serverAddr := util.GetEnvOrDefault("SERVER_ADDRESS", "localhost:8080")
	baseURL := util.GetEnvOrDefault("BASE_URL", "http://localhost:8080/")

	appHandler := handlers.NewAppHandler(baseURL, repository.NewStorage())
	log.Fatal(http.ListenAndServe(serverAddr, appHandler.Router))
}
