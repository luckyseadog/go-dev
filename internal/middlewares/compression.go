package middlewares

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type gzipWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (w gzipWriter) Write(b []byte) (int, error) {
	return w.Writer.Write(b)
}

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Encoding") == "gzip" {
			gzr, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "HandlerUpdateJSON: error in reading gzip", http.StatusInternalServerError)
				return
			}
			defer gzr.Close()
			r.Body = gzr
		}

		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			gzw, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
			if err != nil {
				http.Error(w, "HandlerValueJSON: con not wrap writer as gzipWriter", http.StatusInternalServerError)
				return
			}
			defer gzw.Close()
			w = gzipWriter{ResponseWriter: w, Writer: gzw}
			w.Header().Set("Content-Encoding", "gzip")
		}

		next.ServeHTTP(w, r)
	})
}
