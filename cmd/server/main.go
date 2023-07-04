package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/luckyseadog/go-dev/internal/handlers"
	"github.com/luckyseadog/go-dev/internal/middlewares"
	"github.com/luckyseadog/go-dev/internal/server"
	"github.com/luckyseadog/go-dev/internal/storage"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	envVariables := server.SetUp()

	storageChan := make(chan struct{})
	s := storage.NewStorage(storageChan, envVariables.StoreInterval)

	var db *sql.DB
	if envVariables.DataSourceName != "" {
		var err error
		db, err = sql.Open("pgx", envVariables.DataSourceName)
		if err != nil {
			log.Fatal(err)
		}
		err = storage.CreateTables(db)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
	}

	if envVariables.Restore {
		if envVariables.DataSourceName != "" {
			err := s.LoadFromDB(db)
			if err != nil {
				log.Println(err)
			}
		} else if _, err := os.Stat(envVariables.StoreFile); err == nil {
			err := s.LoadFromFile(envVariables.StoreFile)
			if err != nil {
				log.Println(err)
			}
		}
	}

	cancel := make(chan struct{})
	defer close(cancel)

	server.PassSignal(cancel, storageChan, envVariables, s, db)

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
		})
		r.Post("/{_}", func(w http.ResponseWriter, r *http.Request) {
			handlers.HandlerUpdateJSON(w, r, s, envVariables.SecretKey)
		})
	})

	srv := server.NewServer(envVariables.Address, middlewares.GzipMiddleware(r))
	srv.Run()
	defer srv.Close()
}
