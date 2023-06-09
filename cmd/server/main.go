package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/luckyseadog/go-dev/internal/handlers"
	"github.com/luckyseadog/go-dev/internal/server"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", handlers.HandlerDefault)
	r.Get("/value/{^+}/*", handlers.HandlerGet)
	r.Post("/value/*", handlers.HandlerValueJSON)
	r.Post("/update/{^+}/*", handlers.HandlerUpdate)
	r.Post("/update/*", handlers.HandlerUpdateJSON)

	server := server.NewServer("127.0.0.1:8080", r)
	server.Run()
	defer server.Close()
}
