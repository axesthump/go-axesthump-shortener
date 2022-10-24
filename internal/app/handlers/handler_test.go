package handlers

import (
	"bytes"
	"errors"
	"github.com/stretchr/testify/assert"
	"go-axesthump-shortener/internal/app/repository"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	shortURL = "http://shortURL/short"
	longURL  = "http://shortURL/looooooooong"
)

type mockStorage struct {
	needError bool
}

func (m *mockStorage) CreateShortURL(beginURL string, url string) string {
	return shortURL
}

func (m *mockStorage) GetFullURL(shortURL int64) (string, error) {
	if m.needError {
		return "", errors.New("error")
	} else {
		return longURL, nil
	}
}

func TestAppHandler_HandleRequest(t *testing.T) {
	type fields struct {
		storage    repository.Repository
		requestURL string
		method     string
		body       []byte
	}
	type want struct {
		statusCode int
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "check with wrong url request",
			fields: fields{
				requestURL: "/some/wrong/url",
				method:     http.MethodGet,
				storage: &mockStorage{
					needError: false,
				},
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "check with put request",
			fields: fields{
				requestURL: "/l",
				method:     http.MethodPut,
				storage: &mockStorage{
					needError: false,
				},
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "check with delete request",
			fields: fields{
				requestURL: "/l",
				method:     http.MethodPut,
				storage: &mockStorage{
					needError: false,
				},
			},
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name: "check with get url request",
			fields: fields{
				requestURL: "/0",
				method:     http.MethodGet,
				storage: &mockStorage{
					needError: false,
				},
			},
			want: want{
				statusCode: http.StatusTemporaryRedirect,
			},
		},
		{
			name: "check with post url request",
			fields: fields{
				requestURL: "/",
				method:     http.MethodPost,
				storage: &mockStorage{
					needError: false,
				},
				body: []byte("url"),
			},
			want: want{
				statusCode: http.StatusCreated,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AppHandler{
				storage: tt.fields.storage,
			}

			request := httptest.NewRequest(tt.fields.method, tt.fields.requestURL, bytes.NewBuffer(tt.fields.body))
			handler := http.HandlerFunc(a.HandleRequest)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, request)

			res := w.Result()
			defer res.Body.Close()
			assert.Equal(t, tt.want.statusCode, res.StatusCode)
		})
	}
}

func TestAppHandler_getURL(t *testing.T) {
	type fields struct {
		storage    repository.Repository
		requestURL string
	}
	type want struct {
		statusCode int
		location   string
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "check success add",
			fields: fields{
				requestURL: "/0",
				storage: &mockStorage{
					needError: false,
				},
			},
			want: want{
				statusCode: http.StatusTemporaryRedirect,
				location:   longURL,
			},
		},
		{
			name: "check shortURL dont exist",
			fields: fields{
				requestURL: "/1",
				storage: &mockStorage{
					needError: true,
				},
			},
			want: want{
				statusCode: http.StatusBadRequest,
				location:   "",
			},
		},
		{
			name: "check bad requestURL",
			fields: fields{
				requestURL: "/some",
				storage: &mockStorage{
					needError: false,
				},
			},
			want: want{
				statusCode: http.StatusBadRequest,
				location:   "",
			},
		},
		{
			name: "check bad requestURL",
			fields: fields{
				requestURL: "/",
				storage: &mockStorage{
					needError: false,
				},
			},
			want: want{
				statusCode: http.StatusBadRequest,
				location:   "",
			},
		},
		{
			name: "check bad requestURL",
			fields: fields{
				requestURL: "/0/0",
				storage: &mockStorage{
					needError: false,
				},
			},
			want: want{
				statusCode: http.StatusBadRequest,
				location:   "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AppHandler{
				storage: tt.fields.storage,
			}

			request := httptest.NewRequest(http.MethodGet, tt.fields.requestURL, nil)
			handler := http.HandlerFunc(a.HandleRequest)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, request)

			res := w.Result()
			defer res.Body.Close()
			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			if len(tt.want.location) != 0 {
				assert.Equal(t, tt.want.location, res.Header.Get("Location"))
			}
		})
	}
}

func TestAppHandler_addURL(t *testing.T) {
	type fields struct {
		storage    repository.Repository
		requestURL string
		body       []byte
	}
	type want struct {
		statusCode  int
		body        string
		contentType string
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "check success add",
			fields: fields{
				requestURL: "/",
				storage: &mockStorage{
					needError: false,
				},
				body: []byte("url"),
			},
			want: want{
				statusCode:  http.StatusCreated,
				body:        shortURL,
				contentType: "text/plain",
			},
		},
		{
			name: "check bad requestURL",
			fields: fields{
				requestURL: "/0",
				storage: &mockStorage{
					needError: false,
				},
				body: []byte("url"),
			},
			want: want{
				statusCode:  http.StatusBadRequest,
				body:        "",
				contentType: "text/plain",
			},
		},
		{
			name: "check empty body",
			fields: fields{
				requestURL: "/",
				storage: &mockStorage{
					needError: false,
				},
			},
			want: want{
				statusCode:  http.StatusBadRequest,
				body:        "",
				contentType: "text/plain",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AppHandler{
				storage: tt.fields.storage,
			}

			request := httptest.NewRequest(http.MethodPost, tt.fields.requestURL, bytes.NewBuffer(tt.fields.body))
			handler := http.HandlerFunc(a.HandleRequest)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, request)

			res := w.Result()
			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
			if len(tt.want.body) != 0 {
				defer res.Body.Close()
				body, err := io.ReadAll(res.Body)
				if err != nil {
					t.Fatal(err)
				}
				assert.Equal(t, tt.want.body, string(body))
			}
		})
	}
}
