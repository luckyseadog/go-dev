package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/luckyseadog/go-dev/internal/handlers"
	"github.com/luckyseadog/go-dev/internal/middlewares"
	"github.com/luckyseadog/go-dev/internal/server"
	"github.com/luckyseadog/go-dev/internal/storage"
)

func main() {
	envVariables, err := server.SetUp()
	if err != nil {
		server.MyLog.Fatal(err)
	}

	if envVariables.IsLog {
		flog, err := os.OpenFile(`server.log`, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
		if err == nil {
			server.MyLog = log.New(flog, `server `, log.LstdFlags|log.Lshortfile)
			defer flog.Close()
		} else {
			server.MyLog.Fatal("error in creating file for logging")
		}
	}

	var s storage.Storage

	if envVariables.DataSourceName != "" {
		var err error
		db, err := sql.Open("pgx", envVariables.DataSourceName)
		if err != nil {
			server.MyLog.Fatal(err)
		}
		defer db.Close()

		s = storage.NewSQLStorage(db)
		if ss, ok := s.(*storage.SQLStorage); ok {
			err = ss.CreateTables()
			if err != nil {
				server.MyLog.Fatal(err)
			}
		} else {
			server.MyLog.Fatal(storage.ErrNotSQLStorage)
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
						server.MyLog.Fatal(err)
					}
				} else {
					server.MyLog.Fatal(storage.ErrNotMyStorage)
				}
			}
		}
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middlewares.GzipMiddleware)

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

	srv := server.NewServer(envVariables.Address, r)
	defer srv.Close()
	srv.Run()
}
