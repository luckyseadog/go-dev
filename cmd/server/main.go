// This file contains the main function for the server application.
// The server stores the metrics received from the agent.
package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/luckyseadog/go-dev/internal/handlers"
	"github.com/luckyseadog/go-dev/internal/middlewares"
	"github.com/luckyseadog/go-dev/internal/server"
	"github.com/luckyseadog/go-dev/internal/storage"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func init() {
	fmt.Fprintf(os.Stdout, "Build version: %s\n", buildVersion)
	fmt.Fprintf(os.Stdout, "Build date: %s\n", buildDate)
	fmt.Fprintf(os.Stdout, "Build commit: %s\n", buildCommit)
}

func main() {
	// SetUp initializes environment variables for the application based on command-line flags and environment variables.
	// It returns an EnvVariables struct with the configured values.
	// If any configuration error occurs, it returns an error.
	envVariables, err := server.SetUp()
	if err != nil {
		server.MyLog.Fatal(err)
	}

	// If logging is enabled in the environment variables configuration:
	// - Open or create the "server.log" file for writing logs.
	// - Set up the server's logger to write to the log file with timestamp and file information.
	if envVariables.Logging {
		flog, err := os.OpenFile(`server.log`, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
		if err == nil {
			server.MyLog = log.New(flog, `server `, log.LstdFlags|log.Lshortfile)
			defer flog.Close()
		} else {
			server.MyLog.Fatal("error in creating file for logging")
		}
	}

	// The initialization of Storage
	var s storage.Storage

	// If envVariables.DataSourceName is set, SQLStorage will be used, MyStorage otherwise.
	if envVariables.DataSourceName != "" {
		// Open a database connection based on the provided data source name.
		var err error
		db, err := sql.Open("pgx", envVariables.DataSourceName)
		if err != nil {
			server.MyLog.Fatal(err)
		}
		defer db.Close()

		// Create a new storage of type SQLStorage that will be used for storing metrics.
		s = storage.NewSQLStorage(db)

		// Perform table creation for SQL storage, if applicable.
		if ss, ok := s.(*storage.SQLStorage); ok {
			err = ss.CreateTables()
			if err != nil {
				server.MyLog.Fatal(err)
			}
		} else {
			server.MyLog.Fatal(storage.ErrNotSQLStorage)
		}
	} else {
		// Create a new storage of type MyStorage that will be used for storing metrics.
		// MyStorage uses chanel for storing metrics to file.
		storageChan := make(chan struct{})
		cancel := make(chan struct{})
		defer close(cancel)

		s = storage.NewStorage(storageChan, envVariables.StoreInterval)

		// Start a goroutine that saves metrics from MyStorage to file.
		server.PassSignal(cancel, storageChan, envVariables, s)

		// If previous metrics should be restored from a file, the program will use LoadFromFile function.
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

	// Create a new Chi router instance.
	r := chi.NewRouter()

	// Attach middleware to the router.
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middlewares.GzipMiddleware)

	// Define routes and handlers for various endpoints.
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
	r.Route("/debug/pprof", func(r chi.Router) {
		r.HandleFunc("/*", func(w http.ResponseWriter, r *http.Request) {
			pprof.Index(w, r)
		})
		r.HandleFunc("/profile", func(w http.ResponseWriter, r *http.Request) {
			pprof.Profile(w, r)
		})
		r.HandleFunc("/cmdline", func(w http.ResponseWriter, r *http.Request) {
			pprof.Cmdline(w, r)
		})
		r.HandleFunc("/symbol", func(w http.ResponseWriter, r *http.Request) {
			pprof.Symbol(w, r)
		})
		r.HandleFunc("/trace", func(w http.ResponseWriter, r *http.Request) {
			pprof.Trace(w, r)
		})
	})

	fmt.Println(envVariables.Address)
	// Create a new server instance with the provided address and router.
	srv := server.NewServer(envVariables.Address, r)
	defer srv.Close()

	// Start the server's operation.
	srv.Run()
}
