package server

import (
	"log"
	"net/http"
	"crypto/tls"
)

var MyLog = log.Default()

type Server struct {
	http.Server
}

func NewServer(address string, handler http.Handler) *Server {
	return &Server{http.Server{Addr: address, Handler: handler}}
}

func (s *Server) Run() {
	MyLog.Fatal(s.ListenAndServe())
}

func NewServerTLS(address string, handler http.Handler, tlsConfig *tls.Config) *Server {
	return &Server{http.Server{Addr: address, Handler: handler, TLSConfig: tlsConfig}}
}

func (s *Server) RunTLS() {
	MyLog.Fatal(s.ListenAndServeTLS("", ""))
}
