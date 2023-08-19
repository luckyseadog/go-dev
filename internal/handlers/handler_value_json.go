package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/luckyseadog/go-dev/internal/metrics"
	"github.com/luckyseadog/go-dev/internal/security"
	"github.com/luckyseadog/go-dev/internal/storage"
)

// HandlerValueJSON is an HTTP handler that responds to POST requests by sending required metrics.
// It performs verification of the provided metric data integrity
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
func HandlerValueJSON(w http.ResponseWriter, r *http.Request, storage storage.Storage, key []byte) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "HandlerValueJSON: Only POST requests are allowed!", http.StatusMethodNotAllowed)
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "HandlerValueJSON: Read body error", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

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

	for i := 0; i < len(metricsCurrent); i++ {
		if metricsCurrent[i].Value != nil || metricsCurrent[i].Delta != nil {
			http.Error(w, "HandlerValueJSON: Fields value and delta should be empty", http.StatusBadRequest)
			return
		}
		metricID, metricType := metricsCurrent[i].ID, metricsCurrent[i].MType

		switch metricType {
		case "gauge":
			res := storage.LoadContext(r.Context(), metricType, metrics.Metric(metricID))
			if res.Err != nil {
				http.Error(w, "HandlerValueJSON: No such metric", http.StatusNotFound)
				return
			}
			valueFloat64 := float64(res.Value.(metrics.Gauge))
			metricsCurrent[i].Value = &valueFloat64
			if len(key) > 0 {
				metricsCurrent[i].Hash = security.Hash(fmt.Sprintf("%s:gauge:%f", metricsCurrent[i].ID, valueFloat64), key)
			}
		case "counter":
			res := storage.LoadContext(r.Context(), metricType, metrics.Metric(metricID))
			if res.Err != nil {
				http.Error(w, "HandlerValueJSON: No such metric", http.StatusNotFound)
				return
			}
			valueInt64 := int64(res.Value.(metrics.Counter))
			metricsCurrent[i].Delta = &valueInt64
			if len(key) > 0 {
				metricsCurrent[i].Hash = security.Hash(fmt.Sprintf("%s:counter:%d", metricsCurrent[i].ID, valueInt64), key)
			}
		default:
			http.Error(w, "HandlerValueJSON: Not allowed type", http.StatusNotImplemented)
			return
		}
	}

	var jsonData []byte
	if isPackData {
		jsonData, err = json.Marshal(metricsCurrent)
	} else {
		jsonData, err = json.Marshal(metricsCurrent[0])
	}

	if err != nil {
		http.Error(w, "HandlerValueJSON: Error in making response", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonData)

	if err != nil {
		http.Error(w, "HandlerValueJSON: Error in making response", http.StatusInternalServerError)
		return
	}
}
