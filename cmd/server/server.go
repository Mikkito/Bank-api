package server

import (
	"net/http"

	"github.com/gorilla/mux"
)

// Server содержит роутер и ссылки на сервисы/репозитории, если нужно
type Server struct {
	Router *mux.Router
}

// NewServer создает новый экземпляр сервера с роутером
func NewServer() *Server {
	router := mux.NewRouter()
	return &Server{
		Router: router,
	}
}

// Start server
func (s *Server) ListenAndServe() error {
	return http.ListenAndServe(":8080", s.Router)
}
