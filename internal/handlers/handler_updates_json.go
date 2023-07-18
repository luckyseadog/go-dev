package handlers

import (
	"compress/gzip"
	"crypto/hmac"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/luckyseadog/go-dev/internal/security"

	"github.com/luckyseadog/go-dev/internal/metrics"
	"github.com/luckyseadog/go-dev/internal/storage"
)

func HandlerUpdatesJSON(w http.ResponseWriter, r *http.Request, storage storage.Storage, key []byte) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "HandlerUpdatesJSON: Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	var reader io.Reader
	if r.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, "HandlerUpdatesJSON: error in reading gzip", http.StatusInternalServerError)
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
		http.Error(w, "HandlerUpdatesJSON: Read body error", http.StatusBadRequest)
		return
	}

	metricsCurrent := make([]metrics.Metrics, 0)

	err = json.Unmarshal(body, &metricsCurrent)
	if err != nil {
		http.Error(w, "HandlerUpdatesJSON: unmarshal error", http.StatusBadRequest)
		return
	}

	for _, metric := range metricsCurrent {
		switch metric.MType {
		case "gauge":
			if metric.Value == nil || metric.Delta != nil {
				http.Error(w, "HandlerUpdatesJSON: Error in passing metric gauge", http.StatusBadRequest)
				return
			}

			if len(key) > 0 {
				computedHash := security.Hash(fmt.Sprintf("%s:gauge:%f", metric.ID, *metric.Value), key)
				decodedComputedHash, err := hex.DecodeString(computedHash)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				decodedMetricHash, err := hex.DecodeString(metric.Hash)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				if !hmac.Equal(decodedComputedHash, decodedMetricHash) {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
			}
			err = storage.StoreContext(r.Context(), metrics.Metric(metric.ID), metrics.Gauge(*metric.Value))
			if err != nil {
				http.Error(w, "HandlerUpdatesJSON: Could not store gauge", http.StatusInternalServerError)
				return
			}

		case "counter":
			if metric.Delta == nil || metric.Value != nil {
				http.Error(w, "HandlerUpdatesJSON: Error in passing metric counter", http.StatusBadRequest)
				return
			}

			if len(key) > 0 {
				computedHash := security.Hash(fmt.Sprintf("%s:counter:%d", metric.ID, *metric.Delta), key)
				decodedComputedHash, err := hex.DecodeString(computedHash)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				decodedMetricHash, err := hex.DecodeString(metric.Hash)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				if !hmac.Equal(decodedComputedHash, decodedMetricHash) {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
			}
			err = storage.StoreContext(r.Context(), metrics.Metric(metric.ID), metrics.Counter(*metric.Delta))
			if err != nil {
				http.Error(w, "HandlerUpdatesJSON: Could not store counter", http.StatusInternalServerError)
				return
			}

		default:
			http.Error(w, "HandlerUpdatesJSON: Not allowed type", http.StatusNotImplemented)
			return
		}
	}

	metricsAnswer := make([]metrics.Metrics, 0)

	for _, metric := range metricsCurrent {
		res := storage.LoadContext(r.Context(), metric.MType, metrics.Metric(metric.ID))
		if res.Err != nil {
			http.Error(w, "HandlerUpdatesJSON: Load error", http.StatusInternalServerError)
			return
		}
		switch metric.MType {
		case "gauge":
			valueFloat64 := float64(res.Value.(metrics.Gauge))
			hashMetric := security.Hash(fmt.Sprintf("%s:gauge:%f", metric.ID, valueFloat64), key)
			metricsAnswer = append(metricsAnswer, metrics.Metrics{ID: metric.ID, MType: metric.MType, Value: &valueFloat64, Hash: hashMetric})
		case "counter":
			valueInt64 := int64(res.Value.(metrics.Counter))
			hashMetric := security.Hash(fmt.Sprintf("%s:counter:%d", metric.ID, valueInt64), key)
			metricsAnswer = append(metricsAnswer, metrics.Metrics{ID: metric.ID, MType: metric.MType, Delta: &valueInt64, Hash: hashMetric})
		default:
			http.Error(w, "HandlerUpdatesJSON: Load error", http.StatusInternalServerError)
			return
		}
	}

	jsonData, err := json.Marshal(metricsAnswer)

	if err != nil {
		http.Error(w, "HandlerUpdatesJSON: Error in making response", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonData)

	if err != nil {
		http.Error(w, "HandlerUpdatesJSON: Error in making response", http.StatusInternalServerError)
		return
	}
}
