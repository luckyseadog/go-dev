package agent

import (
	"github.com/luckyseadog/go-dev/internal/metrics"
	"net/http"
	"runtime"
	"sync"
	"time"
)

const UPDATE = "update"

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
