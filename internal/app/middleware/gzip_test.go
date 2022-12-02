package middleware

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_setUnpackBody(t *testing.T) {
	type testData struct {
		body          io.Reader
		expectedErr   bool
		addHeader     bool
		createZipBody bool
	}
	tests := []struct {
		name string
		td   testData
	}{
		{
			name: "Test without Content-Encoding header",
			td: testData{
				body:          bytes.NewReader([]byte("body")),
				expectedErr:   false,
				addHeader:     false,
				createZipBody: false,
			},
		},
		{
			name: "Test with Content-Encoding header",
			td: testData{
				body:          bytes.NewReader([]byte("body")),
				expectedErr:   false,
				addHeader:     true,
				createZipBody: true,
			},
		},
		{
			name: "Test with Content-Encoding header and bad body",
			td: testData{
				body:          bytes.NewReader([]byte("body")),
				expectedErr:   true,
				addHeader:     true,
				createZipBody: false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gz *gzip.Writer
			if tt.td.createZipBody {
				var newBody bytes.Buffer
				var err error
				gz, err = gzip.NewWriterLevel(&newBody, gzip.BestSpeed)
				assert.NoError(t, err)
				_, err = gz.Write([]byte("Hello world!"))
				assert.NoError(t, err)
				err = gz.Close()
				assert.NoError(t, err)
				tt.td.body = &newBody
			}
			request := httptest.NewRequest(http.MethodGet, "http://localhost:8080/test", tt.td.body)
			if tt.td.addHeader {
				request.Header.Set("Content-Encoding", "gzip")
			}
			err := setUnpackBody(request)
			if tt.td.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}

}

func TestGzip(t *testing.T) {
	type testData struct {
		body                []byte
		wantZipResponseBody bool
		createZipBody       bool
		wantErr             bool
		addingHeader        bool
	}
	tests := []struct {
		name string
		td   testData
	}{
		{
			name: "Test gzip with not zip body and without zip response",
			td: testData{
				body:                []byte("Hello world!!!"),
				wantZipResponseBody: false,
				createZipBody:       false,
				wantErr:             false,
				addingHeader:        false,
			},
		},
		{
			name: "Test gzip with zip header and bad body",
			td: testData{
				body:                []byte("Hello world!!!"),
				wantZipResponseBody: false,
				createZipBody:       false,
				wantErr:             true,
				addingHeader:        true,
			},
		},
		{
			name: "Test gzip with zip header and no zip body and want zip answer",
			td: testData{
				body:                []byte("Hello world!!!"),
				createZipBody:       true,
				wantZipResponseBody: true,
				wantErr:             false,
				addingHeader:        true,
			},
		},
		{
			name: "Test gzip with zip header and no zip body and want regular answer",
			td: testData{
				body:                []byte("Hello world!!!"),
				createZipBody:       true,
				wantZipResponseBody: false,
				wantErr:             false,
				addingHeader:        true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var newBody bytes.Buffer
			if tt.td.createZipBody {
				var gz *gzip.Writer
				var err error
				gz, err = gzip.NewWriterLevel(&newBody, gzip.BestSpeed)
				assert.NoError(t, err)
				_, err = gz.Write(tt.td.body)
				assert.NoError(t, err)
				err = gz.Close()
				assert.NoError(t, err)
			}
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				data, err := io.ReadAll(r.Body)
				assert.NoError(t, err)
				assert.Equal(t, tt.td.body, data)
				fmt.Fprintf(w, "%s", data)
			})

			zipHandler := Gzip(handler)
			var req *http.Request
			if tt.td.createZipBody {
				req = httptest.NewRequest(http.MethodGet, "http://localhost:8080", &newBody)
			} else {
				req = httptest.NewRequest(http.MethodGet, "http://localhost:8080", bytes.NewBuffer(tt.td.body))
			}
			if tt.td.addingHeader {
				req.Header.Set("Content-Encoding", "gzip")
			}
			if tt.td.wantZipResponseBody {
				req.Header.Set("Accept-Encoding", "gzip")
			}
			res := httptest.NewRecorder()
			zipHandler.ServeHTTP(res, req)

			switch {
			case tt.td.wantErr:
				assert.Equal(t, http.StatusBadRequest, res.Code)
			case tt.td.wantZipResponseBody:
				var buff bytes.Buffer
				gz, err := gzip.NewReader(res.Body)
				assert.NoError(t, err)
				_, err = buff.ReadFrom(gz)
				assert.NoError(t, err)
				assert.Equal(t, string(tt.td.body), buff.String())
			default:
				data, err := io.ReadAll(res.Body)
				assert.NoError(t, err)
				assert.Equal(t, tt.td.body, data)
			}
		})
	}

}
