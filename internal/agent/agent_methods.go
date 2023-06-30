package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/luckyseadog/go-dev/internal/security"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"runtime"
	"time"

	"github.com/luckyseadog/go-dev/internal/metrics"
)

func (a *Agent) GetStats() {
	ticker := time.NewTicker(a.ruler.pollInterval)
	for {
		select {
		case <-a.cancel:
			return
		case <-ticker.C:
			a.mu.Lock()
			runtime.ReadMemStats(&a.metrics.MemStats)
			a.metrics.RandomValue = metrics.Gauge(rand.Float64())
			a.metrics.PollCount += 1
			a.mu.Unlock()
		}
	}
}

func (a *Agent) PostStats() {
	ticker := time.NewTicker(a.ruler.reportInterval)

	for {
		select {
		case <-a.cancel:
			return
		case <-ticker.C:
			a.mu.Lock()
			metricsGauge := metrics.GetMetrics(a.metrics.MemStats)
			metricsGauge[metrics.RandomValue] = a.metrics.RandomValue
			metricsCounter := map[metrics.Metric]metrics.Counter{
				metrics.PollCount: a.metrics.PollCount,
			}
			a.metrics.PollCount = 0
			a.mu.Unlock()
			metricsCurrent := make([]metrics.Metrics, 0)

			for key, value := range metricsGauge {
				valueFloat64 := float64(value)

				var metricHash string
				if len(a.ruler.secretKey) > 0 {
					metricHash = security.Hash(fmt.Sprintf("%s:gauge:%f", string(key), valueFloat64), a.ruler.secretKey)
				}
				metricsCurrent = append(
					metricsCurrent,
					metrics.Metrics{ID: string(key), MType: "gauge", Value: &valueFloat64, Hash: metricHash},
				)
			}
			for key, value := range metricsCounter {
				valueInt64 := int64(value)

				var metricHash string
				if len(a.ruler.secretKey) > 0 {
					metricHash = security.Hash(fmt.Sprintf("%s:counter:%d", string(key), valueInt64), a.ruler.secretKey)
				}
				metricsCurrent = append(
					metricsCurrent,
					metrics.Metrics{ID: string(key), MType: "counter", Delta: &valueInt64, Hash: metricHash},
				)
			}

			data, err := json.Marshal(metricsCurrent)
			if err != nil {
				log.Println(err)
				continue
			}

			address, err := url.Parse(a.ruler.address)
			if err != nil {
				log.Println(err)
				continue
			}
			address.Path = address.Path + UPDATE

			ctx := context.Background()
			fmt.Println(string(data))
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, address.String(), bytes.NewBuffer(data))
			if err != nil {
				log.Println(err)
				continue
			}
			req.Header.Set("Content-Type", a.ruler.contentType)
			req.Header.Add("Accept", "application/json")
			response, err := a.client.Do(req)
			if err != nil {
				log.Println(err)
				continue
			}
			defer response.Body.Close()
			_, err = io.Copy(io.Discard, response.Body)
			if err != nil {
				log.Println(err)
				continue
			}
		}
	}
}

func (a *Agent) Run() {
	go a.GetStats()
	go a.PostStats()
	<-a.cancel
}

func (a *Agent) Stop() {
	close(a.cancel)
}
