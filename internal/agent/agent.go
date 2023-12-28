// Package agent provides functionalities for creating an agent responsible for monitoring metrics
// and reporting updates to a server.
package agent

import (
	"log"
	"net/http"
	"runtime"
	"sync"
	"time"
	"crypto/tls"
	"path"

	"github.com/shirou/gopsutil/v3/mem"

	"github.com/luckyseadog/go-dev/internal/metrics"
)

// UPDATE defines the API endpoint for sending updates to the server.
const UPDATE = "updates/"

// MyLog is the logger used for agent logs. It is initialized with log.Default() by default.
var MyLog = log.Default()

// InteractionRules holds configuration parameters for the agent's behavior.
// It specifies the address of the server, content type for requests, poll and report intervals,
// secret key for digital signature, and a rate limit channel.
type InteractionRules struct {
	address        string
	contentType    string
	pollInterval   time.Duration
	reportInterval time.Duration
	secretKey      []byte
	rateLimitChan  chan struct{}
}

// Metrics struct holds various metric data collected by the agent.
type Metrics struct {
	MemStats       runtime.MemStats      // Memory statistics collected from the runtime.
	VirtualMemory  mem.VirtualMemoryStat // Virtual memory statistics.
	CPUUtilization []metrics.Gauge       // Slice of CPU utilization.
	PollCount      metrics.Counter       // Counter for tracking polling operations.
	RandomValue    metrics.Gauge         // Gauge for tracking a random value.
}

// Agent struct represents the monitoring agent responsible for collecting and reporting metrics.
type Agent struct {
	client  *http.Client      // HTTP client responsible for sending metric updates.
	metrics Metrics          // Metrics collected by the agent.
	mu      sync.RWMutex     // Mutex for safe concurrent access to metrics.
	cancel  chan struct{}    // Channel for signaling agent cancellation.
	ruler   InteractionRules // Configuration rules for agent behavior.
}

// NewAgent creates and initializes a new instance of the Agent with the provided parameters.
// It returns a pointer to the initialized Agent.
// The agent is responsible for monitoring metrics, sending updates to a server, and managing concurrency.
//
// Parameters:
//   - address: The address of the server to send metric updates.
//   - contentType: The content type for HTTP requests.
//   - pollInterval: The time interval for polling metrics from programs.
//   - reportInterval: The time interval for sending metrics to the server.
//   - secretKey: The secret key used for digital signature.
//   - rateLimit: The maximum number of concurrent requests the agent can handle.
//
// Returns:
//   - A pointer to a newly created and initialized Agent instance.
func NewAgent(address string, contentType string, pollInterval time.Duration, reportInterval time.Duration, secretKey []byte, rateLimit int, cryptoKeyDir string) *Agent {
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

	var client *http.Client
	if cryptoKeyDir != "" {
		clientTLSCert, err := tls.LoadX509KeyPair(path.Join(cryptoKeyDir, "certAgent.pem"), path.Join(cryptoKeyDir, "privateKeyAgent.pem"))
		if err != nil {
			log.Fatalf("Error loading certificate and key file: %v", err)
			return nil
		}

		tlsConfig := &tls.Config{
			ClientAuth: tls.NoClientCert,
			InsecureSkipVerify: true,
			Certificates: []tls.Certificate{clientTLSCert},
		}
		tr := &http.Transport{
			TLSClientConfig: tlsConfig,
		}
		client = &http.Client{Transport: tr}
	} else {
		client = &http.Client{}
	}
	return &Agent{client: client, ruler: interactionRules, cancel: cancel}
}
