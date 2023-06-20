package storage

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/luckyseadog/go-dev/internal/metrics"
)

var errNotExpectedType = errors.New("not expected type")

type Storage interface {
	Store(metric metrics.Metric, metricValue any) error
	Load(metricType string, metric metrics.Metric) (any, error)
	LoadDataGauge() map[metrics.Metric]metrics.Gauge
	LoadDataCounter() map[metrics.Metric]metrics.Counter
	SaveToFile(filepath string) error
	LoadFromFile(filepath string) error
}

type MyStorage struct {
	DataGauge   map[metrics.Metric]metrics.Gauge
	DataCounter map[metrics.Metric]metrics.Counter
	mu          sync.RWMutex
}

func (s *MyStorage) Store(metric metrics.Metric, metricValue any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch metricValue := metricValue.(type) {
	case metrics.Gauge:
		s.DataGauge[metric] = metricValue
		return nil
	case float64:
		s.DataGauge[metric] = metrics.Gauge(metricValue)
		return nil
	case metrics.Counter:
		s.DataCounter[metric] += metricValue
		return nil
	case int64:
		s.DataCounter[metric] += metrics.Counter(metricValue)
		return nil
	default:
		return errNotExpectedType
	}
}

func (s *MyStorage) Load(metricType string, metric metrics.Metric) (any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if metricType == "gauge" {
		if valueGauge, ok := s.DataGauge[metric]; ok {
			return valueGauge, nil
		} else {
			return nil, errors.New("no such metric")
		}
	} else if metricType == "counter" {
		if valueCounter, ok := s.DataCounter[metric]; ok {
			return valueCounter, nil
		} else {
			return nil, errors.New("no such metric")
		}
	} else {
		return nil, errors.New("no such metric")
	}
}

func (s *MyStorage) LoadDataGauge() map[metrics.Metric]metrics.Gauge {
	s.mu.RLock()
	defer s.mu.RUnlock()
	copyDataGauge := make(map[metrics.Metric]metrics.Gauge)

	for key, value := range s.DataGauge {
		copyDataGauge[key] = value
	}

	return copyDataGauge
}

func (s *MyStorage) LoadDataCounter() map[metrics.Metric]metrics.Counter {
	s.mu.RLock()
	defer s.mu.RUnlock()
	copyDataCounter := make(map[metrics.Metric]metrics.Counter)

	for key, value := range s.DataCounter {
		copyDataCounter[key] = value
	}

	return copyDataCounter
}

func (s *MyStorage) SaveToFile(filepath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0777)
	if err != nil {
		return err
	}

	defer file.Close()
	copyMetrics := map[metrics.Metric]any{}
	for key, value := range s.DataGauge {
		copyMetrics[key] = value
	}
	for key, value := range s.DataCounter {
		copyMetrics[key] = value
	}

	writer := bufio.NewWriter(file)
	data, err := json.Marshal(copyMetrics)
	if err != nil {
		return err
	}

	_, err = writer.Write(data)
	if err != nil {
		return err
	}
	_ = writer.WriteByte('\n')

	return writer.Flush()
}

func (s *MyStorage) LoadFromFile(filepath string) error {
	s.mu.Lock()

	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_RDONLY, 0777)
	if err != nil {
		return err
	}

	defer file.Close()

	var lastData []byte
	var metricsLoaded map[string]any
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lastData = scanner.Bytes()
	}

	//s.mu.Unlock()

	if scanner.Err() != nil {
		return scanner.Err()
	}

	err = json.Unmarshal(lastData, &metricsLoaded)
	if err != nil {
		return err
	}
	s.mu.Unlock()

	for key, value := range metricsLoaded {
		err = s.Store(metrics.Metric(key), value)
		if err != nil {
			return err
		}
	}

	return nil

}

func NewStorage() *MyStorage {
	dataGauge := map[metrics.Metric]metrics.Gauge{}
	dataCounter := map[metrics.Metric]metrics.Counter{}
	return &MyStorage{DataGauge: dataGauge, DataCounter: dataCounter, mu: sync.RWMutex{}}
}
