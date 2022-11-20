package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go-axesthump-shortener/internal/app/config"
	"go-axesthump-shortener/internal/app/generator"
	myMiddleware "go-axesthump-shortener/internal/app/middleware"
	"go-axesthump-shortener/internal/app/mocks"
	"go-axesthump-shortener/internal/app/repository"
	"go-axesthump-shortener/internal/app/service"
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

func (m *mockStorage) DeleteURLs(urlsForDelete []repository.DeleteURL) error {
	return nil
}

func (m *mockStorage) CreateShortURL(ctx context.Context, beginURL string, originalURL string, userID uint32) (string, error) {
	if m.needError {
		return "", &repository.LongURLConflictError{}
	}
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
			name: "check bad arrURLRequest",
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
			name: "check bad arrURLRequest",
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
			name: "check bad arrURLRequest",
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
				userIDGenerator: generator.NewIDGenerator(0),
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
			name: "check bad arrURLRequest",
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
		{
			name: "check conflict url body",
			fields: fields{
				requestURL: "/",
				storage: &mockStorage{
					needError: true,
				},
				body: []byte("url"),
			},
			want: want{
				statusCode:  http.StatusConflict,
				body:        "",
				contentType: "text/plain",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AppHandler{
				repo:            tt.fields.storage,
				userIDGenerator: generator.NewIDGenerator(0),
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
		{
			name: "check conflict url body",
			fields: fields{
				requestURL: "/api/shorten",
				storage: &mockStorage{
					needError: true,
				},
				body: []byte(`{"url":"url"}`),
			},
			want: want{
				statusCode:  http.StatusConflict,
				body:        "",
				contentType: "application/json",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AppHandler{
				repo:            tt.fields.storage,
				userIDGenerator: generator.NewIDGenerator(0),
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
				userIDGenerator: generator.NewIDGenerator(0),
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

func BenchmarkAppHandler_addURLRest(b *testing.B) {
	b.Run("Endpoint api/shorten", func(b *testing.B) {
		a := &AppHandler{
			repo:            &mockStorage{},
			userIDGenerator: generator.NewIDGenerator(0),
		}
		r, _ := http.NewRequestWithContext(
			context.WithValue(context.TODO(), myMiddleware.UserIDKey, uint32(1)),
			http.MethodPost,
			"api/shorten",
			bytes.NewBuffer([]byte(`{"url":"1"}`)),
		)
		w := httptest.NewRecorder()
		handler := http.HandlerFunc(a.addURLRest)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			handler.ServeHTTP(w, r)
		}
	})
}

func BenchmarkAppHandler_getURL(b *testing.B) {
	b.Run("Endpoint /1", func(b *testing.B) {
		ctrl := gomock.NewController(b)
		repo := mocks.NewMockRepository(ctrl)
		defer ctrl.Finish()
		a := &AppHandler{
			repo:            repo,
			userIDGenerator: generator.NewIDGenerator(0),
		}
		r, _ := http.NewRequest(
			http.MethodGet,
			"/{shortURL}",
			nil,
		)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("shortURL", "1")
		r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
		w := httptest.NewRecorder()
		handler := http.HandlerFunc(a.getURL)

		repo.EXPECT().GetFullURL(gomock.Any(), int64(1)).Return("fullURL", nil).AnyTimes()

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			handler.ServeHTTP(w, r)
		}
	})
}

func TestAppHandler_ping(t *testing.T) {
	a := &AppHandler{
		dbConn:          nil,
		repo:            &mockStorage{},
		userIDGenerator: generator.NewIDGenerator(0),
	}
	r, _ := http.NewRequest(
		http.MethodGet,
		"/ping",
		nil,
	)
	w := httptest.NewRecorder()
	handler := http.HandlerFunc(a.ping)
	handler.ServeHTTP(w, r)
}

func TestNewAppHandler(t *testing.T) {
	repo := mockStorage{}
	conf := config.AppConfig{
		Repo:            &repo,
		BaseURL:         "baseURL",
		Conn:            nil,
		UserIDGenerator: generator.NewIDGenerator(0),
		DeleteService:   service.NewDeleteService(&repo, "baseURL"),
	}

	appHandler := NewAppHandler(&conf)

	assert.Equal(t, appHandler.baseURL, "baseURL/")
	assert.Nil(t, appHandler.dbConn)
	assert.Equal(t, appHandler.repo, &repo)
	assert.Equal(t, appHandler.userIDGenerator.GetID(), int64(0))
}

func TestAppHandler_addListURLRest(t *testing.T) {
	type fields struct {
		body            []byte
		userIDGenerator *generator.IDGenerator
		baseURL         string
		dbConn          *pgx.Conn
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	type Want struct {
		needError bool
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Want
	}{
		{
			name: "addListURLRest success",
			fields: fields{
				body: []byte(`[
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
				] `),
				userIDGenerator: generator.NewIDGenerator(0),
				baseURL:         "http://localhost:8080/",
				dbConn:          nil,
			},
			want: Want{
				needError: false,
			},
		},
		{
			name: "addListURLRest with empty body",
			fields: fields{
				body:            nil,
				userIDGenerator: generator.NewIDGenerator(0),
				baseURL:         "http://localhost:8080/",
				dbConn:          nil,
			},
			want: Want{
				needError: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			repo := mocks.NewMockRepository(ctrl)
			defer ctrl.Finish()

			r, _ := http.NewRequestWithContext(
				context.WithValue(context.TODO(), myMiddleware.UserIDKey, uint32(1)),
				http.MethodPost,
				"/batch",
				bytes.NewReader(tt.fields.body),
			)
			a := &AppHandler{
				repo:            repo,
				userIDGenerator: tt.fields.userIDGenerator,
				baseURL:         tt.fields.baseURL,
				dbConn:          tt.fields.dbConn,
				deleteService:   service.NewDeleteService(repo, tt.fields.baseURL),
			}

			repo.EXPECT().CreateShortURLs(gomock.Any(), a.baseURL, gomock.Any(), uint32(1)).Return([]repository.URLWithID{
				{
					CorrelationID: "first",
					URL:           "http://localhost:8080/1",
				},
				{
					CorrelationID: "second",
					URL:           "http://localhost:8080/2",
				},
				{
					CorrelationID: "last",
					URL:           "http://localhost:8080/3",
				},
			}, nil).AnyTimes()
			expected := []addListURLsResponse{
				{
					CorrelationID: "first",
					ShortURL:      "http://localhost:8080/1",
				},
				{
					CorrelationID: "second",
					ShortURL:      "http://localhost:8080/2",
				},
				{
					CorrelationID: "last",
					ShortURL:      "http://localhost:8080/3",
				},
			}
			w := httptest.NewRecorder()
			handler := http.HandlerFunc(a.addListURLRest)
			handler.ServeHTTP(w, r)

			if tt.want.needError {
				assert.Equal(t, http.StatusBadRequest, w.Result().StatusCode)
			} else {
				actualBody := make([]addListURLsResponse, 0)
				body, err := io.ReadAll(w.Result().Body)
				assert.NoError(t, err)
				err = json.Unmarshal(body, &actualBody)
				assert.NoError(t, err)
				assert.Equal(t, http.StatusCreated, w.Result().StatusCode)
				assert.Equal(t, 3, len(actualBody))
				assert.Equalf(t, expected, actualBody, "")
			}
			w.Result().Body.Close()
		})
	}
}
