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
		a.getUrl(w, r)
	case http.MethodPost:
		a.addUrl(w, r)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (a *app) addUrl(w http.ResponseWriter, r *http.Request) {
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
	response := strconv.FormatInt(a.storage.CreateShortURL(url), 10)

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(response))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

}

func (a *app) getUrl(w http.ResponseWriter, r *http.Request) {
	url := strings.Split(r.URL.Path, "/")
	if len(url) != 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortUrl, err := strconv.ParseInt(url[0], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fullUrl, err := a.storage.GetFullURL(shortUrl)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusTemporaryRedirect)
	w.Header().Set("Location", fullUrl)

	return
}

func main() {
	mux := http.NewServeMux()
	storage := app{storage: service.NewStorage()}
	mux.HandleFunc("/", storage.handleRequest)
	log.Fatal(http.ListenAndServe(":8080", mux))
}
