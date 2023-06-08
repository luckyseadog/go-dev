package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/luckyseadog/go-dev/internal/metrics"
	"github.com/luckyseadog/go-dev/internal/storage"
	"net/http"
	"strconv"
	"strings"
)

func HandlerDefault(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	_, err := fmt.Fprintf(w, "<html><body>")
	if err != nil {
		http.Error(w, "error when writing to html", http.StatusInternalServerError)
		return
	}
	for key := range storage.StorageVar.DataGauge {
		_, err = fmt.Fprintf(w, "<p>%s</p>", string(key))
		if err != nil {
			http.Error(w, "error when writing to html", http.StatusInternalServerError)
			return
		}
	}
	for key := range storage.StorageVar.DataCounter {
		_, err = fmt.Fprintf(w, "<p>%s</p>", string(key))
		if err != nil {
			http.Error(w, "error when writing to html", http.StatusInternalServerError)
			return
		}
	}
	_, err = fmt.Fprintf(w, "</body></html>")
	if err != nil {
		http.Error(w, "error when writing to html", http.StatusInternalServerError)
		return
	}
}

func HandlerGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET requests are allowed!", http.StatusMethodNotAllowed)
	}

	splitPath := strings.Split(r.URL.Path, "/")
	if len(splitPath) != 4 {
		http.Error(w, "invalid update", http.StatusNotFound)
		return
	}

	metricType, metricName := splitPath[len(splitPath)-2], splitPath[len(splitPath)-1]

	if metricType == "gauge" {
		value, err := storage.StorageVar.Load(metrics.Metric(metricName))
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		valueGauge, ok := value.(metrics.Gauge)
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(fmt.Sprintf("%g", valueGauge)))
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

	} else if metricType == "counter" {
		value, err := storage.StorageVar.Load(metrics.Metric(metricName))
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		valueCounter, ok := value.(metrics.Counter)
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(fmt.Sprintf("%d", valueCounter)))
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	}

}

func HandlerUpdate(w http.ResponseWriter, r *http.Request) {
	//log.Println(StorageVar)
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	splitPath := strings.Split(r.URL.Path, "/")
	if len(splitPath) != 5 {
		http.Error(w, "invalid update", http.StatusNotFound)
		return
	}

	metricType, metric, metricValueString := splitPath[len(splitPath)-3],
		metrics.Metric(splitPath[len(splitPath)-2]), splitPath[len(splitPath)-1]

	if metricType == "gauge" {
		metricValue, err := strconv.ParseFloat(metricValueString, 64)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}
		err = storage.StorageVar.Store(metric, metrics.Gauge(metricValue))
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		jsonData, err := json.Marshal(storage.StorageVar)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(jsonData)

		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	} else if metricType == "counter" {
		metricValue, err := strconv.Atoi(metricValueString)
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
		}

		err = storage.StorageVar.Store(metric, metrics.Counter(metricValue))
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		jsonData, err := json.Marshal(storage.StorageVar)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write(jsonData)

		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

	} else {
		http.Error(w, "", http.StatusNotImplemented)
		return
	}
}
