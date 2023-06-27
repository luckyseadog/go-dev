package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/luckyseadog/go-dev/internal/metrics"
	"github.com/luckyseadog/go-dev/internal/storage"
)

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

	if metricType == "gauge" {
		metricValue, err := strconv.ParseFloat(metricValueString, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = storage.Store(metric, metrics.Gauge(metricValue))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		jsonData, err := json.Marshal(storage.LoadDataGauge())
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
	} else if metricType == "counter" {
		metricValue, err := strconv.Atoi(metricValueString)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}

		err = storage.Store(metric, metrics.Counter(metricValue))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		jsonData, err := json.Marshal(storage.LoadDataCounter())
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

	} else {
		w.WriteHeader(http.StatusNotImplemented)
		return
	}
}
