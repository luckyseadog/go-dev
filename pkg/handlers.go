package pkg

import (
	"encoding/json"
	"github.com/luckyseadog/go-dev/internal"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func HandlerDefault(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("This is a server for updating metrics."))
}

func HandlerUpdate(w http.ResponseWriter, r *http.Request) {
	log.Println(StorageVar)
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed!", http.StatusMethodNotAllowed)
		return
	}

	splitPath := strings.Split(r.URL.Path, "/")

	metricType, metric, metricValueString := splitPath[len(splitPath)-3],
		internal.Metric(splitPath[len(splitPath)-2]), splitPath[len(splitPath)-1]

	if metricType == "gauge" {
		metricValue, err := strconv.ParseFloat(metricValueString, 64)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		err = StorageVar.Store(metric, internal.Gauge(metricValue))
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		jsonData, err := json.Marshal(StorageVar)
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
	} else if metricType == "counter" {
		metricValue, err := strconv.Atoi(metricValueString)
		if err != nil {
			http.Error(w, "", http.StatusInternalServerError)
		}

		err = StorageVar.Store(metric, internal.Counter(metricValue))
		if err != nil {
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		jsonData, err := json.Marshal(StorageVar)
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
		http.Error(w, "", http.StatusNotFound)
	}
}
