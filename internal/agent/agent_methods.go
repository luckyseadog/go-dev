package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/luckyseadog/go-dev/internal/metrics"
	"github.com/luckyseadog/go-dev/internal/security"
)

func (a *Agent) GetStats() {
	ticker := time.NewTicker(a.ruler.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-a.cancel:
			return
		case <-ticker.C:
			<-a.ruler.rateLimitChan
			a.mu.RLock()
			runtime.ReadMemStats(&a.metrics.MemStats)
			a.metrics.RandomValue = metrics.Gauge(rand.Float64())
			a.metrics.PollCount += 1
			a.mu.RUnlock()
			a.ruler.rateLimitChan <- struct{}{}
		}
	}
}

func (a *Agent) GetExtendedStats() {
	ticker := time.NewTicker(a.ruler.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-a.cancel:
			return
		case <-ticker.C:
			<-a.ruler.rateLimitChan
			a.mu.RLock()
			v, err := mem.VirtualMemory()
			if err != nil {
				MyLog.Println(err)
			}
			a.metrics.VirtualMemory = *v

			CPUUtilizationFloat, err := cpu.Percent(0, true)
			if err != nil {
				MyLog.Println(err)
			}
			CPUUtilization := make([]metrics.Gauge, 0, len(CPUUtilizationFloat))
			for _, el := range CPUUtilizationFloat {
				CPUUtilization = append(CPUUtilization, metrics.Gauge(el))
			}
			a.metrics.CPUUtilization = CPUUtilization
			a.mu.RUnlock()
			a.ruler.rateLimitChan <- struct{}{}
		}
	}
}

func (a *Agent) PostStats() {
	ticker := time.NewTicker(a.ruler.reportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-a.cancel:
			return
		case <-ticker.C:
			a.mu.Lock()
			metricsGauge := metrics.GetMetrics(a.metrics.MemStats)
			metricsGauge[metrics.RandomValue] = a.metrics.RandomValue
			metricsGauge[metrics.TotalMemory] = metrics.Gauge(a.metrics.VirtualMemory.Total)
			metricsGauge[metrics.FreeMemory] = metrics.Gauge(a.metrics.VirtualMemory.Available)
			for i, cpuMetric := range a.metrics.CPUUtilization {
				metricsGauge[metrics.Metric(fmt.Sprintf("CPUutilization%d", i+1))] = cpuMetric
			}
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
				MyLog.Println(err)
				continue
			}

			address, err := url.Parse(a.ruler.address)
			if err != nil {
				MyLog.Println(err)
			}
			address.Path = address.Path + UPDATE

			ctx := context.Background()
			fmt.Println(string(data))
			req, err := http.NewRequestWithContext(ctx, http.MethodPost, address.String(), bytes.NewBuffer(data))
			if err != nil {
				MyLog.Println(err)
				continue
			}
			req.Header.Set("Content-Type", a.ruler.contentType)
			req.Header.Add("Accept", "application/json")
			response, err := a.client.Do(req)
			if err != nil {
				MyLog.Println(err)
				continue
			}
			defer response.Body.Close()
			_, err = io.Copy(io.Discard, response.Body)
			if err != nil {
				MyLog.Println(err)
				continue
			}
		}
	}
}

func (a *Agent) Run() {
	go a.GetStats()
	go a.GetExtendedStats()
	go a.PostStats()
	<-a.cancel
}

func (a *Agent) Stop() {
	close(a.cancel)
}
