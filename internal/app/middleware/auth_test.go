package middleware

import (
	"github.com/stretchr/testify/assert"
	"go-axesthump-shortener/internal/app/generator"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewAuthService(t *testing.T) {
	gen := generator.NewIDGenerator(0)
	authService := NewAuthService(gen)

	assert.Equal(t, gen, authService.idGenerator)
	assert.Equal(t, []byte("secret_key"), authService.secretKey)
}

func Test_authService_generateCookie(t *testing.T) {
	gen := generator.NewIDGenerator(0)
	authService := NewAuthService(gen)
	writer := httptest.NewRecorder()
	expected := "00000000013fd79b1f129e8734c9c4d34828a3cc4b170e964910a7d662ea3d63ac387a56"
	id := authService.generateCookie(writer)

	assert.Equal(t, id, uint32(0))
	assert.Equal(t, expected, writer.Header().Get("Set-Cookie")[5:])
}

func Test_authService_validateCookie(t *testing.T) {
	type testData struct {
		cookieValue string
		isValid     bool
	}
	tests := []struct {
		name string
		td   testData
	}{
		{
			name: "Test authService with valid cookie",
			td: testData{
				cookieValue: "00000000013fd79b1f129e8734c9c4d34828a3cc4b170e964910a7d662ea3d63ac387a56",
				isValid:     true,
			},
		},

		{
			name: "Test authService with invalid cookie",
			td: testData{
				cookieValue: "12300000013fd79b1f129e8734c9c4d34828a3cc4b170e964910a7d662ea3d63ac387a56",
				isValid:     false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := generator.NewIDGenerator(1)
			authService := NewAuthService(gen)
			cookie := &http.Cookie{
				Name:  "auth",
				Value: tt.td.cookieValue,
			}

			actualValid, actualID := authService.validateCookie(cookie)
			assert.Equal(t, uint32(0), actualID)
			assert.Equal(t, actualValid, tt.td.isValid)
		})
	}
}

func Test_authService_Auth(t *testing.T) {
	type testData struct {
		cookie     *http.Cookie
		expectedID uint32
	}
	tests := []struct {
		name string
		td   testData
	}{
		{
			name: "Test authService with valid cookie",
			td: testData{
				cookie: &http.Cookie{
					Name:  "auth",
					Value: "00000000013fd79b1f129e8734c9c4d34828a3cc4b170e964910a7d662ea3d63ac387a56",
				},
				expectedID: 0,
			},
		},
		{
			name: "Test authService with invalid cookie",
			td: testData{
				cookie: &http.Cookie{
					Name:  "auth",
					Value: "invalid",
				},
				expectedID: 1,
			},
		},
		{
			name: "Test authService without cookie",
			td: testData{
				cookie:     nil,
				expectedID: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := generator.NewIDGenerator(1)
			authService := NewAuthService(gen)
			req := httptest.NewRequest(http.MethodGet, "http://testing", nil)
			if tt.td.cookie != nil {
				req.AddCookie(tt.td.cookie)
			}
			httpHandler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				id := request.Context().Value(UserIDKey)
				assert.Equal(t, tt.td.expectedID, id)
			})
			newHandler := authService.Auth(httpHandler)

			writer := httptest.NewRecorder()
			newHandler.ServeHTTP(writer, req)
		})
	}
}
