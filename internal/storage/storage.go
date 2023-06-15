package storage

import (
	"errors"
	"sync"

	"github.com/luckyseadog/go-dev/internal/metrics"
)

var errNotExpectedType = errors.New("not expected type")

var StorageVar = NewStorage()

type Storage interface {
	Store(metric metrics.Metric, metricValue any) error
	Load(metricType string, metric metrics.Metric) (any, error)
	LoadDataGauge() map[metrics.Metric]metrics.Gauge
	LoadDataCounter() map[metrics.Metric]metrics.Counter
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
	case metrics.Counter:
		s.DataCounter[metric] += metricValue
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

func NewStorage() *MyStorage {
	dataGauge := map[metrics.Metric]metrics.Gauge{}
	dataCounter := map[metrics.Metric]metrics.Counter{}
	return &MyStorage{DataGauge: dataGauge, DataCounter: dataCounter, mu: sync.RWMutex{}}
}
