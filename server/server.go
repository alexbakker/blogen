package server

import (
	"net/http"
	"path/filepath"
)

type Server struct {
	mux    *http.ServeMux
	config Config
}

func New(config Config, dir string) *Server {
	dir = filepath.Join(dir, "public")

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(dir)))

	return &Server{
		mux:    mux,
		config: config,
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
