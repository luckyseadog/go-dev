package metrics

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"time"
)

func SaveMetricsTypes(cancelChan chan struct{}, filepath string) {
	go func() {
		for {
			select {
			case <-cancelChan:
				return
			default:
				data, err := json.Marshal(MapMetricTypes)
				if err != nil {
					log.Println(err)
				}
				err = os.WriteFile(filepath, data, 0777)
				if err != nil {
					log.Println(err)
				}
				time.Sleep(1 * time.Second)
			}

		}
	}()
}

func LoadMetricsTypes(filepath string) error {
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

	for key, value := range tempMap {
		if _, ok := MapMetricTypes[key]; !ok {
			MapMetricTypes[key] = value
		}
	}

	return nil
}
