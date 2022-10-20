package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go-axesthump-shortener/internal/app/repository"
	"io"
	"net/http"
	"strconv"
)

type AppHandler struct {
	repo    repository.Repository
	baseURL string
	Router  chi.Router
}

func NewRouter(appHandler *AppHandler) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Get("/{shortURL}", appHandler.getURL())
	r.Post("/", appHandler.addURL())

	return r
}

func NewAppHandler(baseURL string, repo repository.Repository) *AppHandler {
	h := &AppHandler{
		repo:    repo,
		baseURL: baseURL,
	}
	h.Router = NewRouter(h)
	return h
}

func (a *AppHandler) addURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain")

		defer r.Body.Close()
		body, err := io.ReadAll(r.Body)
		if err != nil || len(body) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		url := string(body)
		shortURL := a.repo.CreateShortURL(a.baseURL, url)

		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte(shortURL))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
}

func (a *AppHandler) getURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url := chi.URLParam(r, "shortURL")

		shortURL, err := strconv.ParseInt(url, 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		fullURL, err := a.repo.GetFullURL(shortURL)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Location", fullURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}

}
