package api

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type Server struct {
	router    chi.Router
	serverURL string
}

func NewServer(serverURL string, router chi.Router) *Server {
	return &Server{
		router:    router,
		serverURL: serverURL,
	}
}

func (s *Server) Start() error {
	return http.ListenAndServe(fmt.Sprintf("%s", s.serverURL), s.router)
}
