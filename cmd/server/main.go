package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/luckyseadog/go-dev/pkg"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", pkg.HandlerDefault)
	r.Get("/update/", pkg.HandlerUpdate)

	server := NewServer("127.0.0.1:8080", r)
	server.Run()
	defer server.Close()
}
