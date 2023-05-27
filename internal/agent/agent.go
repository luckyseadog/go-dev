package agent

import (
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"strings"
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
			_ = <-ticker.C
			a.mu.Lock()
			runtime.ReadMemStats(&a.metrics.memStats)
			a.metrics.pollCount += 1
			a.metrics.randomValue = rand.Intn(100)
			a.mu.Unlock()
		}
	}
}

func (a *Agent) PostStats() error {
	ticker := time.NewTicker(a.ruler.reportInterval)

	for {
		select {
		case <-a.cancel:
			return nil
		default:
			_ = <-ticker.C
			a.mu.Lock()
			stringsReader := strings.NewReader(getMetrics(a.metrics.memStats) + " " + strconv.Itoa(a.metrics.pollCount) +
				" " + strconv.Itoa(a.metrics.randomValue))
			req, err := http.NewRequest(http.MethodPost, a.ruler.address, stringsReader)
			a.mu.Unlock()
			if err != nil {
				return err
			}
			req.Header.Set("Content-Type", a.ruler.contentType)
			a.client.Do(req)
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

func getMetrics(memStats runtime.MemStats) string {
	metrics := []string{
		fmt.Sprintf("%v", memStats.Alloc),
		fmt.Sprintf("%v", memStats.BuckHashSys),
		fmt.Sprintf("%v", memStats.Frees),
		fmt.Sprintf("%v", memStats.GCCPUFraction),
		fmt.Sprintf("%v", memStats.GCSys),
		fmt.Sprintf("%v", memStats.HeapAlloc),
		fmt.Sprintf("%v", memStats.HeapIdle),
		fmt.Sprintf("%v", memStats.HeapInuse),
		fmt.Sprintf("%v", memStats.HeapObjects),
		fmt.Sprintf("%v", memStats.HeapReleased),
		fmt.Sprintf("%v", memStats.HeapSys),
		fmt.Sprintf("%v", memStats.LastGC),
		fmt.Sprintf("%v", memStats.Lookups),
		fmt.Sprintf("%v", memStats.MCacheInuse),
		fmt.Sprintf("%v", memStats.MCacheSys),
		fmt.Sprintf("%v", memStats.MSpanInuse),
		fmt.Sprintf("%v", memStats.MSpanSys),
		fmt.Sprintf("%v", memStats.Mallocs),
		fmt.Sprintf("%v", memStats.NextGC),
		fmt.Sprintf("%v", memStats.NumForcedGC),
		fmt.Sprintf("%v", memStats.NumGC),
		fmt.Sprintf("%v", memStats.OtherSys),
		fmt.Sprintf("%v", memStats.PauseTotalNs),
		fmt.Sprintf("%v", memStats.StackInuse),
		fmt.Sprintf("%v", memStats.StackSys),
		fmt.Sprintf("%v", memStats.Sys),
		fmt.Sprintf("%v", memStats.TotalAlloc),
	}

	return strings.Join(metrics, " ")
}
