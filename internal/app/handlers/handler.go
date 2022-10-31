package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	myMiddleware "go-axesthump-shortener/internal/app/middleware"
	"go-axesthump-shortener/internal/app/repository"
	"io"
	"log"
	"net/http"
	"strconv"
)

type AppHandler struct {
	repo    repository.Repository
	baseURL string
	Router  chi.Router
}

type requestURL struct {
	URL string `json:"url"`
}

type response struct {
	Result string `json:"result"`
}

func NewRouter(appHandler *AppHandler) chi.Router {
	r := chi.NewRouter()
	r.Use(myMiddleware.NewAuthService().Auth)
	r.Use(myMiddleware.Gzip)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/{shortURL}", appHandler.getURL())
	r.Post("/", appHandler.addURL())

	r.Route("/api", func(r chi.Router) {
		r.Post("/shorten", appHandler.addURLRest())
		r.Get("/user/urls", appHandler.listURLs())
	})

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

func (a *AppHandler) addURLRest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")

		body, err := readBody(w, r.Body)
		if err != nil {
			return
		}

		var requestURL requestURL

		err = json.Unmarshal(body, &requestURL)
		if err != nil || len(requestURL.URL) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		userID := r.Context().Value("id").(uint32)
		shortURL, err := a.repo.CreateShortURL(a.baseURL, requestURL.URL, userID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		resp := response{Result: shortURL}
		buf := bytes.NewBuffer([]byte{})
		encoder := json.NewEncoder(buf)
		encoder.SetEscapeHTML(false)
		err = encoder.Encode(resp)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		sendResponse(w, buf.Bytes())
	}
}

func (a *AppHandler) addURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/plain")

		body, err := readBody(w, r.Body)
		if err != nil {
			return
		}
		url := string(body)
		userID := r.Context().Value("id").(uint32)
		shortURL, _ := a.repo.CreateShortURL(a.baseURL, url, userID)

		sendResponse(w, []byte(shortURL))
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
		userID := r.Context().Value("id").(uint32)
		fullURL, err := a.repo.GetFullURL(shortURL, userID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Location", fullURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func (a *AppHandler) listURLs() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		userID := r.Context().Value("id").(uint32)
		urls := a.repo.GetAllURLs(a.baseURL, userID)

		log.Printf("Urls len - %d\n", len(urls))

		if len(urls) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		var resp []byte
		var err error
		if resp, err = json.Marshal(&urls); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		_, err = w.Write(resp)
		if err != nil {
			return
		}
	}
}

func sendResponse(w http.ResponseWriter, res []byte) {
	w.WriteHeader(http.StatusCreated)
	_, err := w.Write(res)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func readBody(w http.ResponseWriter, body io.ReadCloser) ([]byte, error) {
	defer body.Close()
	bodyBytes, err := io.ReadAll(body)
	if err != nil || len(bodyBytes) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return nil, err
	}
	return bodyBytes, nil
}
