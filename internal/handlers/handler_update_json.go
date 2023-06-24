package handlers

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"

	"github.com/luckyseadog/go-dev/internal/metrics"
	"github.com/luckyseadog/go-dev/internal/storage"
)

func HandlerUpdateJSON(w http.ResponseWriter, r *http.Request, storage storage.Storage) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "HandlerUpdateJSON: Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	var reader io.Reader
	if r.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, "HandlerUpdateJSON: error in reading gzip", http.StatusInternalServerError)
			defer r.Body.Close()
			return
		}
		reader = gz
		defer gz.Close()
	} else {
		reader = r.Body
		defer r.Body.Close()
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		http.Error(w, "HandlerUpdateJSON: Read body error", http.StatusBadRequest)
		return
	}

	if body[0] != byte('[') {
		body = append([]byte{byte('[')}, body...)
		body = append(body, byte(']'))
	}

	metricsCurrent := make([]metrics.Metrics, 0)

	err = json.Unmarshal(body, &metricsCurrent)
	if err != nil {
		http.Error(w, "HandlerUpdateJSON: Unmarshal error", http.StatusBadRequest)
		return
	}

	for _, metric := range metricsCurrent {
		if metric.MType == "gauge" {
			if metric.Value == nil || metric.Delta != nil {
				http.Error(w, "HandlerUpdateJSON: Error in passing metric gauge", http.StatusBadRequest)
				return
			}
			err = storage.Store(metrics.Metric(metric.ID), metrics.Gauge(*metric.Value))
			if err != nil {
				http.Error(w, "HandlerUpdateJSON: Could not store gauge", http.StatusInternalServerError)
				return
			}

		} else if metric.MType == "counter" {
			if metric.Delta == nil || metric.Value != nil {
				http.Error(w, "HandlerUpdateJSON: Error in passing metric counter", http.StatusBadRequest)
				return
			}
			err = storage.Store(metrics.Metric(metric.ID), metrics.Counter(*metric.Delta))
			if err != nil {
				http.Error(w, "HandlerUpdateJSON: Could not store counter", http.StatusInternalServerError)
				return
			}

		} else {
			http.Error(w, "HandlerUpdateJSON: Not allowed type", http.StatusNotImplemented)
			return
		}
	}

	metricsAnswer := make([]metrics.Metrics, 0)

	for _, metric := range metricsCurrent {
		value, err := storage.Load(metric.MType, metrics.Metric(metric.ID))
		if err != nil {
			http.Error(w, "HandlerUpdateJSON: Load error", http.StatusInternalServerError)
			return
		}
		if metric.MType == "gauge" {
			valueFloat64 := float64(value.(metrics.Gauge))
			metricsAnswer = append(metricsAnswer, metrics.Metrics{ID: metric.ID, MType: metric.MType, Value: &valueFloat64})
		} else if metric.MType == "counter" {
			valueInt64 := int64(value.(metrics.Counter))
			metricsAnswer = append(metricsAnswer, metrics.Metrics{ID: metric.ID, MType: metric.MType, Delta: &valueInt64})
		} else {
			http.Error(w, "HandlerUpdateJSON: Load error", http.StatusInternalServerError)
			return
		}
	}

	jsonData, err := json.Marshal(metricsAnswer)
	if err != nil {
		http.Error(w, "HandlerUpdateJSON: Error in making response", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if len(metricsAnswer) == 1 {
		_, err = w.Write(jsonData[1 : len(jsonData)-1])
	} else {
		_, err = w.Write(jsonData)
	}

	if err != nil {
		http.Error(w, "HandlerUpdateJSON: Error in making response", http.StatusInternalServerError)
		return
	}
}
