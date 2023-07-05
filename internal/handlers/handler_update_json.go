package handlers

import (
	"compress/gzip"
	"crypto/hmac"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/luckyseadog/go-dev/internal/security"

	"github.com/luckyseadog/go-dev/internal/metrics"
	"github.com/luckyseadog/go-dev/internal/storage"
)

func HandlerUpdateJSON(w http.ResponseWriter, r *http.Request, storage storage.Storage, key []byte) {
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

	var isPackData = true
	metricsCurrent := make([]metrics.Metrics, 0)

	err = json.Unmarshal(body, &metricsCurrent)
	if err != nil {
		var metric metrics.Metrics
		err = json.Unmarshal(body, &metric)
		if err != nil {
			http.Error(w, "HandlerValueJSON: unmarshal error", http.StatusBadRequest)
			return
		} else {
			isPackData = false
			metricsCurrent = append(metricsCurrent, metric)
		}
	}

	for _, metric := range metricsCurrent {
		if metric.MType == "gauge" {
			if metric.Value == nil || metric.Delta != nil {
				http.Error(w, "HandlerUpdateJSON: Error in passing metric gauge", http.StatusBadRequest)
				return
			}

			if len(key) > 0 {
				computedHash := security.Hash(fmt.Sprintf("%s:gauge:%f", metric.ID, *metric.Value), key)
				decodedComputedHash, err := hex.DecodeString(computedHash)
				if err != nil {
					log.Println(err)
				}
				decodedMetricHash, err := hex.DecodeString(metric.Hash)
				if err != nil {
					log.Println(err)
				}
				if !hmac.Equal(decodedComputedHash, decodedMetricHash) {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
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

			if len(key) > 0 {
				//fmt.Println(fmt.Sprintf("%s:counter:%d", string(key), *metric.Delta))
				//fmt.Println(key)
				computedHash := security.Hash(fmt.Sprintf("%s:counter:%d", metric.ID, *metric.Delta), key)
				decodedComputedHash, err := hex.DecodeString(computedHash)
				if err != nil {
					log.Println(err)
				}
				decodedMetricHash, err := hex.DecodeString(metric.Hash)
				if err != nil {
					log.Println(err)
				}
				if !hmac.Equal(decodedComputedHash, decodedMetricHash) {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
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
			hashMetric := security.Hash(fmt.Sprintf("%s:gauge:%f", metric.ID, valueFloat64), key)
			metricsAnswer = append(metricsAnswer, metrics.Metrics{ID: metric.ID, MType: metric.MType, Value: &valueFloat64, Hash: hashMetric})
		} else if metric.MType == "counter" {
			valueInt64 := int64(value.(metrics.Counter))
			hashMetric := security.Hash(fmt.Sprintf("%s:counter:%d", metric.ID, valueInt64), key)
			metricsAnswer = append(metricsAnswer, metrics.Metrics{ID: metric.ID, MType: metric.MType, Delta: &valueInt64, Hash: hashMetric})
		} else {
			http.Error(w, "HandlerUpdateJSON: Load error", http.StatusInternalServerError)
			return
		}
	}

	var jsonData []byte
	if isPackData {
		jsonData, err = json.Marshal(metricsAnswer)
	} else {
		jsonData, err = json.Marshal(metricsAnswer[0])
	}

	if err != nil {
		http.Error(w, "HandlerUpdateJSON: Error in making response", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonData)

	if err != nil {
		http.Error(w, "HandlerUpdateJSON: Error in making response", http.StatusInternalServerError)
		return
	}
}
