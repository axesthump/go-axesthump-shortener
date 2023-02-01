// Package handlers define AppHandler for handle app requests.
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
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// AppHandler contains tools to work with requests.
type AppHandler struct {
	repo            repository.Repository
	userIDGenerator *generator.IDGenerator
	baseURL         string
	dbConn          *pgx.Conn
	deleteService   *service.DeleteService
	Router          chi.Router
	wg              *sync.WaitGroup
	trustedSubnet   string
}

type (
	// arrURLRequest url shortening request data.
	arrURLRequest struct {
		// URL - url for shortening.
		URL string `json:"url"`
	}

	// addURLResponse url shortening response.
	addURLResponse struct {
		// Result - shorten url.
		Result string `json:"result"`
	}

	// addListURLsRequest urls shortening request data.
	addListURLsRequest struct {
		// CorrelationID - url id.
		CorrelationID string `json:"correlation_id"`
		// OriginalURL - url for shortening.
		OriginalURL string `json:"original_url"`
	}

	// addListURLsResponse urls shortening response.
	addListURLsResponse struct {
		// CorrelationID - url id.
		CorrelationID string `json:"correlation_id"`
		// Result - shorten url.
		ShortURL string `json:"short_url"`
	}
)

// NewAppHandler returns new AppHandler.
func NewAppHandler(config *config.AppConfig) *AppHandler {
	h := &AppHandler{
		repo:            config.Repo,
		baseURL:         config.BaseURL + "/",
		dbConn:          config.Conn,
		userIDGenerator: config.UserIDGenerator,
		deleteService:   config.DeleteService,
		wg:              config.RequestWait,
		trustedSubnet:   config.TrustedSubnet,
	}
	h.Router = NewRouter(h)
	return h
}

// NewRouter returns new router.
func NewRouter(appHandler *AppHandler) chi.Router {
	r := chi.NewRouter()
	r.Use(myMiddleware.NewWaitRequest(appHandler.wg).WaitRequest)
	r.Use(myMiddleware.NewAuthService(appHandler.userIDGenerator).Auth)
	r.Use(myMiddleware.Gzip)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Mount("/debug", middleware.Profiler())

	r.Post("/", appHandler.addURL)
	r.Get("/{shortURL}", appHandler.getURL)
	r.Get("/ping", appHandler.ping)

	r.Route("/api", func(r chi.Router) {
		r.Route("/shorten", func(r chi.Router) {
			r.Post("/", appHandler.addURLRest)
			r.Post("/batch", appHandler.addListURLRest)
		})
		r.Route("/user/urls", func(r chi.Router) {
			r.Get("/", appHandler.listURLs)
			r.Delete("/", appHandler.deleteListURLs)
		})
		r.Get("/internal/stats", appHandler.getStats)
	})

	return r
}

// addURLRest handles a request to create a short url in json format.
func (a *AppHandler) addURLRest(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	var body []byte
	var err error
	if body, err = readBody(w, r.Body); err != nil {
		return
	}

	var requestURL arrURLRequest
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

// addListURLRest handles a request to create a short urls in json format.
func (a *AppHandler) addListURLRest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	body, err := readBody(w, r.Body)
	if err != nil {
		return
	}
	userID := r.Context().Value(myMiddleware.UserIDKey).(uint32)
	var urlsForShort []addListURLsRequest
	if err = json.Unmarshal(body, &urlsForShort); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	convertedURLs := make([]repository.URLWithID, len(urlsForShort))
	for i, url := range urlsForShort {
		convertedURLs[i] = repository.URLWithID{
			CorrelationID: url.CorrelationID,
			URL:           url.OriginalURL,
		}
	}

	shortenURLs, err := a.repo.CreateShortURLs(r.Context(), a.baseURL, convertedURLs, userID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	shortenURLsResponse := make([]addListURLsResponse, len(shortenURLs))
	for i, shortenURL := range shortenURLs {
		shortenURLsResponse[i] = addListURLsResponse{
			CorrelationID: shortenURL.CorrelationID,
			ShortURL:      shortenURL.URL,
		}
	}

	resBody, err := json.Marshal(&shortenURLsResponse)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	sendResponse(w, resBody, http.StatusCreated)
}

// addURL handles a request to create a short url in text/plain format.
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

// getURL handles a request to get full url by short url in query param.
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

// listURLs handles a request to get all the shortened urls of a specific user.
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
	sendResponse(w, resp, http.StatusOK)
}

// deleteListURLs handles a request to delete urls owned by a specific user.
func (a *AppHandler) deleteListURLs(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(myMiddleware.UserIDKey).(uint32)
	body, err := readBody(w, r.Body)
	if err != nil {
		return
	}
	a.deleteService.AddURLs(string(body), userID)
	w.WriteHeader(http.StatusAccepted)
}

// ping checks the database connection
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

// addURLRest returns json data from addURLResponse.
func (a *AppHandler) createAddURLResponse(w http.ResponseWriter, shortURL string) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)
	resp := addURLResponse{Result: shortURL}
	if err := encoder.Encode(resp); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, err
	}
	return buf.Bytes(), nil
}

// getStats returns count of users and shortURLs
func (a *AppHandler) getStats(w http.ResponseWriter, r *http.Request) {
	_, ipNetTrusted, err := net.ParseCIDR(a.trustedSubnet)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	realIP := r.Header.Get("X-Real-IP")
	ip := net.ParseIP(realIP)

	if ipNetTrusted.Contains(ip) {
		stats, err := a.repo.GetStats()
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		js, err := json.Marshal(stats)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		sendResponse(w, js, http.StatusOK)
	} else {
		w.WriteHeader(http.StatusForbidden)
		return
	}
}

// sendResponse writes res in w with status and handle errors.
func sendResponse(w http.ResponseWriter, res []byte, status int) {
	w.WriteHeader(status)
	log.Printf("Response: %s\n", res)
	_, err := w.Write(res)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}
}

// readBody reads data from body.
func readBody(w http.ResponseWriter, body io.ReadCloser) ([]byte, error) {
	defer body.Close()
	bodyBytes, err := io.ReadAll(body)
	if err != nil || len(bodyBytes) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return nil, err
	}
	return bodyBytes, nil
}
