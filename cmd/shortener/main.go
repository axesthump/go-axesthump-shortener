package main

import (
	"go-axesthump-shortener/internal/app/service"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type app struct {
	storage *service.Storage
}

func (a *app) handleRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.getURL(w, r)
	case http.MethodPost:
		a.addURL(w, r)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (a *app) addURL(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	url := string(body)
	shortURL, err := a.storage.CreateShortURL(url)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(shortURL))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

}

func (a *app) getURL(w http.ResponseWriter, r *http.Request) {
	url := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
	if len(url) != 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortURL, err := strconv.ParseInt(url[0], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fullURL, err := a.storage.GetFullURL(shortURL)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.Header().Set("Location", fullURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func main() {
	mux := http.NewServeMux()
	storage := app{storage: service.NewStorage()}
	mux.HandleFunc("/", storage.handleRequest)
	log.Fatal(http.ListenAndServe(":8080", mux))
}
