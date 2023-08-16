package main

import (
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/luckyseadog/go-dev/internal/agent"
)

func main() {
	var addressFlag string
	var pollIntervalStrFlag string
	var reportIntervalStrFlag string
	var secretKeyFlag string
	var rateLimitFlag string
	var logging bool

	flag.StringVar(&addressFlag, "a", "127.0.0.1:8080", "address of server")
	flag.StringVar(&pollIntervalStrFlag, "p", "2s", "time to catch metrics from program")
	flag.StringVar(&reportIntervalStrFlag, "r", "10s", "time to send metrics to server")
	flag.StringVar(&secretKeyFlag, "k", "", "secret key for digital signature")
	flag.StringVar(&rateLimitFlag, "l", "10", "how many concurrent requests could be send")
	flag.BoolVar(&logging, "log", false, "whether to save log to file agent.log")
	flag.Parse()

	if logging {
		flog, err := os.OpenFile(`agent.log`, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
		if err == nil {
			agent.MyLog = log.New(flog, `agent `, log.LstdFlags|log.Lshortfile)
			defer flog.Close()
		} else {
			agent.MyLog.Fatal("error in creating file")
		}
	}

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
			agent.MyLog.Fatal("Invalid pollInterval")
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
			agent.MyLog.Fatal("Invalid reportInterval")
		}
	}

	secretKeyStr := os.Getenv("KEY")
	if secretKeyStr == "" {
		secretKeyStr = secretKeyFlag
	}

	var rateLimit int
	rateLimitStr := os.Getenv("RATE_LIMIT")
	if rateLimitStr == "" {
		rateLimitStr = rateLimitFlag
	}
	if rateLimitStr == "" {
		rateLimit = 10
	} else {
		var err error
		rateLimit, err = strconv.Atoi(rateLimitStr)
		if err != nil {
			agent.MyLog.Fatal("Invalid rateLimit")
		}
	}

	agent := agent.NewAgent(address, contentType, pollInterval, reportInterval, []byte(secretKeyStr), rateLimit)
	time.AfterFunc(10*time.Minute, func() {
		agent.Stop()
	})
	agent.Run()
}
