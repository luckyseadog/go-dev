package pkg

import (
	"errors"
	"github.com/luckyseadog/go-dev/internal"
	"sync"
)

var errNotExpectedType = errors.New("not expected type")

var StorageVar = NewStorage()

type StorageInterface interface {
	Store(metric internal.Metric, metricValue any) error
	Load(metric internal.Metric) (any, error)
}

type Storage struct {
	DataGauge   map[internal.Metric]internal.Gauge
	DataCounter map[internal.Metric]internal.Counter
	mu          sync.Mutex
}

func (s *Storage) Store(metric internal.Metric, metricValue any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	switch metricValue := metricValue.(type) {
	case internal.Gauge:
		s.DataGauge[metric] = metricValue
		return nil
	case internal.Counter:
		s.DataCounter[metric] += metricValue
		return nil
	default:
		return errNotExpectedType
	}
}

func (s *Storage) Load(metric internal.Metric) (any, error) {
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
	dataGauge := map[internal.Metric]internal.Gauge{}
	dataCounter := map[internal.Metric]internal.Counter{}
	return &Storage{DataGauge: dataGauge, DataCounter: dataCounter, mu: sync.Mutex{}}
}
