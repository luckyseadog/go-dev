// This file contains the main function for the server application.
// The server stores the metrics received from the agent.
package main

import (
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"path"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/luckyseadog/go-dev/internal/handlers"
	"github.com/luckyseadog/go-dev/internal/middlewares"
	"github.com/luckyseadog/go-dev/internal/server"
	"github.com/luckyseadog/go-dev/internal/storage"

	pb "github.com/luckyseadog/go-dev/protobuf"
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
		fmt.Println("HERE")
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

	// Create a new server instance with the provided address and router.
	var srv server.ServerInterface
	if envVariables.GRPC {
		fmt.Println("BEGIN with ~186 line and implement gRPC over HTTP and HTTPS; storage inside gRPC")
		// srv := server.NewServerGRPC(envVariables.Address)
		// pb.RegisterMetricsCollectServer(srv, &server.MetricsCollectServer{Storage: s})

		// if envVariables.CryptoKeyDir != "" {
		// grpcServer := grpc.NewServer(
		// 	grpc.Creds(credentials.NewTLS(tlsConfig)),
		// )
		if envVariables.CryptoKeyDir != "" {
			serverTLSCert, err := tls.LoadX509KeyPair(
				path.Join(envVariables.CryptoKeyDir, "server/certServer.pem"),
				path.Join(envVariables.CryptoKeyDir, "server/privateKeyServer.pem"),
			)

			if err != nil {
				server.MyLog.Fatalf("Error loading certificate and key file: %v", err)
			}

			// Configure the server to trust TLS client cert issued by your CA.
			certPool := x509.NewCertPool()
			if caCertPEM, err := os.ReadFile(path.Join(envVariables.CryptoKeyDir, "root/certRoot.pem")); err != nil {
				server.MyLog.Fatal(err)
			} else if ok := certPool.AppendCertsFromPEM(caCertPEM); !ok {
				server.MyLog.Fatal("invalid cert in CA PEM")
			}
			tlsConfig := &tls.Config{
				ClientAuth:   tls.RequireAndVerifyClientCert,
				ClientCAs:    certPool,
				Certificates: []tls.Certificate{serverTLSCert},
			}

			srv := server.NewServerGRPC(envVariables.Address, tlsConfig)
			pb.RegisterMetricsCollectServer(srv, &server.MetricsCollectServer{Storage: s})
			// defer srv.Close()
			srv.Run()
		} else {
			srv := server.NewServerGRPC(envVariables.Address, nil)
			pb.RegisterMetricsCollectServer(srv, &server.MetricsCollectServer{Storage: s})
			// defer srv.Close()
			srv.Run()
		}

	} else {
		// Create a new Chi router instance.
		r := chi.NewRouter()

		// Attach middleware to the router.
		r.Use(middleware.RequestID)
		r.Use(middleware.RealIP)
		r.Use(middleware.Logger)
		r.Use(middleware.Recoverer)
		r.Use(middlewares.GzipMiddleware)
		r.Use(middlewares.SubnetMiddleware(envVariables.TrustedSubnet))

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
		if envVariables.CryptoKeyDir != "" {
			serverTLSCert, err := tls.LoadX509KeyPair(
				path.Join(envVariables.CryptoKeyDir, "server/certServer.pem"),
				path.Join(envVariables.CryptoKeyDir, "server/privateKeyServer.pem"),
			)

			if err != nil {
				server.MyLog.Fatalf("Error loading certificate and key file: %v", err)
			}

			// Configure the server to trust TLS client cert issued by your CA.
			certPool := x509.NewCertPool()
			if caCertPEM, err := os.ReadFile(path.Join(envVariables.CryptoKeyDir, "root/certRoot.pem")); err != nil {
				server.MyLog.Fatal(err)
			} else if ok := certPool.AppendCertsFromPEM(caCertPEM); !ok {
				server.MyLog.Fatal("invalid cert in CA PEM")
			}
			tlsConfig := &tls.Config{
				ClientAuth:   tls.RequireAndVerifyClientCert,
				ClientCAs:    certPool,
				Certificates: []tls.Certificate{serverTLSCert},
			}

			srv = server.NewServerTLS(envVariables.Address, r, tlsConfig)
			// defer srv.Close()
			srv.RunTLS()
		} else {
			srv = server.NewServer(envVariables.Address, r)
			// defer srv.Close()
			srv.Run()
		}
	}
}
