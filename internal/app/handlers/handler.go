package handlers

import (
	"fmt"
	"go-axesthump-shortener/internal/app/service"
	"io"
	"net/http"
	"strconv"
	"strings"
)

const (
	serverURL = "localhost:8080"
	protocol  = "http://"
)

type AppHandler struct {
	storage *service.Storage
}

func NewAppHandler() *AppHandler {
	return &AppHandler{
		storage: service.NewStorage(),
	}
}

func (a *AppHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		a.getURL(w, r)
	case http.MethodPost:
		a.addURL(w, r)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (a *AppHandler) addURL(w http.ResponseWriter, r *http.Request) {
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
	fmt.Println(r.URL.Host)
	shortURL, err := a.storage.CreateShortURL(protocol+serverURL+"/", url)
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

func (a *AppHandler) getURL(w http.ResponseWriter, r *http.Request) {
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
