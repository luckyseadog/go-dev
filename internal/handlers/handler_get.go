package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/luckyseadog/go-dev/internal/metrics"
	"github.com/luckyseadog/go-dev/internal/storage"
)

// HandlerGet is an HTTP handler that responds to GET requests by sending required metric.
// It gets required metric by parsing URL, where should be parameters such as metricType and metricName.
// The function checks for errors while retrieving metrics and writing to the response, returning appropriate HTTP error
// responses in case of errors.
//
// Parameters:
//   - w: The http.ResponseWriter to write the HTTP response.
//   - r: The http.Request received from the client.
//   - storage: An instance of storage.Storage used to retrieve metric data.
func HandlerGet(w http.ResponseWriter, r *http.Request, storage storage.Storage) {
	if r.Method != http.MethodGet {
		http.Error(w, "HandlerGet: Only GET requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	splitPath := strings.Split(r.URL.Path, "/")
	if len(splitPath) != 4 {
		http.Error(w, "HandlerGet: invalid update", http.StatusNotFound)
		return
	}

	metricType, metricName := splitPath[len(splitPath)-2], splitPath[len(splitPath)-1]

	switch metricType {
	case "gauge":
		res := storage.LoadContext(r.Context(), metricType, metrics.Metric(metricName))
		if res.Err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		valueGauge, ok := res.Value.(metrics.Gauge)
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(fmt.Sprintf("%g", valueGauge)))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	case "counter":
		res := storage.LoadContext(r.Context(), metricType, metrics.Metric(metricName))
		if res.Err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		valueCounter, ok := res.Value.(metrics.Counter)
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(fmt.Sprintf("%d", valueCounter)))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

}
