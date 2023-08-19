// Package storage provides functionalities for keeping data on server
// this package have two types of storage:
// - MyStorage that keeps data in map and periodically saves it to file
// - SQLStorage that keeps data as SQL database
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

// MyStorage holds Gauge metrics as DataGauge and Counter metrics as DataCounter
// and provides synchronization mechanisms for concurrent access.
//
// AutoSavingParams serves for sending save-to-file-signal in case of storeInterval = 0.
// If storeInterval != 0 then data is saved at intervals and MyStorage don't need to send signal
type MyStorage struct {
	DataGauge   map[metrics.Metric]metrics.Gauge
	DataCounter map[metrics.Metric]metrics.Counter
	mu          sync.RWMutex

	autoSavingParams AutoSavingParams
}

// NewStorage creates and initializes a new instance of MyStorage with the provided parameters.
// It returns a pointer to the initialized MyStorage.
//
// Parameters:
//   - storageChan: channel used to signal for writing to file.
//   - storeInterval: Interval after which data should be saved.
//     StoreInterval = 0 means saving at every update in MyStorage
//
// Returns:
//   - A pointer to a newly created and initialized MyStorage instance.
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

// StoreContext stores a metric value associated with the given metric key in the storage.
// It operates within the provided context, allowing for cancellation and timeout management.
//
// Parameters:
//   - ctx: The context in which the operation should be performed.
//   - metric: The metric key associated with the value to be stored.
//   - metricValue: The value to be stored for the specified metric key.
//
// Returns:
//   - An error if the storage operation fails or if the context is canceled.
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

// Store stores a metric value associated with the given metric key in the storage.
//
// Parameters:
//   - metric: The metric key associated with the value to be stored.
//   - metricValue: The value to be stored for the specified metric key.
//
// Returns:
//   - An error if the storage operation fails or if the provided metric value type is not expected.
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

// LoadContext retrieves the value of a specific metric associated with the provided metric type and key from the storage.
// It operates within the provided context, allowing for cancellation and timeout management.
//
// Parameters:
//   - ctx: The context in which the operation should be performed.
//   - metricType: The type of metric to load ("gauge" or "counter").
//   - metric: The metric key associated with the value to be retrieved.
//
// Returns:
//   - A Result containing the retrieved metric value and any associated error.
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

// Load retrieves the value of a specific metric associated with the provided metric type and key from the storage.
//
// Parameters:
//   - metricType: The type of metric to load ("gauge" or "counter").
//   - metric: The metric key associated with the value to be retrieved.
//
// Returns:
//   - A Result containing the retrieved metric value and any associated error.
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

// LoadDataGaugeContext retrieves a copy of the data stored in the gauge metrics of the storage.
// It operates within the provided context, allowing for cancellation and timeout management.
//
// Parameters:
//   - ctx: The context in which the operation should be performed.
//
// Returns:
//   - A Result containing the retrieved copy of gauge metric data and any associated error.
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

// LoadDataGauge retrieves a copy of the data stored in the gauge metrics of the storage.
//
// Returns:
//   - A Result containing the retrieved copy of gauge metric data and any associated error.
func (s *MyStorage) LoadDataGauge() Result {
	s.mu.RLock()
	defer s.mu.RUnlock()
	copyDataGauge := make(map[metrics.Metric]metrics.Gauge)

	for key, value := range s.DataGauge {
		copyDataGauge[key] = value
	}

	return Result{Value: copyDataGauge, Err: nil}
}

// LoadDataCounterContext retrieves a copy of the data stored in the counter metrics of the storage.
// It operates within the provided context, allowing for cancellation and timeout management.
//
// Parameters:
//   - ctx: The context in which the operation should be performed.
//
// Returns:
//   - A Result containing the retrieved copy of counter metric data and any associated error.
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

// LoadDataCounter retrieves a copy of the data stored in the counter metrics of the storage.
//
// Returns:
//   - A Result containing the retrieved copy of counter metric data and any associated error.
func (s *MyStorage) LoadDataCounter() Result {
	s.mu.RLock()
	defer s.mu.RUnlock()
	copyDataCounter := make(map[metrics.Metric]metrics.Counter)

	for key, value := range s.DataCounter {
		copyDataCounter[key] = value
	}

	return Result{Value: copyDataCounter, Err: nil}
}

// SaveToFile saves the gauge and counter metric data stored in the storage to the specified file.
// The data is serialized into JSON format and written to the file.
//
// Parameters:
//   - filepath: The path to the file where the data should be saved.
//
// Returns:
//   - An error if any issue occurs during serialization or file writing, or nil if successful.
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

// LoadFromFile loads gauge and counter metric data from the specified file and populates the storage.
// The data is deserialized from JSON format and stored in the respective gauge and counter maps.
//
// Parameters:
//   - filepath: The path to the file from which data should be loaded.
//
// Returns:
//   - An error if any issue occurs during file reading, deserialization, or data storing,
//     or nil if successful.
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
