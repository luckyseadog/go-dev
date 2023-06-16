package agent

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"path"
	"runtime"
	"sync"
	"time"

	"github.com/luckyseadog/go-dev/internal/metrics"
)

type InteractionRules struct {
	address        string
	contentType    string
	pollInterval   time.Duration
	reportInterval time.Duration
}

type Metrics struct {
	MemStats    runtime.MemStats
	PollCount   metrics.Counter
	RandomValue metrics.Counter
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
	cancel := make(chan struct{})
	return &Agent{client: http.Client{}, ruler: interactionRules, cancel: cancel}
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
			metricsInt := map[metrics.Metric]metrics.Counter{
				metrics.PollCount:   a.metrics.PollCount,
				metrics.RandomValue: a.metrics.RandomValue,
			}
			for key, value := range metricsGauge {
				address, err := url.Parse(a.ruler.address)
				if err != nil {
					log.Println(err)
					return
				}

				address.Path = path.Join(address.Path, "update", "gauge", string(key), fmt.Sprintf("%f", value))
				ctx := context.Background()
				req, err := http.NewRequestWithContext(ctx, http.MethodPost, address.String(), nil)
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

				address.Path = path.Join(address.Path, "update", "counter", string(key), fmt.Sprintf("%d", value))
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
	<-a.cancel
}

func (a *Agent) Stop() {
	close(a.cancel)
}
