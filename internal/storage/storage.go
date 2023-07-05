package storage

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/luckyseadog/go-dev/internal/metrics"
)

var errNotExpectedType = errors.New("not expected type")

type AutoSavingParams struct {
	storageChan   chan struct{}
	storeInterval time.Duration
}

type Storage interface {
	Store(metric metrics.Metric, metricValue any) error
	Load(metricType string, metric metrics.Metric) (any, error)
	LoadDataGauge() (map[metrics.Metric]metrics.Gauge, error)
	LoadDataCounter() (map[metrics.Metric]metrics.Counter, error)

	//SaveToDB(filepath string) error
	//LoadFromDB(filepath string) error
	//SaveToDB(db *sql.DB) error
	//LoadFromDB(db *sql.DB) error
}

type MyStorage struct {
	DataGauge   map[metrics.Metric]metrics.Gauge
	DataCounter map[metrics.Metric]metrics.Counter
	mu          sync.RWMutex

	autoSavingParams AutoSavingParams
}

func (s *MyStorage) Store(metric metrics.Metric, metricValue any) error {
	s.mu.Lock()
	defer func() {
		s.mu.Unlock()
		if s.autoSavingParams.storeInterval == 0 {
			s.autoSavingParams.storageChan <- struct{}{}
		}
	}()
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

func (s *MyStorage) LoadDataGauge() (map[metrics.Metric]metrics.Gauge, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	copyDataGauge := make(map[metrics.Metric]metrics.Gauge)

	for key, value := range s.DataGauge {
		copyDataGauge[key] = value
	}

	return copyDataGauge, nil
}

func (s *MyStorage) LoadDataCounter() (map[metrics.Metric]metrics.Counter, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	copyDataCounter := make(map[metrics.Metric]metrics.Counter)

	for key, value := range s.DataCounter {
		copyDataCounter[key] = value
	}

	return copyDataCounter, nil
}

func (s *MyStorage) SaveToFile(filepath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if filepath == "" {
		return nil
	}

	dataGauge := map[metrics.Metric]metrics.Gauge{}
	for key, value := range s.DataGauge {
		dataGauge[key] = value
	}
	dataCounter := map[metrics.Metric]metrics.Counter{}
	for key, value := range s.DataCounter {
		dataCounter[key] = value
	}

	fileData := metrics.FileData{DataGauge: dataGauge, DataCounter: dataCounter}

	data, err := json.Marshal(fileData)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath, data, 0777)
	if err != nil {
		return err
	}

	return nil
}

func (s *MyStorage) LoadFromFile(filepath string) error {
	s.mu.Lock()

	data, err := os.ReadFile(filepath)
	if err != nil {
		log.Fatal(err)
	}

	s.mu.Unlock()

	if !json.Valid(data) {
		return nil
	}

	var fileData metrics.FileData
	err = json.Unmarshal(data, &fileData)
	if err != nil {
		return err
	}

	for key, value := range fileData.DataGauge {
		err = s.Store(key, value)
		if err != nil {
			return err
		}
	}
	for key, value := range fileData.DataCounter {
		err = s.Store(key, value)
		if err != nil {
			return err
		}
	}

	return nil
}

func NewStorage(storageChan chan struct{}, storeInterval time.Duration) *MyStorage {
	dataGauge := map[metrics.Metric]metrics.Gauge{}
	dataCounter := map[metrics.Metric]metrics.Counter{}
	return &MyStorage{
		DataGauge:   dataGauge,
		DataCounter: dataCounter,
		mu:          sync.RWMutex{},
		autoSavingParams: AutoSavingParams{
			storageChan:   storageChan,
			storeInterval: storeInterval,
		},
	}
}
