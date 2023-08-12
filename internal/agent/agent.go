package agent

import (
	"log"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/mem"

	"github.com/luckyseadog/go-dev/internal/metrics"
)

const UPDATE = "updates/"

var MyLog = log.Default()

type InteractionRules struct {
	address        string
	contentType    string
	pollInterval   time.Duration
	reportInterval time.Duration
	secretKey      []byte
	rateLimitChan  chan struct{}
}

type Metrics struct {
	MemStats       runtime.MemStats
	VirtualMemory  mem.VirtualMemoryStat
	CPUUtilization []metrics.Gauge
	PollCount      metrics.Counter
	RandomValue    metrics.Gauge
}

type Agent struct {
	client  http.Client
	metrics Metrics
	mu      sync.RWMutex
	cancel  chan struct{}
	ruler   InteractionRules
}

func NewAgent(address string, contentType string, pollInterval time.Duration, reportInterval time.Duration, secretKey []byte, rateLimit int) *Agent {
	rateLimitChan := make(chan struct{}, rateLimit)
	for i := 0; i < rateLimit; i++ {
		rateLimitChan <- struct{}{}
	}
	interactionRules := InteractionRules{
		address:        address,
		contentType:    contentType,
		pollInterval:   pollInterval,
		reportInterval: reportInterval,
		secretKey:      secretKey,
		rateLimitChan:  rateLimitChan,
	}
	cancel := make(chan struct{})
	return &Agent{client: http.Client{}, ruler: interactionRules, cancel: cancel}
}
