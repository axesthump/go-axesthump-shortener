package main

import (
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type storage struct {
	mx     *sync.RWMutex
	urls   map[int64]string
	lastId int64
}

func (s *storage) handleRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getUrl(w, r)
	case http.MethodPost:
		s.addUrl(w, r)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (s *storage) addUrl(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	url := string(body)

	s.mx.Lock()
	s.urls[s.lastId] = url
	s.mx.Unlock()

	response := strconv.FormatInt(s.lastId, 10)

	s.lastId++

	w.WriteHeader(http.StatusCreated)
	_, err = w.Write([]byte(response))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

}

func (s *storage) getUrl(w http.ResponseWriter, r *http.Request) {
	s.mx.RLock()
	url := strings.Split(r.URL.Path, "/")
	if len(url) != 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	urlInt, err := strconv.ParseInt(url[0], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if longUrl, ok := s.urls[urlInt]; ok {
		w.WriteHeader(http.StatusTemporaryRedirect)
		w.Header().Set("Location", longUrl)
		return
	}

	w.WriteHeader(http.StatusBadRequest)
	return
}

func main() {
	mux := http.NewServeMux()
	storage := storage{
		mx:     &sync.RWMutex{},
		urls:   make(map[int64]string, 0),
		lastId: int64(0),
	}
	mux.HandleFunc("/", storage.handleRequest)
	log.Fatal(http.ListenAndServe(":8080", mux))
}
