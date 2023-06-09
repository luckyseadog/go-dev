package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/luckyseadog/go-dev/internal/metrics"
	"github.com/luckyseadog/go-dev/internal/storage"
	"io"
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

func HandlerValueJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
	}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body error", http.StatusBadRequest)
		return
	}

	metricsCurrent := make([]metrics.Metrics, 0)

	err = json.Unmarshal(body, &metricsCurrent)
	if err != nil {
		http.Error(w, "unmarshal error", http.StatusBadRequest)
		return
	}

	for i := 0; i < len(metricsCurrent); i++ {
		if metricsCurrent[i].Value != nil || metricsCurrent[i].Delta != nil {
			http.Error(w, "fields value and delta should be empty", http.StatusBadRequest)
			return
		}
		metricID, metricType := metricsCurrent[i].ID, metricsCurrent[i].MType

		if metricType == "gauge" {
			value, err := storage.StorageVar.Load(metrics.Metric(metricID))
			if err != nil {
				http.Error(w, "no such metric", http.StatusNotFound)
				return
			}
			valueFloat64 := float64(value.(metrics.Gauge))
			metricsCurrent[i].Value = &valueFloat64
		} else if metricType == "counter" {
			value, err := storage.StorageVar.Load(metrics.Metric(metricID))
			if err != nil {
				http.Error(w, "no such metric", http.StatusNotFound)
				return
			}
			valueInt64 := int64(value.(metrics.Counter))
			metricsCurrent[i].Delta = &valueInt64
		} else {
			http.Error(w, "not allowed type", http.StatusNotImplemented)
			return
		}
	}
	jsonData, err := json.Marshal(metricsCurrent)
	if err != nil {
		http.Error(w, "error in making response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonData)

	if err != nil {
		http.Error(w, "error in making response", http.StatusInternalServerError)
		return
	}
}

func HandlerUpdate(w http.ResponseWriter, r *http.Request) {
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

func HandlerUpdateJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "read body error", http.StatusBadRequest)
		return
	}

	metricsCurrent := make([]metrics.Metrics, 0)

	err = json.Unmarshal(body, &metricsCurrent)
	if err != nil {
		http.Error(w, "unmarshal error", http.StatusBadRequest)
		return
	}

	for _, metric := range metricsCurrent {
		if metric.MType == "gauge" {
			if metric.Value == nil || metric.Delta != nil {
				http.Error(w, "error in passing metric: gauge", http.StatusBadRequest)
				return
			}
			err = storage.StorageVar.Store(metrics.Metric(metric.ID), metrics.Gauge(*metric.Value))
			if err != nil {
				http.Error(w, "could not store gauge", http.StatusInternalServerError)
				return
			}

		} else if metric.MType == "counter" {
			if metric.Delta == nil || metric.Value != nil {
				http.Error(w, "error in passing metric: counter", http.StatusInternalServerError)
				return
			}
			err = storage.StorageVar.Store(metrics.Metric(metric.ID), metrics.Counter(*metric.Delta))
			if err != nil {
				http.Error(w, "could not store counter", http.StatusInternalServerError)
				return
			}

		} else {
			http.Error(w, "not allowed type", http.StatusNotImplemented)
			return
		}
	}

	metricsCurrent = make([]metrics.Metrics, 0)

	for key, value := range storage.StorageVar.DataGauge {
		valueFloat64 := float64(value)
		metricsCurrent = append(metricsCurrent, metrics.Metrics{ID: string(key), MType: "gauge", Value: &valueFloat64})
	}

	for key, value := range storage.StorageVar.DataCounter {
		valueInt64 := int64(value)
		metricsCurrent = append(metricsCurrent, metrics.Metrics{ID: string(key), MType: "counter", Delta: &valueInt64})
	}

	jsonData, err := json.Marshal(metricsCurrent)
	if err != nil {
		http.Error(w, "error in making response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonData)

	if err != nil {
		http.Error(w, "error in making response", http.StatusInternalServerError)
		return
	}
}
