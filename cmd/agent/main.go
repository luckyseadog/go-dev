package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/luckyseadog/go-dev/internal/agent"
)

func main() {
	var addressFlag string
	var pollIntervalStrFlag string
	var reportIntervalStrFlag string

	flag.StringVar(&addressFlag, "a", "127.0.0.1:8080", "address of server")
	flag.StringVar(&pollIntervalStrFlag, "p", "2s", "time to catch metrics from program")
	flag.StringVar(&reportIntervalStrFlag, "r", "10s", "time to send metrics to server")
	flag.Parse()

	address := os.Getenv("ADDRESS")
	if address == "" {
		address = "http://" + addressFlag
	} else {
		address = "http://" + address
	}

	contentType := "application/json"

	var pollInterval time.Duration
	pollIntervalStr := os.Getenv("POLL_INTERVAL")
	if pollIntervalStr == "" {
		pollIntervalStr = pollIntervalStrFlag
	}
	if pollIntervalStr == "" {
		pollInterval = 2 * time.Second
	} else {
		var err error
		pollInterval, err = time.ParseDuration(pollIntervalStr)
		if err != nil {
			log.Fatal("Invalid pollInterval")
		}
	}

	var reportInterval time.Duration
	reportIntervalStr := os.Getenv("REPORT_INTERVAL")
	if reportIntervalStr == "" {
		reportIntervalStr = reportIntervalStrFlag
	}
	if reportIntervalStr == "" {
		reportInterval = 10 * time.Second
	} else {
		var err error
		reportInterval, err = time.ParseDuration(reportIntervalStr)
		if err != nil {
			log.Fatal("Invalid reportInterval")
		}
	}

	agent := agent.NewAgent(address, contentType, pollInterval, reportInterval)
	time.AfterFunc(2*time.Minute, func() {
		agent.Stop()
	})
	agent.Run()
}
