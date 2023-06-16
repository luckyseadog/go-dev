package main

import (
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/luckyseadog/go-dev/internal/handlers"
	"github.com/luckyseadog/go-dev/internal/server"
	"github.com/luckyseadog/go-dev/internal/storage"
)

func main() {
	address := os.Getenv("ADDRESS")
	if address == "" {
		address = "127.0.0.1:8080"
	}

	s := storage.NewStorage()
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandlerDefault(w, r, s)
	})
	r.Get("/value/{^+}/*", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandlerGet(w, r, s)
	})
	r.Route("/value", func(r chi.Router) {
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			handlers.HandlerValueJSON(w, r, s)
		})
		r.Post("/{_}", func(w http.ResponseWriter, r *http.Request) {
			handlers.HandlerValueJSON(w, r, s)
		})
	})

	r.Post("/update/{^+}/*", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandlerUpdate(w, r, s)
	})
	r.Route("/update", func(r chi.Router) {
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			handlers.HandlerUpdateJSON(w, r, s)
		})
		r.Post("/{_}", func(w http.ResponseWriter, r *http.Request) {
			handlers.HandlerUpdateJSON(w, r, s)
		})
	})

	server := server.NewServer(address, r)
	server.Run()
	defer server.Close()
}
