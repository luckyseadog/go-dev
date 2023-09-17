package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/luckyseadog/go-dev/internal/metrics"
	"github.com/luckyseadog/go-dev/internal/storage"
)

// HandlerUpdate is an HTTP handler that responds to POST requests by soring metric into the storage.
// It stores metric into the storage by retrieving URL parameters and generates an JSON response
// containing the new value of metric.
// The function checks for errors while retrieving metrics and writing to the response, returning appropriate HTTP error
// responses in case of errors.
//
// Parameters:
//   - w: The http.ResponseWriter to write the HTTP response.
//   - r: The http.Request received from the client.
//   - storage: An instance of storage.Storage used to retrieve metric data.
//
// Notes:
//   - Only POST requests are allowed. For other request methods, the function responds with a "Method Not Allowed" error.
//   - The function generates an JSON response containing metric value.
func HandlerUpdate(w http.ResponseWriter, r *http.Request, storage storage.Storage) {
	if r.Method != http.MethodPost {
		http.Error(w, "HandlerUpdate: Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	splitPath := strings.Split(r.URL.Path, "/")
	if len(splitPath) != 5 {
		http.Error(w, "HandlerUpdate: Invalid update", http.StatusNotFound)
		return
	}

	metricType, metric, metricValueString := splitPath[len(splitPath)-3],
		metrics.Metric(splitPath[len(splitPath)-2]), splitPath[len(splitPath)-1]

	switch metricType {
	case "gauge":
		metricValue, err := strconv.ParseFloat(metricValueString, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = storage.StoreContext(r.Context(), metric, metrics.Gauge(metricValue))
		if err != nil {
			http.Error(w, "HandlerUpdate: "+err.Error(), http.StatusInternalServerError)
			return
		}

		res := storage.LoadContext(r.Context(), metricType, metric)
		if res.Err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		valueGauge, ok := res.Value.(metrics.Gauge)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		jsonData, err := json.Marshal(valueGauge)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(jsonData)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	case "counter":
		metricValue, err := strconv.Atoi(metricValueString)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}

		err = storage.StoreContext(r.Context(), metric, metrics.Counter(metricValue))
		if err != nil {
			http.Error(w, "HandlerUpdate: "+err.Error(), http.StatusInternalServerError)
			return
		}

		res := storage.LoadContext(r.Context(), metricType, metric)
		if res.Err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		valueCounter, ok := res.Value.(metrics.Counter)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		jsonData, err := json.Marshal(valueCounter)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(jsonData)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	default:
		w.WriteHeader(http.StatusNotImplemented)
		return
	}
}
