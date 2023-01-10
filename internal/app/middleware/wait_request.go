package middleware

import (
	"net/http"
	"sync"
)

// waitRequest struct for wait clients request
type waitRequest struct {
	wg *sync.WaitGroup
}

// NewWaitRequest returns waitRequest
func NewWaitRequest(wg *sync.WaitGroup) *waitRequest {
	return &waitRequest{
		wg: wg,
	}
}

// WaitRequest middleware for check response done.
func (wr *waitRequest) WaitRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wr.wg.Add(1)
		next.ServeHTTP(w, r)
		wr.wg.Done()
	})
}
