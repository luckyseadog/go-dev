package pkg

import (
	"errors"
	"github.com/luckyseadog/go-dev/internal"
	"sync"
)

var errNotExpectedType = errors.New("not expected type")

var StorageVar = NewStorage(100)

type StorageInterface interface {
	Store(metric internal.Metric, metricValue any) error
	Load(metric internal.Metric) (any, error)
}

type Storage struct {
	DataGauge   map[internal.Metric]internal.Gauge
	DataCounter map[internal.Metric][]internal.Counter
	Size        int
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
		if len(s.DataCounter[metric]) == s.Size {
			s.DataCounter[metric] = s.DataCounter[metric][1:]
		}
		s.DataCounter[metric] = append(s.DataCounter[metric], metricValue)
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

func NewStorage(size int) StorageInterface {
	dataGauge := map[internal.Metric]internal.Gauge{}
	dataCounter := map[internal.Metric][]internal.Counter{}
	return &Storage{DataGauge: dataGauge, DataCounter: dataCounter, Size: size, mu: sync.Mutex{}}
}
