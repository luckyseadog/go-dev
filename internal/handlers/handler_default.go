package handlers

import (
	"fmt"
	"net/http"

	"github.com/luckyseadog/go-dev/internal/metrics"
	"github.com/luckyseadog/go-dev/internal/storage"
)

// HandlerDefault is an HTTP handler that responds to GET requests by displaying the list of metric IDs in HTML format.
// It retrieves gauge and counter metrics from the provided storage and generates an HTML response containing these metrics.
// The function checks for errors while retrieving metrics and writing to the response, returning appropriate HTTP error
// responses in case of errors.
//
// Parameters:
//   - w: The http.ResponseWriter to write the HTTP response.
//   - r: The http.Request received from the client.
//   - storage: An instance of storage.Storage used to retrieve metric data.
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

	res := storage.LoadDataGaugeContext(r.Context())
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

	res = storage.LoadDataCounterContext(r.Context())
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
