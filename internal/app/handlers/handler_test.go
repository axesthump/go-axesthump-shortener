package handlers

import (
	"bytes"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
				statusCode: http.StatusMethodNotAllowed,
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
				statusCode: http.StatusNotFound,
				location:   "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AppHandler{
				repo: tt.fields.storage,
			}
			r := NewRouter(a)
			ts := httptest.NewServer(r)
			defer ts.Close()
			a.baseURL = ts.URL
			request, err := http.NewRequest(http.MethodGet, ts.URL+tt.fields.requestURL, nil)
			require.NoError(t, err)

			transport := http.Transport{}
			res, err := transport.RoundTrip(request)
			require.NoError(t, err)
			if err != nil {
				t.Fatal(err)
			}

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
				statusCode: http.StatusMethodNotAllowed,
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
				repo: tt.fields.storage,
			}
			r := NewRouter(a)
			ts := httptest.NewServer(r)
			defer ts.Close()
			a.baseURL = ts.URL
			request, err := http.NewRequest(http.MethodPost, ts.URL+tt.fields.requestURL, bytes.NewBuffer(tt.fields.body))
			require.NoError(t, err)
			res, err := http.DefaultClient.Do(request)
			require.NoError(t, err)

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
