package main

import (
	"go-axesthump-shortener/internal/app/handlers"
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	app := handlers.NewAppHandler()
	mux.HandleFunc("/", app.HandleRequest)
	log.Fatal(http.ListenAndServe(":8080", mux))
}
