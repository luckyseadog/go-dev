package agent

import (
	"bytes"
	"context"
	"encoding/json"
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
		default:
			<-ticker.C
			a.mu.Lock()
			runtime.ReadMemStats(&a.metrics.MemStats)
			a.metrics.PollCount += 1
			a.metrics.RandomValue = metrics.Counter(rand.Intn(100))
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
		default:
			<-ticker.C
			a.mu.Lock()
			metricsGauge := metrics.GetMetrics(a.metrics.MemStats)
			a.mu.Unlock()
			metricsCounter := map[metrics.Metric]metrics.Counter{
				metrics.PollCount:   a.metrics.PollCount,
				metrics.RandomValue: a.metrics.RandomValue,
			}
			metricsCurrent := make([]metrics.Metrics, 0)

			for key, value := range metricsGauge {
				valueFloat64 := float64(value)
				metricsCurrent = append(metricsCurrent, metrics.Metrics{ID: string(key), MType: "gauge", Value: &valueFloat64})
			}
			for key, value := range metricsCounter {
				valueInt64 := int64(value)
				metricsCurrent = append(metricsCurrent, metrics.Metrics{ID: string(key), MType: "counter", Delta: &valueInt64})
			}

			data, err := json.Marshal(metricsCurrent)
			if err != nil {
				log.Println(err)
				return
			}

			address, err := url.Parse(a.ruler.address)
			if err != nil {
				log.Println(err)
				return
			}
			address.Path = address.Path + UPDATE

			ctx := context.Background()
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, address.String(), bytes.NewBuffer(data))
			if err != nil {
				log.Println(err)
				return
			}
			req.Header.Set("Content-Type", a.ruler.contentType)
			req.Header.Add("Accept", "application/json")
			response, err := a.client.Do(req)
			if err != nil {
				log.Println(err)
				return
			}
			defer response.Body.Close()
			_, err = io.Copy(io.Discard, response.Body)
			if err != nil {
				log.Println(err)
				return
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
