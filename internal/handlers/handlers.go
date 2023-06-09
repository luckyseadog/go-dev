package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/luckyseadog/go-dev/internal/metrics"
	"github.com/luckyseadog/go-dev/internal/storage"
	"io"
	"net/http"
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

func HandlerValue(w http.ResponseWriter, r *http.Request) {
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
		metricId, metricType := metricsCurrent[i].ID, metricsCurrent[i].MType

		if metricType == "gauge" {
			value, err := storage.StorageVar.Load(metrics.Metric(metricId))
			if err != nil {
				http.Error(w, "no such metric", http.StatusNotFound)
				return
			}
			valueFloat64 := float64(value.(metrics.Gauge))
			metricsCurrent[i].Value = &valueFloat64
		} else if metricType == "counter" {
			value, err := storage.StorageVar.Load(metrics.Metric(metricId))
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
