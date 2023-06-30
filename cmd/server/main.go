package main

import (
	"database/sql"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/luckyseadog/go-dev/internal/handlers"
	"github.com/luckyseadog/go-dev/internal/middlewares"
	"github.com/luckyseadog/go-dev/internal/server"
	"github.com/luckyseadog/go-dev/internal/storage"
	"log"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	storageChan := make(chan struct{})
	s := storage.NewStorage(storageChan)
	envVariables := server.SetUp(s)
	s.SetUp(envVariables.StoreInterval)

	db, err := sql.Open("pgx", envVariables.DataSourceName)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

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
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandlerPing(w, r, db)
	})
	r.Route("/value", func(r chi.Router) {
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			handlers.HandlerValueJSON(w, r, s, envVariables.SecretKey)
		})
		r.Post("/{_}", func(w http.ResponseWriter, r *http.Request) {
			handlers.HandlerValueJSON(w, r, s, envVariables.SecretKey)
		})
	})

	r.Post("/update/{^+}/*", func(w http.ResponseWriter, r *http.Request) {
		handlers.HandlerUpdate(w, r, s)
	})
	r.Route("/update", func(r chi.Router) {
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			handlers.HandlerUpdateJSON(w, r, s, envVariables.SecretKey)
			//if envVariables.StoreInterval == 0 {
			//	go server.SyncUpdate(envVariables, s)
			//}
		})
		r.Post("/{_}", func(w http.ResponseWriter, r *http.Request) {
			handlers.HandlerUpdateJSON(w, r, s, envVariables.SecretKey)
			//if envVariables.StoreInterval == 0 {
			//	go server.SyncUpdate(envVariables, s)
			//}
		})
	})

	srv := server.NewServer(envVariables.Address, middlewares.GzipMiddleware(r))
	defer srv.Close()
	srv.Run()
}
