package storage

import (
	"errors"
	"sync"

	"github.com/luckyseadog/go-dev/internal/metrics"
)

var errNotExpectedType = errors.New("not expected type")

var StorageVar = NewStorage()

type StorageInterface interface {
	Store(metric metrics.Metric, metricValue any) error
	Load(metric metrics.Metric) (any, error)
}

type Storage struct {
	DataGauge   map[metrics.Metric]metrics.Gauge
	DataCounter map[metrics.Metric]metrics.Counter
	mu          sync.Mutex
}

func (s *Storage) Store(metric metrics.Metric, metricValue any) error {
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

func (s *Storage) Load(metric metrics.Metric) (any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	valueGauge, ok := s.DataGauge[metric]
	if ok {
		return valueGauge, nil
	} else if valueCounter, ok := s.DataCounter[metric]; ok {
		return valueCounter, nil
	} else {
		return nil, errors.New("no such metric")
	}
}

func NewStorage() *Storage {
	dataGauge := map[metrics.Metric]metrics.Gauge{}
	dataCounter := map[metrics.Metric]metrics.Counter{}
	return &Storage{DataGauge: dataGauge, DataCounter: dataCounter, mu: sync.Mutex{}}
}
