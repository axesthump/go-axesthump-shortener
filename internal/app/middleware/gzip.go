package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// gzipWriter needed for gzip response.
type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (g gzipWriter) Write(data []byte) (int, error) {
	return g.Writer.Write(data)
}

// Gzip middleware for gzip.
func Gzip(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := setUnpackBody(r); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}
		gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer gz.Close()

		w.Header().Set("Content-Encoding", "gzip")
		next.ServeHTTP(gzipWriter{ResponseWriter: w, Writer: gz}, r)
	})
}

// setUnpackBody if request contains gzip then unpack body and insert it in request.
func setUnpackBody(r *http.Request) error {
	if r.Header.Get("Content-Encoding") != "gzip" {
		return nil
	}

	var data bytes.Buffer
	reader, err := gzip.NewReader(r.Body)
	if err != nil {
		return err
	}

	if _, err = data.ReadFrom(reader); err != nil {
		return err
	}
	reader.Close()
	r.Body.Close()
	newBody := io.NopCloser(&data)
	r.Body = newBody
	return nil
}
