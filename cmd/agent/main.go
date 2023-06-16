package main

import (
	"log"
	"os"
	"time"

	"github.com/luckyseadog/go-dev/internal/agent"
)

func main() {
	address := os.Getenv("ADDRESS")
	if address == "" {
		address = "http://127.0.0.1:8080"
	} else {
		address = "http://" + address
	}

	contentType := "application/json"

	var pollInterval time.Duration
	pollIntervalStr := os.Getenv("POLL_INTERVAL")
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
