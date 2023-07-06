package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/luckyseadog/go-dev/internal/handlers"
	"github.com/luckyseadog/go-dev/internal/middlewares"
	"github.com/luckyseadog/go-dev/internal/server"
	"github.com/luckyseadog/go-dev/internal/storage"
	"net/http"
)

func main() {
	storageChan := make(chan struct{})
	s := storage.NewStorage(storageChan)
	envVariables := server.SetUp(s)
	s.SetUp(envVariables.StoreInterval)

	cancel := make(chan struct{})
	defer close(cancel)

	server.PassSignal(cancel, storageChan, envVariables, s)

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

	srv := server.NewServer(envVariables.Address, middlewares.GzipMiddleware(r))
	srv.Run()
	defer srv.Close()
}
