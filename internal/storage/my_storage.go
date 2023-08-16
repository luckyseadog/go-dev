package storage

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	"github.com/luckyseadog/go-dev/internal/metrics"
)

type MyStorage struct {
	DataGauge   map[metrics.Metric]metrics.Gauge
	DataCounter map[metrics.Metric]metrics.Counter
	mu          sync.RWMutex

	autoSavingParams AutoSavingParams
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

func (s *MyStorage) StoreContext(ctx context.Context, metric metrics.Metric, metricValue any) error {
	ch := make(chan error, 1)

	go func() {
		ch <- s.Store(metric, metricValue)
	}()

	select {
	case res := <-ch:
		return res
	case <-ctx.Done():
		return ctx.Err()
	}
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

func (s *MyStorage) LoadContext(ctx context.Context, metricType string, metric metrics.Metric) Result {
	ch := make(chan Result, 1)

	go func() {
		ch <- s.Load(metricType, metric)
	}()

	select {
	case res := <-ch:
		return res
	case <-ctx.Done():
		return Result{Value: nil, Err: ctx.Err()}
	}
}

func (s *MyStorage) Load(metricType string, metric metrics.Metric) Result {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if metricType == "gauge" {
		if valueGauge, ok := s.DataGauge[metric]; ok {
			return Result{Value: valueGauge, Err: nil}
		} else {
			return Result{Value: nil, Err: errNoSuchMetric}
		}
	} else if metricType == "counter" {
		if valueCounter, ok := s.DataCounter[metric]; ok {
			return Result{Value: valueCounter, Err: nil}
		} else {
			return Result{Value: nil, Err: errNoSuchMetric}
		}
	} else {
		return Result{Value: nil, Err: errNoSuchMetric}
	}
}

func (s *MyStorage) LoadDataGaugeContext(ctx context.Context) Result {
	ch := make(chan Result, 1)

	go func() {
		ch <- s.LoadDataGauge()
	}()

	select {
	case res := <-ch:
		return res
	case <-ctx.Done():
		return Result{Value: nil, Err: ctx.Err()}
	}
}

func (s *MyStorage) LoadDataGauge() Result {
	s.mu.RLock()
	defer s.mu.RUnlock()
	copyDataGauge := make(map[metrics.Metric]metrics.Gauge)

	for key, value := range s.DataGauge {
		copyDataGauge[key] = value
	}

	return Result{Value: copyDataGauge, Err: nil}
}

func (s *MyStorage) LoadDataCounterContext(ctx context.Context) Result {
	ch := make(chan Result, 1)

	go func() {
		ch <- s.LoadDataCounter()
	}()

	select {
	case res := <-ch:
		return res
	case <-ctx.Done():
		return Result{Value: nil, Err: ctx.Err()}
	}
}

func (s *MyStorage) LoadDataCounter() Result {
	s.mu.RLock()
	defer s.mu.RUnlock()
	copyDataCounter := make(map[metrics.Metric]metrics.Counter)

	for key, value := range s.DataCounter {
		copyDataCounter[key] = value
	}

	return Result{Value: copyDataCounter, Err: nil}
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
