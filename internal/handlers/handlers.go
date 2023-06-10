package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/luckyseadog/go-dev/internal/metrics"
	"github.com/luckyseadog/go-dev/internal/storage"
)

func HandlerDefault(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "HandlerDefault: Only GET requests are allowed!", http.StatusMethodNotAllowed)
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	_, err := fmt.Fprintf(w, "<html><body>")
	if err != nil {
		http.Error(w, "HandlerDefault: error when writing to html", http.StatusInternalServerError)
		return
	}
	for key := range storage.StorageVar.DataGauge {
		_, err = fmt.Fprintf(w, "<p>%s</p>", string(key))
		if err != nil {
			http.Error(w, "HandlerDefault: error when writing to html", http.StatusInternalServerError)
			return
		}
	}
	for key := range storage.StorageVar.DataCounter {
		_, err = fmt.Fprintf(w, "<p>%s</p>", string(key))
		if err != nil {
			http.Error(w, "HandlerDefault: error when writing to html", http.StatusInternalServerError)
			return
		}
	}
	_, err = fmt.Fprintf(w, "</body></html>")
	if err != nil {
		http.Error(w, "HandlerDefault: error when writing to html", http.StatusInternalServerError)
		return
	}
}

func HandlerGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "HandlerDefault: Only GET requests are allowed!", http.StatusMethodNotAllowed)
	}

	splitPath := strings.Split(r.URL.Path, "/")
	if len(splitPath) != 4 {
		http.Error(w, "HandlerDefault: invalid update", http.StatusNotFound)
		return
	}

	metricType, metricName := splitPath[len(splitPath)-2], splitPath[len(splitPath)-1]

	if metricType == "gauge" {
		value, err := storage.StorageVar.Load(metricType, metrics.Metric(metricName))
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
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	} else if metricType == "counter" {
		value, err := storage.StorageVar.Load(metricType, metrics.Metric(metricName))
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
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

}

func HandlerValueJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "HandlerValueJSON: Only POST requests are allowed!", http.StatusMethodNotAllowed)
	}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
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
			value, err := storage.StorageVar.Load(metricType, metrics.Metric(metricID))
			if err != nil {
				http.Error(w, "HandlerValueJSON: No such metric", http.StatusNotFound)
				return
			}
			valueFloat64 := float64(value.(metrics.Gauge))
			metricsCurrent[i].Value = &valueFloat64
		} else if metricType == "counter" {
			value, err := storage.StorageVar.Load(metricType, metrics.Metric(metricID))
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

func HandlerUpdate(w http.ResponseWriter, r *http.Request) {
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
		err = storage.StorageVar.Store(metric, metrics.Gauge(metricValue))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		jsonData, err := json.Marshal(storage.StorageVar)
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

		err = storage.StorageVar.Store(metric, metrics.Counter(metricValue))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
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
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	} else {
		w.WriteHeader(http.StatusNotImplemented)
		return
	}
}

func HandlerUpdateJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, "HandlerUpdateJSON: Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
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
			err = storage.StorageVar.Store(metrics.Metric(metric.ID), metrics.Gauge(*metric.Value))
			if err != nil {
				http.Error(w, "HandlerUpdateJSON: Could not store gauge", http.StatusInternalServerError)
				return
			}

		} else if metric.MType == "counter" {
			if metric.Delta == nil || metric.Value != nil {
				http.Error(w, "HandlerUpdateJSON: Error in passing metric counter", http.StatusBadRequest)
				return
			}
			err = storage.StorageVar.Store(metrics.Metric(metric.ID), metrics.Counter(*metric.Delta))
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
		value, err := storage.StorageVar.Load(metric.MType, metrics.Metric(metric.ID))
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

	//for key, value := range storage.StorageVar.DataGauge {
	//	valueFloat64 := float64(value)
	//	metricsCurrent = append(metricsCurrent, metrics.Metrics{ID: string(key), MType: "gauge", Value: &valueFloat64})
	//}
	//
	//for key, value := range storage.StorageVar.DataCounter {
	//	valueInt64 := int64(value)
	//	metricsCurrent = append(metricsCurrent, metrics.Metrics{ID: string(key), MType: "counter", Delta: &valueInt64})
	//}

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
