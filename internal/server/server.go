package server

import (
	"log"
	"net/http"
)

type Server struct {
	http.Server
}

func NewServer(address string, handler http.Handler) *Server {
	return &Server{http.Server{Addr: address, Handler: handler}}
}

func (s *Server) Run() {
	log.Fatal(s.ListenAndServe())
}
