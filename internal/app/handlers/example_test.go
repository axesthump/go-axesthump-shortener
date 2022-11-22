package handlers

import (
	"bytes"
	"context"
	"go-axesthump-shortener/internal/app/config"
	myMiddleware "go-axesthump-shortener/internal/app/middleware"
	"log"
	"net/http"
	"net/http/httptest"
)

func ExampleAddURLRest() {
	conf, err := config.NewAppConfig()
	if err != nil {
		log.Fatal(err)
	}
	appHandler := NewAppHandler(conf)

	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(
		context.WithValue(context.TODO(), myMiddleware.UserIDKey, uint32(1)),
		http.MethodPost,
		"/api/shorten",
		bytes.NewReader([]byte(`{"url":"url"`)),
	)
	handler := http.HandlerFunc(appHandler.addURLRest)
	handler.ServeHTTP(w, r)
}

func ExampleAddURL() {
	conf, err := config.NewAppConfig()
	if err != nil {
		log.Fatal(err)
	}
	appHandler := NewAppHandler(conf)

	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(
		context.WithValue(context.TODO(), myMiddleware.UserIDKey, uint32(1)),
		http.MethodPost,
		"/",
		bytes.NewReader([]byte("url")),
	)
	handler := http.HandlerFunc(appHandler.addURLRest)
	handler.ServeHTTP(w, r)
}

func ExampleAddListURLRest() {
	conf, err := config.NewAppConfig()
	if err != nil {
		log.Fatal(err)
	}
	appHandler := NewAppHandler(conf)

	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(
		context.WithValue(context.TODO(), myMiddleware.UserIDKey, uint32(1)),
		http.MethodPost,
		"/api/batch",
		bytes.NewReader([]byte(`[
					{
						"correlation_id": "first",
						"original_url": "<URL для сокращени>"
					},
					{
						"correlation_id": "second",
						"original_url": "<URL для сокраения>"
					},
					{
						"correlation_id": "last",
						"original_url": "<URL дя сокращения>"
					}
				] `)))
	handler := http.HandlerFunc(appHandler.addURLRest)
	handler.ServeHTTP(w, r)
}

func ExampleGetURL() {
	conf, err := config.NewAppConfig()
	if err != nil {
		log.Fatal(err)
	}
	appHandler := NewAppHandler(conf)

	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(
		context.WithValue(context.TODO(), myMiddleware.UserIDKey, uint32(1)),
		http.MethodGet,
		"/1",
		nil,
	)
	handler := http.HandlerFunc(appHandler.addURLRest)
	handler.ServeHTTP(w, r)
}

func ExampleListURLs() {
	conf, err := config.NewAppConfig()
	if err != nil {
		log.Fatal(err)
	}
	appHandler := NewAppHandler(conf)

	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(
		context.WithValue(context.TODO(), myMiddleware.UserIDKey, uint32(1)),
		http.MethodGet,
		"/user/urls",
		nil,
	)
	handler := http.HandlerFunc(appHandler.addURLRest)
	handler.ServeHTTP(w, r)
}

func ExampleDeleteListURLs() {
	conf, err := config.NewAppConfig()
	if err != nil {
		log.Fatal(err)
	}
	appHandler := NewAppHandler(conf)

	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(
		context.WithValue(context.TODO(), myMiddleware.UserIDKey, uint32(1)),
		http.MethodDelete,
		"/user/urls",
		bytes.NewReader([]byte(`["http://localhost:8080/some/url", "http://localhost:8080/another/url"]`)),
	)
	handler := http.HandlerFunc(appHandler.addURLRest)
	handler.ServeHTTP(w, r)
}

func ExamplePing() {
	conf, err := config.NewAppConfig()
	if err != nil {
		log.Fatal(err)
	}
	appHandler := NewAppHandler(conf)

	w := httptest.NewRecorder()
	r, _ := http.NewRequestWithContext(
		context.WithValue(context.TODO(), myMiddleware.UserIDKey, uint32(1)),
		http.MethodDelete,
		"/ping",
		nil,
	)
	handler := http.HandlerFunc(appHandler.addURLRest)
	handler.ServeHTTP(w, r)
}
