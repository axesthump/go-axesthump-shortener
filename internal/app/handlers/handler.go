package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"go-axesthump-shortener/internal/app/config"
	"go-axesthump-shortener/internal/app/generator"
	myMiddleware "go-axesthump-shortener/internal/app/middleware"
	"go-axesthump-shortener/internal/app/repository"
	"go-axesthump-shortener/internal/app/service"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

type AppHandler struct {
	repo            repository.Repository
	userIDGenerator *generator.IDGenerator
	baseURL         string
	dbConn          *pgx.Conn
	deleteService   *service.DeleteService
	Router          chi.Router
}

type requestURL struct {
	URL string `json:"url"`
}

type response struct {
	Result string `json:"result"`
}

type addListURLsRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type addListURLsResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func NewRouter(appHandler *AppHandler) chi.Router {
	r := chi.NewRouter()
	r.Use(myMiddleware.NewAuthService(appHandler.userIDGenerator).Auth)
	r.Use(myMiddleware.Gzip)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/ping", appHandler.ping)
	r.Get("/{shortURL}", appHandler.getURL)
	r.Post("/", appHandler.addURL)

	r.Route("/api", func(r chi.Router) {
		r.Route("/shorten", func(r chi.Router) {
			r.Post("/", appHandler.addURLRest)
			r.Post("/batch", appHandler.addListURLRest)
		})
		r.Route("/user/urls", func(r chi.Router) {
			r.Get("/", appHandler.listURLs)
			r.Delete("/", appHandler.deleteListURLs)
		})
	})

	return r
}

func NewAppHandler(config *config.AppConfig) *AppHandler {
	h := &AppHandler{
		repo:            config.Repo,
		baseURL:         config.BaseURL + "/",
		dbConn:          config.Conn,
		userIDGenerator: config.UserIDGenerator,
		deleteService:   config.DeleteService,
	}
	h.Router = NewRouter(h)
	return h
}

func (a *AppHandler) addURLRest(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	var body []byte
	var err error
	if body, err = readBody(w, r.Body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var requestURL requestURL
	if err = json.Unmarshal(body, &requestURL); err != nil || len(requestURL.URL) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	userID := r.Context().Value(myMiddleware.UserIDKey).(uint32)
	status := http.StatusCreated
	var shortURL string
	if shortURL, err = a.repo.CreateShortURL(r.Context(), a.baseURL, requestURL.URL, userID); err != nil {
		if errors.Is(err, &repository.LongURLConflictError{}) {
			status = http.StatusConflict
		} else {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	buf, err := a.createAddURLResponse(w, shortURL)
	if err != nil {
		return
	}
	sendResponse(w, buf, status)
}

func (a *AppHandler) createAddURLResponse(w http.ResponseWriter, shortURL string) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)
	resp := response{Result: shortURL}
	if err := encoder.Encode(resp); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, err
	}
	return buf.Bytes(), nil
}

func (a *AppHandler) addURL(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain")

	body, err := readBody(w, r.Body)
	if err != nil {
		return
	}
	url := string(body)
	userID := r.Context().Value(myMiddleware.UserIDKey).(uint32)
	shortURL, err := a.repo.CreateShortURL(r.Context(), a.baseURL, url, userID)
	status := http.StatusCreated
	if err != nil {
		if errors.Is(err, &repository.LongURLConflictError{}) {
			status = http.StatusConflict
		} else {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	sendResponse(w, []byte(shortURL), status)
}

func (a *AppHandler) getURL(w http.ResponseWriter, r *http.Request) {
	url := chi.URLParam(r, "shortURL")

	shortURL, err := strconv.ParseInt(url, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	fullURL, err := a.repo.GetFullURL(r.Context(), shortURL)
	if err != nil {
		if errors.Is(err, &repository.DeletedURLError{}) {
			w.WriteHeader(http.StatusGone)
			return
		} else {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}
	w.Header().Set("Location", fullURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (a *AppHandler) listURLs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	userID := r.Context().Value(myMiddleware.UserIDKey).(uint32)
	urls := a.repo.GetAllURLs(r.Context(), a.baseURL, userID)

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

func (a *AppHandler) deleteListURLs(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(myMiddleware.UserIDKey).(uint32)
	body, err := readBody(w, r.Body)
	if err != nil {
		return
	}
	a.deleteService.AddURLs(string(body), userID)
	w.WriteHeader(http.StatusAccepted)
}

func (a *AppHandler) ping(w http.ResponseWriter, r *http.Request) {
	if a.dbConn == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := a.dbConn.Ping(ctx); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (a *AppHandler) addListURLRest(w http.ResponseWriter, r *http.Request) {
	body, err := readBody(w, r.Body)
	if err != nil {
		return
	}
	userID := r.Context().Value(myMiddleware.UserIDKey).(uint32)
	var urlsForShort []addListURLsRequest
	if err := json.Unmarshal(body, &urlsForShort); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	convertedURLs := make([]repository.URLWithID, 0, len(urlsForShort))
	for _, url := range urlsForShort {
		convertedURLs = append(convertedURLs, repository.URLWithID{
			CorrelationID: url.CorrelationID,
			URL:           url.OriginalURL,
		})
	}

	shortenURLs, err := a.repo.CreateShortURLs(r.Context(), a.baseURL, convertedURLs, userID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortenURLsResponse := make([]addListURLsResponse, 0, len(shortenURLs))
	for _, shortenURL := range shortenURLs {
		shortenURLsResponse = append(shortenURLsResponse, addListURLsResponse{
			CorrelationID: shortenURL.CorrelationID,
			ShortURL:      shortenURL.URL,
		})
	}

	resBody, err := json.Marshal(&shortenURLsResponse)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, err = w.Write(resBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}

func sendResponse(w http.ResponseWriter, res []byte, status int) {
	w.WriteHeader(status)
	log.Printf("Response: %s\n", res)
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
