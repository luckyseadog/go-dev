package storage

import (
	"encoding/json"
	"errors"
	"github.com/luckyseadog/go-dev/internal/metrics"
	"log"
	"os"
	"time"
)

func (s *MyStorage) SaveMetricsTypes(cancelChan chan struct{}, filepath string) {
	go func() {
		for {
			select {
			case <-cancelChan:
				return
			default:
				s.muMetric.Lock()
				data, err := json.Marshal(metrics.MapMetricTypes)
				if err != nil {
					log.Println(err)
				}
				s.muMetric.Unlock()
				err = os.WriteFile(filepath, data, 0777)
				if err != nil {
					log.Println(err)
				}
				time.Sleep(1 * time.Millisecond)
			}

		}
	}()
}

func (s *MyStorage) LoadMetricsTypes(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	if !json.Valid(data) {
		return errors.New("nat valid json: empty data")
	}

	tempMap := make(map[string]string)
	err = json.Unmarshal(data, &tempMap)
	if err != nil {
		return err
	}

	s.mu.Lock()
	for key, value := range tempMap {
		if _, ok := metrics.MapMetricTypes[key]; !ok {
			metrics.MapMetricTypes[key] = value
		}
	}
	s.mu.Unlock()

	return nil
}
