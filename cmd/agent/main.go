// This file contains the main function for the agent application.
// The agent gathers and reports metrics to a remote server.
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
	// Command-line flag definitions.
	var addressFlag string           // addressFlag specifies the server address to connect to.
	var pollIntervalStrFlag string   // pollIntervalStrFlag specifies the polling interval for gathering metrics.
	var reportIntervalStrFlag string // reportIntervalStrFlag specifies the interval for sending metrics to the server.
	var secretKeyFlag string         // secretKeyFlag is the secret key used for digital signature of metrics.
	var rateLimitFlag string         // rateLimitFlag specifies the maximum number of concurrent requests that can be sent.
	var logging bool                 // logging indicates whether to save log to the file agent.log.

	// Parse command-line flags and set corresponding variables.
	flag.StringVar(&addressFlag, "a", "127.0.0.1:8080", "address of server")
	flag.StringVar(&pollIntervalStrFlag, "p", "2s", "time to catch metrics from program")
	flag.StringVar(&reportIntervalStrFlag, "r", "10s", "time to send metrics to server")
	flag.StringVar(&secretKeyFlag, "k", "", "secret key for digital signature")
	flag.StringVar(&rateLimitFlag, "l", "10", "how many concurrent requests could be sent")
	flag.BoolVar(&logging, "log", false, "whether to save log to file agent.log")

	// Parse the command-line flags.
	flag.Parse()

	// Initialize logging if the "logging" flag is set.
	if logging {
		// Open or create the "agent.log" file for writing logs.
		flog, err := os.OpenFile(`agent.log`, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
		// Create a new logger that writes to the log file with timestamp and file information.
		if err == nil {
			agent.MyLog = log.New(flog, `agent `, log.LstdFlags|log.Lshortfile)

			// Close the log file when the app is suspended.
			defer flog.Close()
		} else {
			// If there was an error creating the log file, log a fatal error.
			agent.MyLog.Fatal("error in creating file")
		}
	}

	// Retrieve configuration values from environment variables and command-line flags.
	// If corresponding environment variables are set, they take precedence over command-line flags.

	// Retrieve the server address from the environment variable "ADDRESS".
	// If not set, use the value provided by the command-line flag "-a".
	address := os.Getenv("ADDRESS")
	if address == "" {
		address = "http://" + addressFlag
	} else {
		address = "http://" + address
	}

	// Set the content type for the requests to the server.
	contentType := "application/json"

	// Retrieve and parse the polling interval for gathering metrics.
	// If "POLL_INTERVAL" environment variable is set, use its value.
	// Otherwise, use the value provided by the command-line flag "-r".
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

	// Retrieve and parse the report interval for sending metrics to the server.
	// If "REPORT_INTERVAL" environment variable is set, use its value.
	// Otherwise, use the value provided by the command-line flag "-p".
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

	// Retrieve the secret key for digital signature.
	// If "KEY" environment variable is set, use its value.
	// Otherwise, use the value provided by the command-line flag "-k".
	secretKeyStr := os.Getenv("KEY")
	if secretKeyStr == "" {
		secretKeyStr = secretKeyFlag
	}

	// Retrieve the number of how many concurrent requests could be sent.
	// If "RATE_LIMIT" environment variable is set, use its value.
	// Otherwise, use the value provided by the command-line flag "-l".
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

	// Create a new agent instance with the provided configuration parameters.
	// The agent gathers metrics from the specified address and reports them to the server.
	// It uses the specified content type for requests, pollInterval for metric collection,
	// reportInterval for sending metrics, secretKey for digital signature,
	// and rateLimit for controlling the number of concurrent requests.
	agent := agent.NewAgent(address, contentType, pollInterval, reportInterval, []byte(secretKeyStr), rateLimit)

	// Start the agent's operation. It begins collecting and reporting metrics based on the configured intervals.
	agent.Run()
}
