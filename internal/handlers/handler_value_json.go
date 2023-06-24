package handlers

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"

	"github.com/luckyseadog/go-dev/internal/metrics"
	"github.com/luckyseadog/go-dev/internal/storage"
)

func HandlerValueJSON(w http.ResponseWriter, r *http.Request, storage storage.Storage) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "HandlerValueJSON: Only POST requests are allowed!", http.StatusMethodNotAllowed)
	}

	var reader io.Reader
	if r.Header.Get("Content-Encoding") == "gzip" {
		gz, err := gzip.NewReader(r.Body)
		if err != nil {
			http.Error(w, "HandlerValueJSON: error in reading gzip", http.StatusInternalServerError)
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
		http.Error(w, "HandlerValueJSON: Read body error", http.StatusBadRequest)
		return
	}

	metricsCurrent := make([]metrics.Metrics, 0)

	if body[0] != byte('[') {
		body = append([]byte{byte('[')}, body...)
		body = append(body, byte(']'))
	}

	err = json.Unmarshal(body, &metricsCurrent)
	if err != nil {
		http.Error(w, "HandlerValueJSON: unmarshal error", http.StatusBadRequest)
		return
	}

	for i := 0; i < len(metricsCurrent); i++ {
		if metricsCurrent[i].Value != nil || metricsCurrent[i].Delta != nil {
			http.Error(w, "HandlerValueJSON: Fields value and delta should be empty", http.StatusBadRequest)
			return
		}
		metricID, metricType := metricsCurrent[i].ID, metricsCurrent[i].MType

		if metricType == "gauge" {
			value, err := storage.Load(metricType, metrics.Metric(metricID))
			if err != nil {
				http.Error(w, "HandlerValueJSON: No such metric", http.StatusNotFound)
				return
			}
			valueFloat64 := float64(value.(metrics.Gauge))
			metricsCurrent[i].Value = &valueFloat64
		} else if metricType == "counter" {
			value, err := storage.Load(metricType, metrics.Metric(metricID))
			if err != nil {
				http.Error(w, "HandlerValueJSON: No such metric", http.StatusNotFound)
				return
			}
			valueInt64 := int64(value.(metrics.Counter))
			metricsCurrent[i].Delta = &valueInt64
		} else {
			http.Error(w, "HandlerValueJSON: Not allowed type", http.StatusNotImplemented)
			return
		}
	}
	jsonData, err := json.Marshal(metricsCurrent)
	if err != nil {
		http.Error(w, "HandlerValueJSON: Error in making response", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if len(metricsCurrent) == 1 {
		_, err = w.Write(jsonData[1 : len(jsonData)-1])
	} else {
		_, err = w.Write(jsonData)
	}

	if err != nil {
		http.Error(w, "HandlerValueJSON: Error in making response", http.StatusInternalServerError)
		return
	}
}
