package api

import (
	"github.com/gorilla/mux"
)

// Router returns the server's router for testing
func (s *Server) Router() *mux.Router {
	return s.router
}