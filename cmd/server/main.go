package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/luckyseadog/go-dev/internal/handlers"
	"github.com/luckyseadog/go-dev/internal/server"
	"github.com/luckyseadog/go-dev/internal/storage"
)

func main() {
	s := storage.NewStorage()
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandlerDefault(w, r, s)
	})
	r.Get("/value/*", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandlerGet(w, r, s)
	})
	r.Post("/update/*", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandlerUpdate(w, r, s)
	})

	server := server.NewServer("127.0.0.1:8080", r)
	server.Run()
	defer server.Close()
}
