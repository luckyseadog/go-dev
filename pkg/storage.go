package pkg

import (
	"errors"
	"github.com/luckyseadog/go-dev/internal"
)

var errNotExpectedType error = errors.New("not expected type")

var StorageVar = NewStorage(100)

type StorageInterface interface {
	Store(metric internal.Metric, metricValue any) error
	Load(metric internal.Metric) (any, error)
}

type Storage struct {
	DataGauge   map[internal.Metric]internal.Gauge
	DataCounter map[internal.Metric]*internal.Queue
}

func (s *Storage) Store(metric internal.Metric, metricValue any) error {
	switch metricValue.(type) {
	case internal.Gauge:
		s.DataGauge[metric] = metricValue.(internal.Gauge)
		return nil
	case internal.Counter:
		if s.DataCounter[metric].Size == s.DataCounter[metric].GetLength() {
			s.DataCounter[metric].Pop()
		}
		s.DataCounter[metric].Push(metricValue.(internal.Counter))
		return nil
	default:
		return errNotExpectedType
	}
}

func (s *Storage) Load(metric internal.Metric) (any, error) {
	valueGauge, ok := s.DataGauge[metric]
	if ok {
		return valueGauge, nil
	} else if valueCounter, ok := s.DataCounter[metric]; ok {
		return valueCounter, nil
	} else {
		return nil, errors.New("no such metric")
	}
}

func NewStorage(size internal.Counter) StorageInterface {
	dataGauge := map[internal.Metric]internal.Gauge{}
	dataCounter := map[internal.Metric]*internal.Queue{}
	dataCounter[internal.RandomValue] = internal.NewQueue(size)
	dataCounter[internal.PollCount] = internal.NewQueue(size)
	return &Storage{dataGauge, dataCounter}
}
