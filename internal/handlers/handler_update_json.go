package handlers

import (
	"crypto/hmac"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/luckyseadog/go-dev/internal/metrics"
	"github.com/luckyseadog/go-dev/internal/security"
	"github.com/luckyseadog/go-dev/internal/storage"
)

// HandlerUpdateJSON is an HTTP handler that responds to POST requests by processing JSON-encoded metric data
// and storing it into the specified storage. It performs verification of the provided metric data integrity
// using a digital signature (if a secret key is provided).
//
// Parameters:
//   - w: http.ResponseWriter to write the response.
//   - r: *http.Request containing the incoming request data.
//   - storage: Storage instance to store the metric data.
//   - key: Secret key used for digital signature verification.
//
// Notes:
//   - For making requests through this method the agent should send JSON array with id and type and
//
// delta or value fields for consistency with Metrics.
func HandlerUpdateJSON(w http.ResponseWriter, r *http.Request, storage storage.Storage, key []byte) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "HandlerUpdateJSON: Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "HandlerUpdateJSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var metricCurrent metrics.Metrics

	err = json.Unmarshal(body, &metricCurrent)
	if err != nil {
		http.Error(w, "HandlerUpdateJSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	switch metricCurrent.MType {
	case "gauge":
		if metricCurrent.Value == nil || metricCurrent.Delta != nil {
			http.Error(w, "HandlerUpdateJSON: Error in passing metric gauge", http.StatusBadRequest)
			return
		}

		if len(key) > 0 {
			computedHash := security.Hash(fmt.Sprintf("%s:gauge:%f", metricCurrent.ID, *metricCurrent.Value), key)
			decodedComputedHash, err := hex.DecodeString(computedHash)
			if err != nil {
				log.Println(err)
			}
			decodedMetricHash, err := hex.DecodeString(metricCurrent.Hash)
			if err != nil {
				log.Println(err)
			}
			if !hmac.Equal(decodedComputedHash, decodedMetricHash) {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		err = storage.StoreContext(r.Context(), metrics.Metric(metricCurrent.ID), metrics.Gauge(*metricCurrent.Value))
		if err != nil {
			http.Error(w, "HandlerUpdateJSON: "+err.Error(), http.StatusInternalServerError)
			return
		}

	case "counter":
		if metricCurrent.Delta == nil || metricCurrent.Value != nil {
			http.Error(w, "HandlerUpdateJSON: Error in passing metric counter", http.StatusBadRequest)
			return
		}

		if len(key) > 0 {
			//fmt.Println(fmt.Sprintf("%s:counter:%d", string(key), *metric.Delta))
			//fmt.Println(key)
			computedHash := security.Hash(fmt.Sprintf("%s:counter:%d", metricCurrent.ID, *metricCurrent.Delta), key)
			decodedComputedHash, err := hex.DecodeString(computedHash)
			if err != nil {
				log.Println(err)
			}
			decodedMetricHash, err := hex.DecodeString(metricCurrent.Hash)
			if err != nil {
				log.Println(err)
			}
			if !hmac.Equal(decodedComputedHash, decodedMetricHash) {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}

		err = storage.StoreContext(r.Context(), metrics.Metric(metricCurrent.ID), metrics.Counter(*metricCurrent.Delta))
		if err != nil {
			http.Error(w, "HandlerUpdateJSON: "+err.Error(), http.StatusInternalServerError)
			return
		}

	default:
		http.Error(w, "HandlerUpdateJSON: Not allowed type", http.StatusNotImplemented)
		return
	}

	var metricsAnswer metrics.Metrics

	res := storage.LoadContext(r.Context(), metricCurrent.MType, metrics.Metric(metricCurrent.ID))
	if res.Err != nil {
		http.Error(w, "HandlerUpdateJSON: "+res.Err.Error(), http.StatusInternalServerError)
		return
	}
	switch metricCurrent.MType {
	case "gauge":
		valueFloat64 := float64(res.Value.(metrics.Gauge))
		hashMetric := security.Hash(fmt.Sprintf("%s:gauge:%f", metricCurrent.ID, valueFloat64), key)
		metricsAnswer = metrics.Metrics{ID: metricCurrent.ID, MType: metricCurrent.MType, Value: &valueFloat64, Hash: hashMetric}
	case "counter":
		valueInt64 := int64(res.Value.(metrics.Counter))
		hashMetric := security.Hash(fmt.Sprintf("%s:counter:%d", metricCurrent.ID, valueInt64), key)
		metricsAnswer = metrics.Metrics{ID: metricCurrent.ID, MType: metricCurrent.MType, Delta: &valueInt64, Hash: hashMetric}
	default:
		http.Error(w, "HandlerUpdateJSON: Load error", http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(metricsAnswer)
	if err != nil {
		http.Error(w, "HandlerUpdateJSON: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonData)

	if err != nil {
		http.Error(w, "HandlerUpdateJSON: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
