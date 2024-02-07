package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var MyLog = log.Default()

type ServerInterface interface {
	Run()
	RunTLS()
}

type Server struct {
	http.Server
}

func NewServer(address string, handler http.Handler) *Server {
	return &Server{http.Server{Addr: address, Handler: handler}}
}

func (s *Server) Run() {
	serveChan := make(chan error, 1)
	go func() {
		serveChan <- s.ListenAndServe()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	select {
	case <-stop:
		fmt.Println("shutting down gracefully")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.Shutdown(ctx); err != nil {
			MyLog.Println(err)
		}

	case err := <-serveChan:
		MyLog.Fatal(err)
	}
}

func NewServerTLS(address string, handler http.Handler, tlsConfig *tls.Config) *Server {
	return &Server{http.Server{Addr: address, Handler: handler, TLSConfig: tlsConfig}}
}

func (s *Server) RunTLS() {
	serveChan := make(chan error, 1)
	go func() {
		serveChan <- s.ListenAndServeTLS("", "")
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	select {
	case <-stop:
		fmt.Println("shutting down gracefully")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.Shutdown(ctx); err != nil {
			MyLog.Println(err)
		}

	case err := <-serveChan:
		MyLog.Fatal(err)
	}
}
