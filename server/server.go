package server

import (
	"net/http"
)

type Server struct {
	mux    *http.ServeMux
	config Config
}

func New(config Config, dir string) *Server {
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
