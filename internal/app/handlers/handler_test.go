package handlers

import (
	"bytes"
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-axesthump-shortener/internal/app/mocks"
	"go-axesthump-shortener/internal/app/repository"
	"go-axesthump-shortener/internal/app/service"
	"go-axesthump-shortener/internal/app/user"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	shortURL = "http://shortURL/short"
	longURL  = "http://shortURL/looooooooong"
)

type mockStorage struct {
	needError bool
}

func (m *mockStorage) CreateShortURL(ctx context.Context, beginURL string, originalURL string, userID uint32) (string, error) {
	return shortURL, nil
}

func (m *mockStorage) GetFullURL(ctx context.Context, shortURL int64) (string, error) {
	if m.needError {
		return "", errors.New("error")
	} else {
		return longURL, nil
	}
}

func (m *mockStorage) GetAllURLs(ctx context.Context, beginURL string, userID uint32) []repository.URLInfo {
	return make([]repository.URLInfo, 0)
}

func (m *mockStorage) DeleteURLs(urlsForDelete []repository.DeleteURL, ds *service.DeleteService) {
}

func (m *mockStorage) Close() error {
	return nil
}

func (m *mockStorage) CreateShortURLs(
	ctx context.Context,
	beginURL string,
	urls []repository.URLWithID,
	userID uint32,
) ([]repository.URLWithID, error) {
	res := make([]repository.URLWithID, 0, len(urls))
	return res, nil
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
				repo:            tt.fields.storage,
				userIDGenerator: user.NewUserIDGenerator(0),
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
				repo:            tt.fields.storage,
				userIDGenerator: user.NewUserIDGenerator(0),
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

func TestAppHandler_addURLRest(t *testing.T) {
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
				requestURL: "/api/shorten",
				storage: &mockStorage{
					needError: false,
				},
				body: []byte(`{"url":"url"}`),
			},
			want: want{
				statusCode:  http.StatusCreated,
				body:        `{"result":"` + shortURL + `"}`,
				contentType: "application/json",
			},
		},
		{
			name: "check add with bad body",
			fields: fields{
				requestURL: "/api/shorten",
				storage: &mockStorage{
					needError: false,
				},
				body: []byte(`{"url":"url"`),
			},
			want: want{
				statusCode:  http.StatusBadRequest,
				body:        "",
				contentType: "application/json",
			},
		},
		{
			name: "check add with wrong body url type",
			fields: fields{
				requestURL: "/api/shorten",
				storage: &mockStorage{
					needError: false,
				},
				body: []byte(`{"url":1}`),
			},
			want: want{
				statusCode:  http.StatusBadRequest,
				body:        "",
				contentType: "application/json",
			},
		},
		{
			name: "check add with empty body",
			fields: fields{
				requestURL: "/api/shorten",
				storage: &mockStorage{
					needError: false,
				},
			},
			want: want{
				statusCode:  http.StatusBadRequest,
				body:        "",
				contentType: "application/json",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AppHandler{
				repo:            tt.fields.storage,
				userIDGenerator: user.NewUserIDGenerator(0),
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
				assert.Equal(t, tt.want.body, strings.TrimRight(string(body), "\n"))
			}
		})
	}
}

func TestAppHandler_listURLs(t *testing.T) {
	type want struct {
		urls       string
		statusCode int
		needEmpty  bool
	}
	tests := []struct {
		name string
		want want
	}{
		{
			name: "check not empty",
			want: want{
				statusCode: http.StatusOK,
				urls:       `[{"short_url":"short","original_url":"original"}]`,
				needEmpty:  false,
			},
		},
		{
			name: "check empty",
			want: want{
				statusCode: http.StatusNoContent,
				needEmpty:  true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := mocks.NewMockRepository(ctrl)
			defer ctrl.Finish()
			a := &AppHandler{
				repo:            repo,
				userIDGenerator: user.NewUserIDGenerator(0),
			}
			r := NewRouter(a)
			ts := httptest.NewServer(r)
			defer ts.Close()
			a.baseURL = ts.URL

			if tt.want.needEmpty {
				repo.EXPECT().GetAllURLs(gomock.Any(), gomock.Any(), gomock.Any()).Return([]repository.URLInfo{})
			} else {
				repo.EXPECT().GetAllURLs(gomock.Any(), gomock.Any(), gomock.Any()).Return([]repository.URLInfo{
					{
						ShortURL:    "short",
						OriginalURL: "original",
					},
				})
			}

			request, err := http.NewRequest(http.MethodGet, ts.URL+"/api/user/urls", nil)
			require.NoError(t, err)
			res, err := http.DefaultClient.Do(request)
			require.NoError(t, err)

			defer res.Body.Close()
			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			assert.Equal(t, tt.want.urls, strings.TrimRight(string(body), "\n"))
		})
	}
}
