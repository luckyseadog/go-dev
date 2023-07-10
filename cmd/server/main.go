package main

import (
	"database/sql"
	"errors"
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

	var s storage.Storage

	if envVariables.DataSourceName != "" {
		var err error
		db, err := sql.Open("pgx", envVariables.DataSourceName)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		s = storage.NewSQLStorage(db)
		if ss, ok := s.(*storage.SQLStorage); ok {
			err = ss.CreateTables()
			if err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatal(errors.New("database is not SQLStorage"))
		}
	} else {
		storageChan := make(chan struct{})
		cancel := make(chan struct{})
		defer close(cancel)

		s = storage.NewStorage(storageChan, envVariables.StoreInterval)
		server.PassSignal(cancel, storageChan, envVariables, s)

		if envVariables.Restore {
			if _, err := os.Stat(envVariables.StoreFile); err == nil {
				if ms, ok := s.(*storage.MyStorage); ok {
					err = ms.LoadFromFile(envVariables.StoreFile)
					if err != nil {
						log.Fatal(err)
					}
				} else {
					log.Fatal(errors.New("database is not MyStorage"))
				}
			}
		}
	}

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
		handlers.HandlerPing(w, r, s)
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
	r.Route("/updates", func(r chi.Router) {
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			handlers.HandlerUpdatesJSON(w, r, s, envVariables.SecretKey)
		})
		r.Post("/{_}", func(w http.ResponseWriter, r *http.Request) {
			handlers.HandlerUpdatesJSON(w, r, s, envVariables.SecretKey)
		})
	})

	srv := server.NewServer(envVariables.Address, middlewares.GzipMiddleware(r))
	srv.Run()
	defer srv.Close()
}
