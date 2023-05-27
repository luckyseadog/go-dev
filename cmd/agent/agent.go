package main

import (
	"fmt"
	"github.com/luckyseadog/go-dev/internal"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"path"
	"runtime"
	"sync"
	"time"
)

type InteractionRules struct {
	address        string
	contentType    string
	pollInterval   time.Duration
	reportInterval time.Duration
}

type Metrics struct {
	memStats    runtime.MemStats
	pollCount   int
	randomValue int
}

type Agent struct {
	client  http.Client
	metrics Metrics
	mu      sync.Mutex
	cancel  chan struct{}
	ruler   InteractionRules
}

func NewAgent(address string, contentType string, pollInterval time.Duration, reportInterval time.Duration) *Agent {
	interactionRules := InteractionRules{
		address:        address,
		contentType:    contentType,
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
	}
	return &Agent{client: http.Client{}, ruler: interactionRules}
}

func (a *Agent) GetStats() {
	ticker := time.NewTicker(a.ruler.pollInterval)
	for {
		select {
		case <-a.cancel:
			return
		default:
			<-ticker.C
			a.mu.Lock()
			runtime.ReadMemStats(&a.metrics.memStats)
			a.metrics.pollCount += 1
			a.metrics.randomValue = rand.Intn(100)
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
			metricsGauge := internal.GetMetrics(a.metrics.memStats)
			a.mu.Unlock()
			metricsInt := map[string]internal.Counter{
				"PollCount":   internal.Counter(a.metrics.pollCount),
				"RandomValue": internal.Counter(a.metrics.randomValue),
			}
			for key, value := range metricsGauge {
				address, err := url.Parse(a.ruler.address)
				if err != nil {
					log.Println(err)
					return
				}

				address.Path = path.Join(address.Path, "update", "Gauge", key, fmt.Sprintf("%f", value))
				req, err := http.NewRequest(http.MethodPost, address.String(), nil)
				if err != nil {
					log.Println(err)
					return
				}
				req.Header.Set("Content-Type", a.ruler.contentType)
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
			for key, value := range metricsInt {
				address, err := url.Parse(a.ruler.address)
				if err != nil {
					log.Println(err)
					return
				}

				address.Path = path.Join(address.Path, "update", "Counter", key, fmt.Sprintf("%d", value))
				req, err := http.NewRequest(http.MethodPost, address.String(), nil)
				if err != nil {
					log.Println(err)
					return
				}
				req.Header.Set("Content-Type", a.ruler.contentType)
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
}

func (a *Agent) Run() {
	go a.GetStats()
	go a.PostStats()
}

func (a *Agent) Stop() {
	close(a.cancel)
}

func main() {
	agent := NewAgent("http://127.0.0.1:8080", "text/plain", 2*time.Second, 10*time.Second)
	agent.Run()
	time.Sleep(5 * time.Minute)
}
