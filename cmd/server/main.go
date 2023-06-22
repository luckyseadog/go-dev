package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/luckyseadog/go-dev/internal/handlers"
	"github.com/luckyseadog/go-dev/internal/server"
	"github.com/luckyseadog/go-dev/internal/storage"
)

func main() {
	s := storage.NewStorage()

	envVariables := server.SetUp()

	dir := filepath.Dir(envVariables.StoreFile)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.Mkdir(dir, 0777)
		if err != nil {
			log.Fatal(err)
		}
	}

	err := s.LoadMetricsTypes(filepath.Join(dir, "metric_types.json"))
	if err != nil {
		log.Println(err)
	}

	if envVariables.Restore {
		if _, err := os.Stat(envVariables.StoreFile); err == nil {
			err := s.LoadFromFile(envVariables.StoreFile)
			if err != nil {
				log.Println(err)
			}
		}
	}

	fileSaveChan := make(chan time.Time)
	cancel := make(chan struct{})
	defer close(cancel)

	server.PassSignal(cancel, fileSaveChan, envVariables, s)
	s.SaveMetricsTypes(cancel, filepath.Join(dir, "metric_types.json"))

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
			select {
			case <-fileSaveChan:
				err := s.SaveToFile(envVariables.StoreFile)
				if err != nil {
					log.Println(err)
				}
			default:
			}
		})
		r.Post("/{_}", func(w http.ResponseWriter, r *http.Request) {
			handlers.HandlerUpdateJSON(w, r, s)
			select {
			case <-fileSaveChan:
				err := s.SaveToFile(envVariables.StoreFile)
				if err != nil {
					log.Println(err)
				}
			default:
			}
		})
	})

	srv := server.NewServer(envVariables.Address, r)
	srv.Run()
	defer srv.Close()
}
