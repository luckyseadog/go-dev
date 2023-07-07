package handlers

import (
	"fmt"
	"github.com/luckyseadog/go-dev/internal/metrics"
	"net/http"

	"github.com/luckyseadog/go-dev/internal/storage"
)

func HandlerDefault(w http.ResponseWriter, r *http.Request, storage storage.Storage) {
	if r.Method != http.MethodGet {
		http.Error(w, "HandlerDefault: Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	_, err := fmt.Fprintf(w, "<html><body>")
	if err != nil {
		http.Error(w, "HandlerDefault: error when writing to html", http.StatusInternalServerError)
		return
	}

	res := storage.LoadDataGauge()
	if res.Err != nil {
		http.Error(w, "HandlerDefault: error when writing to html", http.StatusInternalServerError)
		return
	}
	for key := range res.Value.(map[metrics.Metric]metrics.Gauge) {
		_, err = fmt.Fprintf(w, "<p>%s</p>", string(key))
		if err != nil {
			http.Error(w, "HandlerDefault: error when writing to html", http.StatusInternalServerError)
			return
		}
	}

	res = storage.LoadDataCounter()
	if res.Err != nil {
		http.Error(w, "HandlerDefault: error when writing to html", http.StatusInternalServerError)
		return
	}

	for key := range res.Value.(map[metrics.Metric]metrics.Counter) {
		_, err = fmt.Fprintf(w, "<p>%s</p>", string(key))
		if err != nil {
			http.Error(w, "HandlerDefault: error when writing to html", http.StatusInternalServerError)
			return
		}
	}
	_, err = fmt.Fprintf(w, "</body></html>")
	if err != nil {
		http.Error(w, "HandlerDefault: error when writing to html", http.StatusInternalServerError)
		return
	}
}
